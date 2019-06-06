// Copyright © 2019 A Medium Corporation.
// Licensed under the Apache License, Version 2.0; see the NOTICE file.

// Code generated by client. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "go.medium.engineering/picchu/pkg/apis/picchu/v1alpha1"
	scheme "go.medium.engineering/picchu/pkg/client/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ClusterSecretsesGetter has a method to return a ClusterSecretsInterface.
// A group's client should implement this interface.
type ClusterSecretsesGetter interface {
	ClusterSecretses(namespace string) ClusterSecretsInterface
}

// ClusterSecretsInterface has methods to work with ClusterSecrets resources.
type ClusterSecretsInterface interface {
	Create(*v1alpha1.ClusterSecrets) (*v1alpha1.ClusterSecrets, error)
	Update(*v1alpha1.ClusterSecrets) (*v1alpha1.ClusterSecrets, error)
	UpdateStatus(*v1alpha1.ClusterSecrets) (*v1alpha1.ClusterSecrets, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ClusterSecrets, error)
	List(opts v1.ListOptions) (*v1alpha1.ClusterSecretsList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterSecrets, err error)
	ClusterSecretsExpansion
}

// clusterSecretses implements ClusterSecretsInterface
type clusterSecretses struct {
	client rest.Interface
	ns     string
}

// newClusterSecretses returns a ClusterSecretses
func newClusterSecretses(c *PicchuV1alpha1Client, namespace string) *clusterSecretses {
	return &clusterSecretses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the clusterSecrets, and returns the corresponding clusterSecrets object, and an error if there is any.
func (c *clusterSecretses) Get(name string, options v1.GetOptions) (result *v1alpha1.ClusterSecrets, err error) {
	result = &v1alpha1.ClusterSecrets{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clustersecretses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterSecretses that match those selectors.
func (c *clusterSecretses) List(opts v1.ListOptions) (result *v1alpha1.ClusterSecretsList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ClusterSecretsList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clustersecretses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterSecretses.
func (c *clusterSecretses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("clustersecretses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a clusterSecrets and creates it.  Returns the server's representation of the clusterSecrets, and an error, if there is any.
func (c *clusterSecretses) Create(clusterSecrets *v1alpha1.ClusterSecrets) (result *v1alpha1.ClusterSecrets, err error) {
	result = &v1alpha1.ClusterSecrets{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("clustersecretses").
		Body(clusterSecrets).
		Do().
		Into(result)
	return
}

// Update takes the representation of a clusterSecrets and updates it. Returns the server's representation of the clusterSecrets, and an error, if there is any.
func (c *clusterSecretses) Update(clusterSecrets *v1alpha1.ClusterSecrets) (result *v1alpha1.ClusterSecrets, err error) {
	result = &v1alpha1.ClusterSecrets{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("clustersecretses").
		Name(clusterSecrets.Name).
		Body(clusterSecrets).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *clusterSecretses) UpdateStatus(clusterSecrets *v1alpha1.ClusterSecrets) (result *v1alpha1.ClusterSecrets, err error) {
	result = &v1alpha1.ClusterSecrets{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("clustersecretses").
		Name(clusterSecrets.Name).
		SubResource("status").
		Body(clusterSecrets).
		Do().
		Into(result)
	return
}

// Delete takes name of the clusterSecrets and deletes it. Returns an error if one occurs.
func (c *clusterSecretses) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("clustersecretses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterSecretses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("clustersecretses").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched clusterSecrets.
func (c *clusterSecretses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterSecrets, err error) {
	result = &v1alpha1.ClusterSecrets{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("clustersecretses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}