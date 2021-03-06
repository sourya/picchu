package revision

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	picchuv1alpha1 "go.medium.engineering/picchu/pkg/apis/picchu/v1alpha1"
	"go.medium.engineering/picchu/pkg/controller/utils"
	promapi "go.medium.engineering/picchu/pkg/prometheus"
	sentry "go.medium.engineering/picchu/pkg/sentry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO(bob): Add these to Revision type
const AcceptancePercentage uint32 = 50

var (
	clog              = logf.Log.WithName("controller_revision")
	AcceptanceTargets = []string{"production"}

	revisionFailedGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "picchu_revision_failed",
		Help: "track failed revisions",
	}, []string{"app", "tag"})
	mirrorFailureCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "picchu_mirror_failure_counter",
		Help: "Record picchu mirror failures",
	}, []string{"app", "mirror"})
)

// Add creates a new Revision Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, c utils.Config) error {
	metrics.Registry.MustRegister(revisionFailedGauge)
	metrics.Registry.MustRegister(mirrorFailureCounter)
	return add(mgr, newReconciler(mgr, c))
}

type PromAPI interface {
	IsRevisionTriggered(ctx context.Context, name, tag string, withCanary bool) (bool, error)
}

type NoopPromAPI struct{}

func (n *NoopPromAPI) IsRevisionTriggered(ctx context.Context, name, tag string, withCanary bool) (bool, error) {
	return false, nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, c utils.Config) reconcile.Reconciler {
	var err error
	var api PromAPI
	if c.PrometheusQueryAddress != "" {
		api, err = promapi.NewAPI(c.PrometheusQueryAddress, c.PrometheusQueryTTL)
	} else {
		api = &NoopPromAPI{}
	}
	if err != nil {
		panic(err)
	}

	var sentryClient *sentry.Client
	if c.SentryAuthToken != "" {
		sentryClient, err = sentry.NewClient(c.SentryAuthToken, nil, nil)
		if err != nil {
			panic(err)
		}
	}

	return &ReconcileRevision{
		client:       mgr.GetClient(),
		scheme:       mgr.GetScheme(),
		config:       c,
		promAPI:      api,
		sentryClient: sentryClient,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	_, err := builder.SimpleController().
		WithManager(mgr).
		ForType(&picchuv1alpha1.Revision{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(_ event.UpdateEvent) bool { return false },
		}).
		Build(r)

	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRevision{}

// ReconcileRevision reconciles a Revision object
type ReconcileRevision struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	scheme       *runtime.Scheme
	config       utils.Config
	promAPI      PromAPI
	sentryClient *sentry.Client
}

// Reconcile reads that state of the cluster for a Revision object and makes changes based on the state read
// and what is in the Revision.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRevision) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := clog.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Revision")

	// Fetch the Revision instance
	ctx := context.TODO()
	instance := &picchuv1alpha1.Revision{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	r.scheme.Default(instance)
	log := reqLogger.WithValues("App", instance.Spec.App.Name, "Tag", instance.Spec.App.Tag)

	mirrors := &picchuv1alpha1.MirrorList{}
	err = r.client.List(ctx, nil, mirrors)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err = r.LabelWithAppAndFleets(log, instance); err != nil {
		return reconcile.Result{}, err
	}

	promLabels := prometheus.Labels{
		"app": instance.Spec.App.Name,
		"tag": instance.Spec.App.Tag,
	}

	status, err := r.syncReleaseManager(log, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range mirrors.Items {
		mirror := mirrors.Items[i]
		err = r.mirrorRevision(ctx, log, &mirror, instance)
		if err != nil {
			log.Error(err, "Failed to mirror revision", "Mirror", mirror.Spec.ClusterName)
			mLabels := prometheus.Labels{
				"app":    instance.Spec.App.Name,
				"mirror": mirror.Spec.ClusterName,
			}
			mirrorFailureCounter.With(mLabels).Inc()
		}
	}

	triggered, err := r.promAPI.IsRevisionTriggered(context.TODO(), instance.Spec.App.Name, instance.Spec.App.Tag, instance.Spec.CanaryWithSLIRules)
	if err != nil {
		return reconcile.Result{}, err
	}
	if triggered && !instance.Spec.IgnoreSLOs {
		accepted := true
		var acceptanceTargets []string
		for _, target := range instance.Spec.Targets {
			if target.AcceptanceTarget {
				acceptanceTargets = append(acceptanceTargets, target.Name)
			} else {
				for _, targetName := range AcceptanceTargets {
					if target.Name == targetName {
						acceptanceTargets = append(acceptanceTargets, target.Name)
					}
				}
			}
		}

		for _, targetStatus := range status.Targets {
			for _, targetName := range acceptanceTargets {
				if targetStatus.Name == targetName {
					if targetStatus.Release.PeakPercent < AcceptancePercentage {
						accepted = false
					}
				}
			}
		}
		if !accepted {
			op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, instance, func(runtime.Object) error {
				instance.Fail()
				return nil
			})
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info("Set Revision State to failed", "Op", op)
			revisionFailedGauge.With(promLabels).Set(float64(1))
		} else {
			revisionFailedGauge.With(promLabels).Set(float64(0))
		}
	} else {
		revisionFailedGauge.With(promLabels).Set(float64(0))
	}

	if r.config.SentryAuthToken != "" && r.config.SentryOrg != "" && instance.Spec.Sentry.Release && !status.Sentry.Release {
		s, err := r.createSentryReleaseForRevision(log, instance, r.config)
		if err != nil {
			return reconcile.Result{}, err
		}
		if s.DateCreated != nil {
			status.Sentry.Release = true
		}
	}

	instance.Status = status
	if err = r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.config.RequeueAfter}, nil
}

