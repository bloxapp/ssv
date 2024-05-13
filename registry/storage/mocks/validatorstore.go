// Code generated by MockGen. DO NOT EDIT.
// Source: ./validatorstore.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	phase0 "github.com/attestantio/go-eth2-client/spec/phase0"
	types "github.com/bloxapp/ssv/protocol/v2/types"
	storage "github.com/bloxapp/ssv/registry/storage"
	gomock "github.com/golang/mock/gomock"
	types0 "github.com/ssvlabs/ssv-spec/types"
)

// MockValidatorStore is a mock of ValidatorStore interface.
type MockValidatorStore struct {
	ctrl     *gomock.Controller
	recorder *MockValidatorStoreMockRecorder
}

// MockValidatorStoreMockRecorder is the mock recorder for MockValidatorStore.
type MockValidatorStoreMockRecorder struct {
	mock *MockValidatorStore
}

// NewMockValidatorStore creates a new mock instance.
func NewMockValidatorStore(ctrl *gomock.Controller) *MockValidatorStore {
	mock := &MockValidatorStore{ctrl: ctrl}
	mock.recorder = &MockValidatorStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockValidatorStore) EXPECT() *MockValidatorStoreMockRecorder {
	return m.recorder
}

// Committee mocks base method.
func (m *MockValidatorStore) Committee(id types0.ClusterID) *storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Committee", id)
	ret0, _ := ret[0].(*storage.Committee)
	return ret0
}

// Committee indicates an expected call of Committee.
func (mr *MockValidatorStoreMockRecorder) Committee(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Committee", reflect.TypeOf((*MockValidatorStore)(nil).Committee), id)
}

// Committees mocks base method.
func (m *MockValidatorStore) Committees() []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Committees")
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// Committees indicates an expected call of Committees.
func (mr *MockValidatorStoreMockRecorder) Committees() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Committees", reflect.TypeOf((*MockValidatorStore)(nil).Committees))
}

// OperatorCommittees mocks base method.
func (m *MockValidatorStore) OperatorCommittees(id types0.OperatorID) []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OperatorCommittees", id)
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// OperatorCommittees indicates an expected call of OperatorCommittees.
func (mr *MockValidatorStoreMockRecorder) OperatorCommittees(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OperatorCommittees", reflect.TypeOf((*MockValidatorStore)(nil).OperatorCommittees), id)
}

// OperatorValidators mocks base method.
func (m *MockValidatorStore) OperatorValidators(id types0.OperatorID) []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OperatorValidators", id)
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// OperatorValidators indicates an expected call of OperatorValidators.
func (mr *MockValidatorStoreMockRecorder) OperatorValidators(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OperatorValidators", reflect.TypeOf((*MockValidatorStore)(nil).OperatorValidators), id)
}

// ParticipatingCommittees mocks base method.
func (m *MockValidatorStore) ParticipatingCommittees(epoch phase0.Epoch) []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParticipatingCommittees", epoch)
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// ParticipatingCommittees indicates an expected call of ParticipatingCommittees.
func (mr *MockValidatorStoreMockRecorder) ParticipatingCommittees(epoch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParticipatingCommittees", reflect.TypeOf((*MockValidatorStore)(nil).ParticipatingCommittees), epoch)
}

// ParticipatingValidators mocks base method.
func (m *MockValidatorStore) ParticipatingValidators(epoch phase0.Epoch) []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParticipatingValidators", epoch)
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// ParticipatingValidators indicates an expected call of ParticipatingValidators.
func (mr *MockValidatorStoreMockRecorder) ParticipatingValidators(epoch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParticipatingValidators", reflect.TypeOf((*MockValidatorStore)(nil).ParticipatingValidators), epoch)
}

// Validator mocks base method.
func (m *MockValidatorStore) Validator(pubKey []byte) *types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validator", pubKey)
	ret0, _ := ret[0].(*types.SSVShare)
	return ret0
}

// Validator indicates an expected call of Validator.
func (mr *MockValidatorStoreMockRecorder) Validator(pubKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validator", reflect.TypeOf((*MockValidatorStore)(nil).Validator), pubKey)
}

// ValidatorByIndex mocks base method.
func (m *MockValidatorStore) ValidatorByIndex(index phase0.ValidatorIndex) *types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorByIndex", index)
	ret0, _ := ret[0].(*types.SSVShare)
	return ret0
}

