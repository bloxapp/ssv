// Code generated by MockGen. DO NOT EDIT.
// Source: ./syncer.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	qbft "github.com/bloxapp/ssv-spec/qbft"
	types "github.com/bloxapp/ssv-spec/types"
	syncing "github.com/bloxapp/ssv/network/syncing"
	protocolp2p "github.com/bloxapp/ssv/protocol/v2/p2p"
	gomock "github.com/golang/mock/gomock"
	zap "go.uber.org/zap"
)

// MockSyncer is a mock of Syncer interface.
type MockSyncer struct {
	ctrl     *gomock.Controller
	recorder *MockSyncerMockRecorder
}

// MockSyncerMockRecorder is the mock recorder for MockSyncer.
type MockSyncerMockRecorder struct {
	mock *MockSyncer
}

// NewMockSyncer creates a new mock instance.
func NewMockSyncer(ctrl *gomock.Controller) *MockSyncer {
	mock := &MockSyncer{ctrl: ctrl}
	mock.recorder = &MockSyncerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSyncer) EXPECT() *MockSyncerMockRecorder {
	return m.recorder
}

// SyncDecidedByRange mocks base method.
func (m *MockSyncer) SyncDecidedByRange(ctx context.Context, logger *zap.Logger, id types.MessageID, from, to qbft.Height, handler syncing.MessageHandler) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncDecidedByRange", ctx, logger, id, from, to, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncDecidedByRange indicates an expected call of SyncDecidedByRange.
func (mr *MockSyncerMockRecorder) SyncDecidedByRange(ctx, logger, id, from, to, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncDecidedByRange", reflect.TypeOf((*MockSyncer)(nil).SyncDecidedByRange), ctx, logger, id, from, to, handler)
}

// SyncHighestDecided mocks base method.
func (m *MockSyncer) SyncHighestDecided(ctx context.Context, logger *zap.Logger, id types.MessageID, handler syncing.MessageHandler) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncHighestDecided", ctx, logger, id, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncHighestDecided indicates an expected call of SyncHighestDecided.
func (mr *MockSyncerMockRecorder) SyncHighestDecided(ctx, logger, id, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncHighestDecided", reflect.TypeOf((*MockSyncer)(nil).SyncHighestDecided), ctx, logger, id, handler)
}

// MockNetwork is a mock of Network interface.
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork.
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance.
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// GetHistory mocks base method.
func (m *MockNetwork) GetHistory(logger *zap.Logger, id types.MessageID, from, to qbft.Height, targets ...string) ([]protocolp2p.SyncResult, qbft.Height, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{logger, id, from, to}
	for _, a := range targets {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetHistory", varargs...)
	ret0, _ := ret[0].([]protocolp2p.SyncResult)
	ret1, _ := ret[1].(qbft.Height)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetHistory indicates an expected call of GetHistory.
func (mr *MockNetworkMockRecorder) GetHistory(logger, id, from, to interface{}, targets ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{logger, id, from, to}, targets...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHistory", reflect.TypeOf((*MockNetwork)(nil).GetHistory), varargs...)
}

// LastDecided mocks base method.
func (m *MockNetwork) LastDecided(logger *zap.Logger, id types.MessageID) ([]protocolp2p.SyncResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastDecided", logger, id)
	ret0, _ := ret[0].([]protocolp2p.SyncResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastDecided indicates an expected call of LastDecided.
func (mr *MockNetworkMockRecorder) LastDecided(logger, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastDecided", reflect.TypeOf((*MockNetwork)(nil).LastDecided), logger, id)
}