func (r *ReconcileRevision) LabelWithAppAndFleets(log logr.Logger, revision *picchuv1alpha1.Revision) error {
	fleetLabels := []string{}
	updated := false
	for _, target := range revision.Spec.Targets {
		name := fmt.Sprintf("%s%s", picchuv1alpha1.LabelFleetPrefix, target.Fleet)
		fleetLabels = append(fleetLabels, name)
		if _, ok := revision.Labels[name]; !ok {
			revision.Labels[name] = ""
			updated = true
		}
	}
	for name, _ := range revision.Labels {
		if strings.HasPrefix(name, picchuv1alpha1.LabelFleetPrefix) {
			found := false
			for _, expected := range fleetLabels {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				delete(revision.Labels, name)
				updated = true
			}
		}
	}

	if _, ok := revision.Labels[picchuv1alpha1.LabelApp]; !ok {
		revision.Labels[picchuv1alpha1.LabelApp] = revision.Spec.App.Name
		updated = true
	}

	if updated {
		return r.client.Update(context.TODO(), revision)
	}
	return nil
}

func (r *ReconcileRevision) getOrCreateReleaseManager(
	log logr.Logger,
	target *picchuv1alpha1.RevisionTarget,
	fleet string,
	revision *picchuv1alpha1.Revision,
) (*picchuv1alpha1.ReleaseManager, error) {
	lbls := map[string]string{
		picchuv1alpha1.LabelTarget: target.Name,
		picchuv1alpha1.LabelFleet:  target.Fleet,
		picchuv1alpha1.LabelApp:    revision.Spec.App.Name,
	}
	rms := &picchuv1alpha1.ReleaseManagerList{}
	opts := client.
		MatchingLabels(lbls).
		InNamespace(revision.Namespace)
	r.client.List(context.TODO(), opts, rms)
	if len(rms.Items) > 1 {
		panic(fmt.Sprintf("Too many ReleaseManagers matching %#v", lbls))
	}
	if len(rms.Items) == 1 {
		return &rms.Items[0], nil
	}
	rm := &picchuv1alpha1.ReleaseManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", revision.Spec.App.Name, target.Name),
			Namespace: revision.Namespace,
			Labels:    lbls,
			Finalizers: []string{
				picchuv1alpha1.FinalizerReleaseManager,
			},
		},
		Spec: picchuv1alpha1.ReleaseManagerSpec{
			Fleet:  target.Fleet,
			App:    revision.Spec.App.Name,
			Target: target.Name,
		},
	}
	if err := r.client.Create(context.TODO(), rm); err != nil {
		log.Error(err, "Failed to sync releaseManager")
		return nil, err
	}
	log.Info("ReleaseManager sync'd", "Type", "ReleaseManager", "Op", "created", "Content", rm, "Audit", true)

	return rm, nil
}

