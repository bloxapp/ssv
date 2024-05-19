// Code generated by MockGen. DO NOT EDIT.
// Source: ./base_handler.go

// Package duties is a generated GoMock package.
package duties

import (
	context "context"
	reflect "reflect"

	networkconfig "github.com/ssvlabs/ssv/networkconfig"
	slotticker "github.com/ssvlabs/ssv/operator/slotticker"
	gomock "github.com/golang/mock/gomock"
	zap "go.uber.org/zap"
)

// MockdutyHandler is a mock of dutyHandler interface.
type MockdutyHandler struct {
	ctrl     *gomock.Controller
	recorder *MockdutyHandlerMockRecorder
}

// MockdutyHandlerMockRecorder is the mock recorder for MockdutyHandler.
type MockdutyHandlerMockRecorder struct {
	mock *MockdutyHandler
}

// NewMockdutyHandler creates a new mock instance.
func NewMockdutyHandler(ctrl *gomock.Controller) *MockdutyHandler {
	mock := &MockdutyHandler{ctrl: ctrl}
	mock.recorder = &MockdutyHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockdutyHandler) EXPECT() *MockdutyHandlerMockRecorder {
	return m.recorder
}

// HandleDuties mocks base method.
func (m *MockdutyHandler) HandleDuties(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleDuties", arg0)
}

// HandleDuties indicates an expected call of HandleDuties.
func (mr *MockdutyHandlerMockRecorder) HandleDuties(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleDuties", reflect.TypeOf((*MockdutyHandler)(nil).HandleDuties), arg0)
}

// HandleInitialDuties mocks base method.
func (m *MockdutyHandler) HandleInitialDuties(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleInitialDuties", arg0)
}

// HandleInitialDuties indicates an expected call of HandleInitialDuties.
func (mr *MockdutyHandlerMockRecorder) HandleInitialDuties(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleInitialDuties", reflect.TypeOf((*MockdutyHandler)(nil).HandleInitialDuties), arg0)
}

// Name mocks base method.
func (m *MockdutyHandler) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockdutyHandlerMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockdutyHandler)(nil).Name))
}

// Setup mocks base method.
func (m *MockdutyHandler) Setup(arg0 string, arg1 *zap.Logger, arg2 BeaconNode, arg3 ExecutionClient, arg4 networkconfig.NetworkConfig, arg5 ValidatorController, arg6 ExecuteDutiesFunc, arg7 slotticker.Provider, arg8 chan ReorgEvent, arg9 chan struct{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Setup", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
}

// Setup indicates an expected call of Setup.
func (mr *MockdutyHandlerMockRecorder) Setup(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockdutyHandler)(nil).Setup), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
}
