package releasemanager

import (
	"time"

	"go.medium.engineering/picchu/pkg/controller/releasemanager/scaling"
)

const reconcileRatio = 0.9

type ScalableTargetAdapter struct {
	Incarnation
}

// IsReconciled returns true if the target is considered ready to be scaled to the next increment.
func (s *ScalableTargetAdapter) IsReconciled() bool {
	target := s.Incarnation.target()
	status := s.Incarnation.status

	if target == nil {
		return false
	}

	desired := s.Incarnation.divideReplicas(*target.Scale.Min)
	ratio := float64(status.Scale.Current) / float64(desired)
	return ratio > reconcileRatio
}

func (s *ScalableTargetAdapter) CurrentPercent() uint32 {
	return s.Incarnation.status.CurrentPercent
}

func (s *ScalableTargetAdapter) PeakPercent() uint32 {
	return s.Incarnation.status.PeakPercent
}

func (s *ScalableTargetAdapter) Delay() time.Duration {
	return time.Duration(*s.Incarnation.target().Release.Rate.DelaySeconds) * time.Second
}

func (s *ScalableTargetAdapter) Increment() uint32 {
	return s.Incarnation.target().Release.Rate.Increment
}

func (s *ScalableTargetAdapter) Max() uint32 {
	return s.Incarnation.target().Release.Max
}

func (s *ScalableTargetAdapter) LastUpdated() time.Time {
	lu := s.Incarnation.status.LastUpdated
	if lu != nil {
		return lu.Time
	}
	return time.Time{}
}

func LinearScale(i Incarnation, max uint32, t time.Time) uint32 {
	sta := ScalableTargetAdapter{i}
	return scaling.LinearScale(&sta, max, t)
}