func (r *ReconcileRevision) syncReleaseManager(log logr.Logger, revision *picchuv1alpha1.Revision) (picchuv1alpha1.RevisionStatus, error) {
	// Sync releasemanagers
	rstatus := picchuv1alpha1.RevisionStatus{}
	rstatus.Sentry = revision.Status.Sentry
	rmCount, retiredCount := 0, 0
	for _, target := range revision.Spec.Targets {
		status := picchuv1alpha1.RevisionTargetStatus{Name: target.Name}
		rm, err := r.getOrCreateReleaseManager(log, &target, target.Fleet, revision)
		if err != nil {
			return rstatus, err
		}
		status.AddReleaseManagerStatus(*rm.RevisionStatus(revision.Spec.App.Tag))

		log.Info("Checking for garbage collection")
		rmCount++
		for _, rl := range rm.Status.Revisions {
			if rl.GitTimestamp == nil {
				log.Info("Not git timestamp found")
				continue
			}
			expiration := rl.GitTimestamp.Add(time.Duration(rl.TTL) * time.Second)
			if rl.Tag == revision.Spec.App.Tag {
				if time.Now().After(expiration) {
					switch rl.State.Current {
					case "pendingrelease", "releasing", "released":
						log.Info("Not removing for state", "State", rl.State.Current)
						// don't delete released revisions
					default:
						log.Info("Marking for deletion for state", "State", rl.State.Current)
						retiredCount++
					}
				} else {
					log.Info("Expiration not reached", "Expiration", expiration, "GitTimestamp", rl.GitTimestamp, "Ttl", rl.TTL)
				}
			} else {
				log.Info("Tag doesn't match", "StateTag", rl.Tag, "Tag", revision.Spec.App.Tag)
			}
		}
		rstatus.AddTarget(status)
	}

	// if Revision is expired in all ReleaseManagers, without deleting
	// clusterless revisions
	if retiredCount == rmCount && retiredCount > 0 {
		if err := r.deleteRevision(log, revision); err != nil {
			log.Error(err, "Failed to delete Revision")
			return rstatus, err
		}
	}
	return rstatus, nil
}

