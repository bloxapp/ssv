// Code generated by MockGen. DO NOT EDIT.
// Source: ./validators_map.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	types "github.com/ssvlabs/ssv-spec/types"
	queue "github.com/ssvlabs/ssv/protocol/v2/ssv/queue"
	types0 "github.com/ssvlabs/ssv/protocol/v2/types"
	zap "go.uber.org/zap"
)

// MockValidator is a mock of Validator interface.
type MockValidator struct {
	ctrl     *gomock.Controller
	recorder *MockValidatorMockRecorder
}

// MockValidatorMockRecorder is the mock recorder for MockValidator.
type MockValidatorMockRecorder struct {
	mock *MockValidator
}

// NewMockValidator creates a new mock instance.
func NewMockValidator(ctrl *gomock.Controller) *MockValidator {
	mock := &MockValidator{ctrl: ctrl}
	mock.recorder = &MockValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockValidator) EXPECT() *MockValidatorMockRecorder {
	return m.recorder
}

// GetShare mocks base method.
func (m *MockValidator) GetShare() *types0.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShare")
	ret0, _ := ret[0].(*types0.SSVShare)
	return ret0
}

// GetShare indicates an expected call of GetShare.
func (mr *MockValidatorMockRecorder) GetShare() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShare", reflect.TypeOf((*MockValidator)(nil).GetShare))
}

// ProcessMessage mocks base method.
func (m *MockValidator) ProcessMessage(logger *zap.Logger, msg *queue.DecodedSSVMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessMessage", logger, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessMessage indicates an expected call of ProcessMessage.
func (mr *MockValidatorMockRecorder) ProcessMessage(logger, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessMessage", reflect.TypeOf((*MockValidator)(nil).ProcessMessage), logger, msg)
}

// StartDuty mocks base method.
func (m *MockValidator) StartDuty(logger *zap.Logger, duty *types.Duty) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartDuty", logger, duty)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartDuty indicates an expected call of StartDuty.
func (mr *MockValidatorMockRecorder) StartDuty(logger, duty interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartDuty", reflect.TypeOf((*MockValidator)(nil).StartDuty), logger, duty)
}