// ValidatorByIndex indicates an expected call of ValidatorByIndex.
func (mr *MockValidatorStoreMockRecorder) ValidatorByIndex(index interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorByIndex", reflect.TypeOf((*MockValidatorStore)(nil).ValidatorByIndex), index)
}

// Validators mocks base method.
func (m *MockValidatorStore) Validators() []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validators")
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// Validators indicates an expected call of Validators.
func (mr *MockValidatorStoreMockRecorder) Validators() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validators", reflect.TypeOf((*MockValidatorStore)(nil).Validators))
}

// WithOperatorID mocks base method.
func (m *MockValidatorStore) WithOperatorID(operatorID func() types0.OperatorID) storage.SelfValidatorStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithOperatorID", operatorID)
	ret0, _ := ret[0].(storage.SelfValidatorStore)
	return ret0
}

// WithOperatorID indicates an expected call of WithOperatorID.
func (mr *MockValidatorStoreMockRecorder) WithOperatorID(operatorID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithOperatorID", reflect.TypeOf((*MockValidatorStore)(nil).WithOperatorID), operatorID)
}

// MockSelfValidatorStore is a mock of SelfValidatorStore interface.
type MockSelfValidatorStore struct {
	ctrl     *gomock.Controller
	recorder *MockSelfValidatorStoreMockRecorder
}

// MockSelfValidatorStoreMockRecorder is the mock recorder for MockSelfValidatorStore.
type MockSelfValidatorStoreMockRecorder struct {
	mock *MockSelfValidatorStore
}

// NewMockSelfValidatorStore creates a new mock instance.
func NewMockSelfValidatorStore(ctrl *gomock.Controller) *MockSelfValidatorStore {
	mock := &MockSelfValidatorStore{ctrl: ctrl}
	mock.recorder = &MockSelfValidatorStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSelfValidatorStore) EXPECT() *MockSelfValidatorStoreMockRecorder {
	return m.recorder
}

// Committee mocks base method.
func (m *MockSelfValidatorStore) Committee(id types0.ClusterID) *storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Committee", id)
	ret0, _ := ret[0].(*storage.Committee)
	return ret0
}

// Committee indicates an expected call of Committee.
func (mr *MockSelfValidatorStoreMockRecorder) Committee(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Committee", reflect.TypeOf((*MockSelfValidatorStore)(nil).Committee), id)
}

// Committees mocks base method.
func (m *MockSelfValidatorStore) Committees() []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Committees")
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// Committees indicates an expected call of Committees.
func (mr *MockSelfValidatorStoreMockRecorder) Committees() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Committees", reflect.TypeOf((*MockSelfValidatorStore)(nil).Committees))
}

// OperatorCommittees mocks base method.
func (m *MockSelfValidatorStore) OperatorCommittees(id types0.OperatorID) []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OperatorCommittees", id)
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// OperatorCommittees indicates an expected call of OperatorCommittees.
func (mr *MockSelfValidatorStoreMockRecorder) OperatorCommittees(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OperatorCommittees", reflect.TypeOf((*MockSelfValidatorStore)(nil).OperatorCommittees), id)
}

// OperatorValidators mocks base method.
func (m *MockSelfValidatorStore) OperatorValidators(id types0.OperatorID) []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OperatorValidators", id)
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// OperatorValidators indicates an expected call of OperatorValidators.
func (mr *MockSelfValidatorStoreMockRecorder) OperatorValidators(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OperatorValidators", reflect.TypeOf((*MockSelfValidatorStore)(nil).OperatorValidators), id)
}

// ParticipatingCommittees mocks base method.
func (m *MockSelfValidatorStore) ParticipatingCommittees(epoch phase0.Epoch) []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParticipatingCommittees", epoch)
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// ParticipatingCommittees indicates an expected call of ParticipatingCommittees.
func (mr *MockSelfValidatorStoreMockRecorder) ParticipatingCommittees(epoch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParticipatingCommittees", reflect.TypeOf((*MockSelfValidatorStore)(nil).ParticipatingCommittees), epoch)
}