func (r *ReconcileRevision) deleteRevision(log logr.Logger, revision *picchuv1alpha1.Revision) error {
	log.Info("Deleting revision", "Name", revision.Name, "Namespace", revision.Namespace)
	if err := r.client.Delete(context.TODO(), revision); err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// createSentryReleaseForRevision performs the Sentry API calls to register a Revision with Sentry.
// It creates Project if missing (based on app name), a Release for the Project (based on tag and commit sha),
// and a Deployment for the Release (based on tag and target).
func (r *ReconcileRevision) createSentryReleaseForRevision(log logr.Logger, revision *picchuv1alpha1.Revision, config utils.Config) (sentry.Release, error) {
	tag, foundtag := revision.Labels[picchuv1alpha1.LabelTag]
	commit, foundref := revision.Labels[picchuv1alpha1.LabelCommit]
	repo, foundrepo := revision.Annotations[picchuv1alpha1.AnnotationRepo]
	app, foundapp := revision.Labels[picchuv1alpha1.LabelApp]

	if r.sentryClient != nil && foundtag && foundref && foundrepo && foundapp {
		log.Info("Registering release with Sentry", "Name", revision.Name, "Namespace", revision.Namespace, "Version", tag, "Commit", commit)

		if _, err := r.sentryClient.GetProject(config.SentryOrg, app); err != nil {
			log.Info("Could not get project, trying to create it", "Project", app)
			if _, err := r.sentryClient.CreateProject(config.SentryOrg, app); err != nil {
				return sentry.Release{}, err
			}
		}
		ref := &sentry.Ref{
			Repository: repo,
			Commit:     commit,
		}
		rel := &sentry.NewRelease{
			Version: tag,
			Ref:     commit,
			Projects: []string{
				app,
			},
			Refs: []sentry.Ref{
				*ref,
			},
		}
		newrel, err := r.sentryClient.CreateRelease(config.SentryOrg, *rel)
		if err != nil {
			return sentry.Release{}, err
		}

		for _, target := range revision.Spec.Targets {
			deploy := &sentry.NewDeploy{
				Version:     tag,
				Environment: target.Name,
			}
			err := r.sentryClient.CreateDeploy(config.SentryOrg, *deploy)
			if err != nil {
				return sentry.Release{}, err
			}
		}

		return newrel, err
	}

	return sentry.Release{}, nil
}

func (r *ReconcileRevision) mirrorRevision(
	ctx context.Context,
	log logr.Logger,
	mirror *picchuv1alpha1.Mirror,
	revision *picchuv1alpha1.Revision,
) error {
	log.Info("Mirroring revision", "Mirror", mirror.Spec.ClusterName)
	cluster := &picchuv1alpha1.Cluster{}
	key := types.NamespacedName{revision.Namespace, mirror.Spec.ClusterName}
	if err := r.client.Get(ctx, key, cluster); err != nil {
		return err
	}
	remoteClient, err := utils.RemoteClient(ctx, log, r.client, cluster)
	if err != nil {
		return err
	}
	for i := range revision.Spec.Targets {
		target := revision.Spec.Targets[i]
		selector, err := metav1.LabelSelectorAsSelector(target.ConfigSelector)
		if err != nil {
			return err
		}
		opts := &client.ListOptions{
			LabelSelector: selector,
			Namespace:     revision.Namespace,
		}
		configMapList := &corev1.ConfigMapList{}
		if err := r.client.List(ctx, opts, configMapList); err != nil {
			return err
		}
		if err := r.copyConfigMapList(ctx, log, remoteClient, configMapList); err != nil {
			return err
		}
		secretList := &corev1.SecretList{}
		if err := r.client.List(ctx, opts, secretList); err != nil {
			return err
		}
		if err := r.copySecretList(ctx, log, remoteClient, secretList); err != nil {
			return err
		}
	}

	// TODO(bob): this is bad because it makes picchu aware of kbfd and should be generalized, probably in the Mirror spec.
	opts := &client.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			"config.kbfd.medium.build/type": "inputs",
			"medium.build/app":              revision.Spec.App.Name,
			"medium.build/tag":              revision.Spec.App.Tag,
		}).AsSelector(),
		Namespace: "build",
	}
	configMapList := &corev1.ConfigMapList{}
	if err := r.client.List(ctx, opts, configMapList); err != nil {
		return err
	}
	if err := r.copyConfigMapList(ctx, log, remoteClient, configMapList); err != nil {
		return err
	}
	// end badness

	revCopy := &picchuv1alpha1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: revision.Annotations,
			Name:        revision.Name,
			Namespace:   revision.Namespace,
			Labels:      revision.Labels,
		},
		Spec: revision.DeepCopy().Spec,
	}
	_, err = controllerutil.CreateOrUpdate(ctx, remoteClient, revCopy, func(runtime.Object) error {
		revCopy.Spec = revision.Spec
		return nil
	})
	return err
}

func (r *ReconcileRevision) copyConfigMapList(
	ctx context.Context,
	log logr.Logger,
	remoteClient client.Client,
	configMapList *corev1.ConfigMapList,
) error {
	for i := range configMapList.Items {
		orig := configMapList.Items[i]
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: orig.Annotations,
				Name:        orig.Name,
				Namespace:   orig.Namespace,
				Labels:      orig.Labels,
			},
			Data: orig.Data,
		}
		_, err := controllerutil.CreateOrUpdate(ctx, remoteClient, configMap, func(runtime.Object) error {
			configMap.Data = configMapList.Items[i].Data
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileRevision) copySecretList(
	ctx context.Context,
	log logr.Logger,
	remoteClient client.Client,
	secretList *corev1.SecretList,
) error {
	for i := range secretList.Items {
		orig := secretList.Items[i]
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: orig.Annotations,
				Name:        orig.Name,
				Namespace:   orig.Namespace,
				Labels:      orig.Labels,
			},
			Type: orig.Type,
			Data: orig.Data,
		}
		_, err := controllerutil.CreateOrUpdate(ctx, remoteClient, secret, func(runtime.Object) error {
			secret.Data = orig.Data
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
