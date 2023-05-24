// Code generated by MockGen. DO NOT EDIT.
// Source: ./client.go

// Package eth1 is a generated GoMock package.
package eth1

import (
	big "math/big"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	event "github.com/prysmaticlabs/prysm/async/event"
	zap "go.uber.org/zap"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// EventsFeed mocks base method.
func (m *MockClient) EventsFeed() *event.Feed {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventsFeed")
	ret0, _ := ret[0].(*event.Feed)
	return ret0
}

// EventsFeed indicates an expected call of EventsFeed.
func (mr *MockClientMockRecorder) EventsFeed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventsFeed", reflect.TypeOf((*MockClient)(nil).EventsFeed))
}

// SetControllerEventHandler mocks base method.
func (m *MockClient) SetControllerEventHandler(handler SyncEventHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetControllerEventHandler", handler)
}

// SetControllerEventHandler indicates an expected call of SetControllerEventHandler.
func (mr *MockClientMockRecorder) SetControllerEventHandler(handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetControllerEventHandler", reflect.TypeOf((*MockClient)(nil).SetControllerEventHandler), handler)
}

// Start mocks base method.
func (m *MockClient) Start(logger *zap.Logger) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", logger)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockClientMockRecorder) Start(logger interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockClient)(nil).Start), logger)
}

// Sync mocks base method.
func (m *MockClient) Sync(logger *zap.Logger, fromBlock *big.Int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", logger, fromBlock)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockClientMockRecorder) Sync(logger, fromBlock interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockClient)(nil).Sync), logger, fromBlock)
}