// ParticipatingValidators mocks base method.
func (m *MockSelfValidatorStore) ParticipatingValidators(epoch phase0.Epoch) []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParticipatingValidators", epoch)
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// ParticipatingValidators indicates an expected call of ParticipatingValidators.
func (mr *MockSelfValidatorStoreMockRecorder) ParticipatingValidators(epoch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParticipatingValidators", reflect.TypeOf((*MockSelfValidatorStore)(nil).ParticipatingValidators), epoch)
}

// SelfCommittees mocks base method.
func (m *MockSelfValidatorStore) SelfCommittees() []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelfCommittees")
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// SelfCommittees indicates an expected call of SelfCommittees.
func (mr *MockSelfValidatorStoreMockRecorder) SelfCommittees() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelfCommittees", reflect.TypeOf((*MockSelfValidatorStore)(nil).SelfCommittees))
}

// SelfParticipatingCommittees mocks base method.
func (m *MockSelfValidatorStore) SelfParticipatingCommittees(arg0 phase0.Epoch) []*storage.Committee {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelfParticipatingCommittees", arg0)
	ret0, _ := ret[0].([]*storage.Committee)
	return ret0
}

// SelfParticipatingCommittees indicates an expected call of SelfParticipatingCommittees.
func (mr *MockSelfValidatorStoreMockRecorder) SelfParticipatingCommittees(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelfParticipatingCommittees", reflect.TypeOf((*MockSelfValidatorStore)(nil).SelfParticipatingCommittees), arg0)
}

// SelfParticipatingValidators mocks base method.
func (m *MockSelfValidatorStore) SelfParticipatingValidators(arg0 phase0.Epoch) []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelfParticipatingValidators", arg0)
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// SelfParticipatingValidators indicates an expected call of SelfParticipatingValidators.
func (mr *MockSelfValidatorStoreMockRecorder) SelfParticipatingValidators(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelfParticipatingValidators", reflect.TypeOf((*MockSelfValidatorStore)(nil).SelfParticipatingValidators), arg0)
}

// SelfValidators mocks base method.
func (m *MockSelfValidatorStore) SelfValidators() []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelfValidators")
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// SelfValidators indicates an expected call of SelfValidators.
func (mr *MockSelfValidatorStoreMockRecorder) SelfValidators() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelfValidators", reflect.TypeOf((*MockSelfValidatorStore)(nil).SelfValidators))
}

// Validator mocks base method.
func (m *MockSelfValidatorStore) Validator(pubKey []byte) *types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validator", pubKey)
	ret0, _ := ret[0].(*types.SSVShare)
	return ret0
}

// Validator indicates an expected call of Validator.
func (mr *MockSelfValidatorStoreMockRecorder) Validator(pubKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validator", reflect.TypeOf((*MockSelfValidatorStore)(nil).Validator), pubKey)
}

// ValidatorByIndex mocks base method.
func (m *MockSelfValidatorStore) ValidatorByIndex(index phase0.ValidatorIndex) *types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorByIndex", index)
	ret0, _ := ret[0].(*types.SSVShare)
	return ret0
}

// ValidatorByIndex indicates an expected call of ValidatorByIndex.
func (mr *MockSelfValidatorStoreMockRecorder) ValidatorByIndex(index interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorByIndex", reflect.TypeOf((*MockSelfValidatorStore)(nil).ValidatorByIndex), index)
}

// Validators mocks base method.
func (m *MockSelfValidatorStore) Validators() []*types.SSVShare {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validators")
	ret0, _ := ret[0].([]*types.SSVShare)
	return ret0
}

// Validators indicates an expected call of Validators.
func (mr *MockSelfValidatorStoreMockRecorder) Validators() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validators", reflect.TypeOf((*MockSelfValidatorStore)(nil).Validators))
}

// WithOperatorID mocks base method.
func (m *MockSelfValidatorStore) WithOperatorID(operatorID func() types0.OperatorID) storage.SelfValidatorStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithOperatorID", operatorID)
	ret0, _ := ret[0].(storage.SelfValidatorStore)
	return ret0
}

// WithOperatorID indicates an expected call of WithOperatorID.
func (mr *MockSelfValidatorStoreMockRecorder) WithOperatorID(operatorID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithOperatorID", reflect.TypeOf((*MockSelfValidatorStore)(nil).WithOperatorID), operatorID)
}
