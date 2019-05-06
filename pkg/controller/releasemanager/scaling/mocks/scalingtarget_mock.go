// Code generated by MockGen. DO NOT EDIT.
// Source: target.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockScalableTarget is a mock of ScalableTarget interface
type MockScalableTarget struct {
	ctrl     *gomock.Controller
	recorder *MockScalableTargetMockRecorder
}

// MockScalableTargetMockRecorder is the mock recorder for MockScalableTarget
type MockScalableTargetMockRecorder struct {
	mock *MockScalableTarget
}

// NewMockScalableTarget creates a new mock instance
func NewMockScalableTarget(ctrl *gomock.Controller) *MockScalableTarget {
	mock := &MockScalableTarget{ctrl: ctrl}
	mock.recorder = &MockScalableTargetMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockScalableTarget) EXPECT() *MockScalableTargetMockRecorder {
	return m.recorder
}

// CurrentPercent mocks base method
func (m *MockScalableTarget) CurrentPercent() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentPercent")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// CurrentPercent indicates an expected call of CurrentPercent
func (mr *MockScalableTargetMockRecorder) CurrentPercent() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentPercent", reflect.TypeOf((*MockScalableTarget)(nil).CurrentPercent))
}

// PeakPercent mocks base method
func (m *MockScalableTarget) PeakPercent() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PeakPercent")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// PeakPercent indicates an expected call of PeakPercent
func (mr *MockScalableTargetMockRecorder) PeakPercent() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeakPercent", reflect.TypeOf((*MockScalableTarget)(nil).PeakPercent))
}

// Delay mocks base method
func (m *MockScalableTarget) Delay() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delay")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// Delay indicates an expected call of Delay
func (mr *MockScalableTargetMockRecorder) Delay() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delay", reflect.TypeOf((*MockScalableTarget)(nil).Delay))
}

// Increment mocks base method
func (m *MockScalableTarget) Increment() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Increment")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// Increment indicates an expected call of Increment
func (mr *MockScalableTargetMockRecorder) Increment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Increment", reflect.TypeOf((*MockScalableTarget)(nil).Increment))
}

// Max mocks base method
func (m *MockScalableTarget) Max() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Max")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// Max indicates an expected call of Max
func (mr *MockScalableTargetMockRecorder) Max() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Max", reflect.TypeOf((*MockScalableTarget)(nil).Max))
}

// LastUpdated mocks base method
func (m *MockScalableTarget) LastUpdated() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastUpdated")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// LastUpdated indicates an expected call of LastUpdated
func (mr *MockScalableTargetMockRecorder) LastUpdated() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastUpdated", reflect.TypeOf((*MockScalableTarget)(nil).LastUpdated))
}