// Code generated by MockGen. DO NOT EDIT.
// Source: ./client.go

// Package beacon is a generated GoMock package.
package beacon

import (
	context "context"
	reflect "reflect"

	client "github.com/attestantio/go-eth2-client"
	api "github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec"
	altair "github.com/attestantio/go-eth2-client/spec/altair"
	bellatrix "github.com/attestantio/go-eth2-client/spec/bellatrix"
	phase0 "github.com/attestantio/go-eth2-client/spec/phase0"
	ssz "github.com/ferranbt/fastssz"
	gomock "github.com/golang/mock/gomock"
	types "github.com/ssvlabs/ssv-spec/types"
)

// MockbeaconDuties is a mock of beaconDuties interface.
type MockbeaconDuties struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconDutiesMockRecorder
}

// MockbeaconDutiesMockRecorder is the mock recorder for MockbeaconDuties.
type MockbeaconDutiesMockRecorder struct {
	mock *MockbeaconDuties
}

// NewMockbeaconDuties creates a new mock instance.
func NewMockbeaconDuties(ctrl *gomock.Controller) *MockbeaconDuties {
	mock := &MockbeaconDuties{ctrl: ctrl}
	mock.recorder = &MockbeaconDutiesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconDuties) EXPECT() *MockbeaconDutiesMockRecorder {
	return m.recorder
}

// AttesterDuties mocks base method.
func (m *MockbeaconDuties) AttesterDuties(ctx context.Context, epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*v1.AttesterDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttesterDuties", ctx, epoch, validatorIndices)
	ret0, _ := ret[0].([]*v1.AttesterDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttesterDuties indicates an expected call of AttesterDuties.
func (mr *MockbeaconDutiesMockRecorder) AttesterDuties(ctx, epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttesterDuties", reflect.TypeOf((*MockbeaconDuties)(nil).AttesterDuties), ctx, epoch, validatorIndices)
}

// Events mocks base method.
func (m *MockbeaconDuties) Events(ctx context.Context, topics []string, handler client.EventHandlerFunc) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Events", ctx, topics, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// Events indicates an expected call of Events.
func (mr *MockbeaconDutiesMockRecorder) Events(ctx, topics, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockbeaconDuties)(nil).Events), ctx, topics, handler)
}

// ProposerDuties mocks base method.
func (m *MockbeaconDuties) ProposerDuties(ctx context.Context, epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*v1.ProposerDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProposerDuties", ctx, epoch, validatorIndices)
	ret0, _ := ret[0].([]*v1.ProposerDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProposerDuties indicates an expected call of ProposerDuties.
func (mr *MockbeaconDutiesMockRecorder) ProposerDuties(ctx, epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProposerDuties", reflect.TypeOf((*MockbeaconDuties)(nil).ProposerDuties), ctx, epoch, validatorIndices)
}

// SyncCommitteeDuties mocks base method.
func (m *MockbeaconDuties) SyncCommitteeDuties(ctx context.Context, epoch phase0.Epoch, indices []phase0.ValidatorIndex) ([]*v1.SyncCommitteeDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncCommitteeDuties", ctx, epoch, indices)
	ret0, _ := ret[0].([]*v1.SyncCommitteeDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncCommitteeDuties indicates an expected call of SyncCommitteeDuties.
func (mr *MockbeaconDutiesMockRecorder) SyncCommitteeDuties(ctx, epoch, indices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncCommitteeDuties", reflect.TypeOf((*MockbeaconDuties)(nil).SyncCommitteeDuties), ctx, epoch, indices)
}

// MockbeaconSubscriber is a mock of beaconSubscriber interface.
type MockbeaconSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconSubscriberMockRecorder
}

// MockbeaconSubscriberMockRecorder is the mock recorder for MockbeaconSubscriber.
type MockbeaconSubscriberMockRecorder struct {
	mock *MockbeaconSubscriber
}

// NewMockbeaconSubscriber creates a new mock instance.
func NewMockbeaconSubscriber(ctrl *gomock.Controller) *MockbeaconSubscriber {
	mock := &MockbeaconSubscriber{ctrl: ctrl}
	mock.recorder = &MockbeaconSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconSubscriber) EXPECT() *MockbeaconSubscriberMockRecorder {
	return m.recorder
}

// SubmitBeaconCommitteeSubscriptions mocks base method.
func (m *MockbeaconSubscriber) SubmitBeaconCommitteeSubscriptions(ctx context.Context, subscription []*v1.BeaconCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBeaconCommitteeSubscriptions", ctx, subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBeaconCommitteeSubscriptions indicates an expected call of SubmitBeaconCommitteeSubscriptions.
func (mr *MockbeaconSubscriberMockRecorder) SubmitBeaconCommitteeSubscriptions(ctx, subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBeaconCommitteeSubscriptions", reflect.TypeOf((*MockbeaconSubscriber)(nil).SubmitBeaconCommitteeSubscriptions), ctx, subscription)
}

// SubmitSyncCommitteeSubscriptions mocks base method.
func (m *MockbeaconSubscriber) SubmitSyncCommitteeSubscriptions(ctx context.Context, subscription []*v1.SyncCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncCommitteeSubscriptions", ctx, subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncCommitteeSubscriptions indicates an expected call of SubmitSyncCommitteeSubscriptions.
func (mr *MockbeaconSubscriberMockRecorder) SubmitSyncCommitteeSubscriptions(ctx, subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncCommitteeSubscriptions", reflect.TypeOf((*MockbeaconSubscriber)(nil).SubmitSyncCommitteeSubscriptions), ctx, subscription)
}

// MockbeaconValidator is a mock of beaconValidator interface.
type MockbeaconValidator struct {
	ctrl     *gomock.Controller
	recorder *MockbeaconValidatorMockRecorder
}

// MockbeaconValidatorMockRecorder is the mock recorder for MockbeaconValidator.
type MockbeaconValidatorMockRecorder struct {
	mock *MockbeaconValidator
}

// NewMockbeaconValidator creates a new mock instance.
func NewMockbeaconValidator(ctrl *gomock.Controller) *MockbeaconValidator {
	mock := &MockbeaconValidator{ctrl: ctrl}
	mock.recorder = &MockbeaconValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockbeaconValidator) EXPECT() *MockbeaconValidatorMockRecorder {
	return m.recorder
}

// GetValidatorData mocks base method.
func (m *MockbeaconValidator) GetValidatorData(validatorPubKeys []phase0.BLSPubKey) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorData", validatorPubKeys)
	ret0, _ := ret[0].(map[phase0.ValidatorIndex]*v1.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorData indicates an expected call of GetValidatorData.
func (mr *MockbeaconValidatorMockRecorder) GetValidatorData(validatorPubKeys interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorData", reflect.TypeOf((*MockbeaconValidator)(nil).GetValidatorData), validatorPubKeys)
}

// Mockproposer is a mock of proposer interface.
type Mockproposer struct {
	ctrl     *gomock.Controller
	recorder *MockproposerMockRecorder
}

// MockproposerMockRecorder is the mock recorder for Mockproposer.
type MockproposerMockRecorder struct {
	mock *Mockproposer
}

// NewMockproposer creates a new mock instance.
func NewMockproposer(ctrl *gomock.Controller) *Mockproposer {
	mock := &Mockproposer{ctrl: ctrl}
	mock.recorder = &MockproposerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockproposer) EXPECT() *MockproposerMockRecorder {
	return m.recorder
}

// SubmitProposalPreparation mocks base method.
func (m *Mockproposer) SubmitProposalPreparation(feeRecipients map[phase0.ValidatorIndex]bellatrix.ExecutionAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitProposalPreparation", feeRecipients)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitProposalPreparation indicates an expected call of SubmitProposalPreparation.
func (mr *MockproposerMockRecorder) SubmitProposalPreparation(feeRecipients interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitProposalPreparation", reflect.TypeOf((*Mockproposer)(nil).SubmitProposalPreparation), feeRecipients)
}

// Mocksigner is a mock of signer interface.
type Mocksigner struct {
	ctrl     *gomock.Controller
	recorder *MocksignerMockRecorder
}

// MocksignerMockRecorder is the mock recorder for Mocksigner.
type MocksignerMockRecorder struct {
	mock *Mocksigner
}

// NewMocksigner creates a new mock instance.
func NewMocksigner(ctrl *gomock.Controller) *Mocksigner {
	mock := &Mocksigner{ctrl: ctrl}
	mock.recorder = &MocksignerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mocksigner) EXPECT() *MocksignerMockRecorder {
	return m.recorder
}

// ComputeSigningRoot mocks base method.
func (m *Mocksigner) ComputeSigningRoot(object interface{}, domain phase0.Domain) ([32]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputeSigningRoot", object, domain)
	ret0, _ := ret[0].([32]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ComputeSigningRoot indicates an expected call of ComputeSigningRoot.
func (mr *MocksignerMockRecorder) ComputeSigningRoot(object, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputeSigningRoot", reflect.TypeOf((*Mocksigner)(nil).ComputeSigningRoot), object, domain)
}

// MockBeaconNode is a mock of BeaconNode interface.
type MockBeaconNode struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconNodeMockRecorder
}

// MockBeaconNodeMockRecorder is the mock recorder for MockBeaconNode.
type MockBeaconNodeMockRecorder struct {
	mock *MockBeaconNode
}

// NewMockBeaconNode creates a new mock instance.
func NewMockBeaconNode(ctrl *gomock.Controller) *MockBeaconNode {
	mock := &MockBeaconNode{ctrl: ctrl}
	mock.recorder = &MockBeaconNodeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeaconNode) EXPECT() *MockBeaconNodeMockRecorder {
	return m.recorder
}

// AttesterDuties mocks base method.
func (m *MockBeaconNode) AttesterDuties(ctx context.Context, epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*v1.AttesterDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttesterDuties", ctx, epoch, validatorIndices)
	ret0, _ := ret[0].([]*v1.AttesterDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttesterDuties indicates an expected call of AttesterDuties.
func (mr *MockBeaconNodeMockRecorder) AttesterDuties(ctx, epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttesterDuties", reflect.TypeOf((*MockBeaconNode)(nil).AttesterDuties), ctx, epoch, validatorIndices)
}

// ComputeSigningRoot mocks base method.
func (m *MockBeaconNode) ComputeSigningRoot(object interface{}, domain phase0.Domain) ([32]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputeSigningRoot", object, domain)
	ret0, _ := ret[0].([32]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ComputeSigningRoot indicates an expected call of ComputeSigningRoot.
func (mr *MockBeaconNodeMockRecorder) ComputeSigningRoot(object, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputeSigningRoot", reflect.TypeOf((*MockBeaconNode)(nil).ComputeSigningRoot), object, domain)
}

// DomainData mocks base method.
func (m *MockBeaconNode) DomainData(epoch phase0.Epoch, domain phase0.DomainType) (phase0.Domain, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DomainData", epoch, domain)
	ret0, _ := ret[0].(phase0.Domain)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DomainData indicates an expected call of DomainData.
func (mr *MockBeaconNodeMockRecorder) DomainData(epoch, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DomainData", reflect.TypeOf((*MockBeaconNode)(nil).DomainData), epoch, domain)
}

// Events mocks base method.
func (m *MockBeaconNode) Events(ctx context.Context, topics []string, handler client.EventHandlerFunc) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Events", ctx, topics, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// Events indicates an expected call of Events.
func (mr *MockBeaconNodeMockRecorder) Events(ctx, topics, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockBeaconNode)(nil).Events), ctx, topics, handler)
}

// GetAttestationData mocks base method.
func (m *MockBeaconNode) GetAttestationData(slot phase0.Slot, committeeIndex phase0.CommitteeIndex) (*phase0.AttestationData, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttestationData", slot, committeeIndex)
	ret0, _ := ret[0].(*phase0.AttestationData)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetAttestationData indicates an expected call of GetAttestationData.
func (mr *MockBeaconNodeMockRecorder) GetAttestationData(slot, committeeIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttestationData", reflect.TypeOf((*MockBeaconNode)(nil).GetAttestationData), slot, committeeIndex)
}

// GetBeaconBlock mocks base method.
func (m *MockBeaconNode) GetBeaconBlock(slot phase0.Slot, graffiti, randao []byte) (ssz.Marshaler, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBeaconBlock", slot, graffiti, randao)
	ret0, _ := ret[0].(ssz.Marshaler)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetBeaconBlock indicates an expected call of GetBeaconBlock.
func (mr *MockBeaconNodeMockRecorder) GetBeaconBlock(slot, graffiti, randao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBeaconBlock", reflect.TypeOf((*MockBeaconNode)(nil).GetBeaconBlock), slot, graffiti, randao)
}

// GetBeaconNetwork mocks base method.
func (m *MockBeaconNode) GetBeaconNetwork() types.BeaconNetwork {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBeaconNetwork")
	ret0, _ := ret[0].(types.BeaconNetwork)
	return ret0
}

// GetBeaconNetwork indicates an expected call of GetBeaconNetwork.
func (mr *MockBeaconNodeMockRecorder) GetBeaconNetwork() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBeaconNetwork", reflect.TypeOf((*MockBeaconNode)(nil).GetBeaconNetwork))
}

// GetBlindedBeaconBlock mocks base method.
func (m *MockBeaconNode) GetBlindedBeaconBlock(slot phase0.Slot, graffiti, randao []byte) (ssz.Marshaler, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlindedBeaconBlock", slot, graffiti, randao)
	ret0, _ := ret[0].(ssz.Marshaler)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetBlindedBeaconBlock indicates an expected call of GetBlindedBeaconBlock.
func (mr *MockBeaconNodeMockRecorder) GetBlindedBeaconBlock(slot, graffiti, randao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlindedBeaconBlock", reflect.TypeOf((*MockBeaconNode)(nil).GetBlindedBeaconBlock), slot, graffiti, randao)
}

// GetSyncCommitteeContribution mocks base method.
func (m *MockBeaconNode) GetSyncCommitteeContribution(slot phase0.Slot, selectionProofs []phase0.BLSSignature, subnetIDs []uint64) (ssz.Marshaler, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSyncCommitteeContribution", slot, selectionProofs, subnetIDs)
	ret0, _ := ret[0].(ssz.Marshaler)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetSyncCommitteeContribution indicates an expected call of GetSyncCommitteeContribution.
func (mr *MockBeaconNodeMockRecorder) GetSyncCommitteeContribution(slot, selectionProofs, subnetIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncCommitteeContribution", reflect.TypeOf((*MockBeaconNode)(nil).GetSyncCommitteeContribution), slot, selectionProofs, subnetIDs)
}

// GetSyncMessageBlockRoot mocks base method.
func (m *MockBeaconNode) GetSyncMessageBlockRoot(slot phase0.Slot) (phase0.Root, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSyncMessageBlockRoot", slot)
	ret0, _ := ret[0].(phase0.Root)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetSyncMessageBlockRoot indicates an expected call of GetSyncMessageBlockRoot.
func (mr *MockBeaconNodeMockRecorder) GetSyncMessageBlockRoot(slot interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncMessageBlockRoot", reflect.TypeOf((*MockBeaconNode)(nil).GetSyncMessageBlockRoot), slot)
}

// GetValidatorData mocks base method.
func (m *MockBeaconNode) GetValidatorData(validatorPubKeys []phase0.BLSPubKey) (map[phase0.ValidatorIndex]*v1.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorData", validatorPubKeys)
	ret0, _ := ret[0].(map[phase0.ValidatorIndex]*v1.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValidatorData indicates an expected call of GetValidatorData.
func (mr *MockBeaconNodeMockRecorder) GetValidatorData(validatorPubKeys interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorData", reflect.TypeOf((*MockBeaconNode)(nil).GetValidatorData), validatorPubKeys)
}

// IsSyncCommitteeAggregator mocks base method.
func (m *MockBeaconNode) IsSyncCommitteeAggregator(proof []byte) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSyncCommitteeAggregator", proof)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsSyncCommitteeAggregator indicates an expected call of IsSyncCommitteeAggregator.
func (mr *MockBeaconNodeMockRecorder) IsSyncCommitteeAggregator(proof interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSyncCommitteeAggregator", reflect.TypeOf((*MockBeaconNode)(nil).IsSyncCommitteeAggregator), proof)
}

// ProposerDuties mocks base method.
func (m *MockBeaconNode) ProposerDuties(ctx context.Context, epoch phase0.Epoch, validatorIndices []phase0.ValidatorIndex) ([]*v1.ProposerDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProposerDuties", ctx, epoch, validatorIndices)
	ret0, _ := ret[0].([]*v1.ProposerDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProposerDuties indicates an expected call of ProposerDuties.
func (mr *MockBeaconNodeMockRecorder) ProposerDuties(ctx, epoch, validatorIndices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProposerDuties", reflect.TypeOf((*MockBeaconNode)(nil).ProposerDuties), ctx, epoch, validatorIndices)
}

// SubmitAggregateSelectionProof mocks base method.
func (m *MockBeaconNode) SubmitAggregateSelectionProof(slot phase0.Slot, committeeIndex phase0.CommitteeIndex, committeeLength uint64, index phase0.ValidatorIndex, slotSig []byte) (ssz.Marshaler, spec.DataVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitAggregateSelectionProof", slot, committeeIndex, committeeLength, index, slotSig)
	ret0, _ := ret[0].(ssz.Marshaler)
	ret1, _ := ret[1].(spec.DataVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SubmitAggregateSelectionProof indicates an expected call of SubmitAggregateSelectionProof.
func (mr *MockBeaconNodeMockRecorder) SubmitAggregateSelectionProof(slot, committeeIndex, committeeLength, index, slotSig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitAggregateSelectionProof", reflect.TypeOf((*MockBeaconNode)(nil).SubmitAggregateSelectionProof), slot, committeeIndex, committeeLength, index, slotSig)
}

// SubmitAttestation mocks base method.
func (m *MockBeaconNode) SubmitAttestation(attestation *phase0.Attestation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitAttestation", attestation)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitAttestation indicates an expected call of SubmitAttestation.
func (mr *MockBeaconNodeMockRecorder) SubmitAttestation(attestation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitAttestation", reflect.TypeOf((*MockBeaconNode)(nil).SubmitAttestation), attestation)
}

// SubmitBeaconBlock mocks base method.
func (m *MockBeaconNode) SubmitBeaconBlock(block *api.VersionedProposal, sig phase0.BLSSignature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBeaconBlock", block, sig)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBeaconBlock indicates an expected call of SubmitBeaconBlock.
func (mr *MockBeaconNodeMockRecorder) SubmitBeaconBlock(block, sig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBeaconBlock", reflect.TypeOf((*MockBeaconNode)(nil).SubmitBeaconBlock), block, sig)
}

// SubmitBeaconCommitteeSubscriptions mocks base method.
func (m *MockBeaconNode) SubmitBeaconCommitteeSubscriptions(ctx context.Context, subscription []*v1.BeaconCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBeaconCommitteeSubscriptions", ctx, subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBeaconCommitteeSubscriptions indicates an expected call of SubmitBeaconCommitteeSubscriptions.
func (mr *MockBeaconNodeMockRecorder) SubmitBeaconCommitteeSubscriptions(ctx, subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBeaconCommitteeSubscriptions", reflect.TypeOf((*MockBeaconNode)(nil).SubmitBeaconCommitteeSubscriptions), ctx, subscription)
}

// SubmitBlindedBeaconBlock mocks base method.
func (m *MockBeaconNode) SubmitBlindedBeaconBlock(block *api.VersionedBlindedProposal, sig phase0.BLSSignature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitBlindedBeaconBlock", block, sig)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitBlindedBeaconBlock indicates an expected call of SubmitBlindedBeaconBlock.
func (mr *MockBeaconNodeMockRecorder) SubmitBlindedBeaconBlock(block, sig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitBlindedBeaconBlock", reflect.TypeOf((*MockBeaconNode)(nil).SubmitBlindedBeaconBlock), block, sig)
}

// SubmitProposalPreparation mocks base method.
func (m *MockBeaconNode) SubmitProposalPreparation(feeRecipients map[phase0.ValidatorIndex]bellatrix.ExecutionAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitProposalPreparation", feeRecipients)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitProposalPreparation indicates an expected call of SubmitProposalPreparation.
func (mr *MockBeaconNodeMockRecorder) SubmitProposalPreparation(feeRecipients interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitProposalPreparation", reflect.TypeOf((*MockBeaconNode)(nil).SubmitProposalPreparation), feeRecipients)
}

// SubmitSignedAggregateSelectionProof mocks base method.
func (m *MockBeaconNode) SubmitSignedAggregateSelectionProof(msg *phase0.SignedAggregateAndProof) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSignedAggregateSelectionProof", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSignedAggregateSelectionProof indicates an expected call of SubmitSignedAggregateSelectionProof.
func (mr *MockBeaconNodeMockRecorder) SubmitSignedAggregateSelectionProof(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSignedAggregateSelectionProof", reflect.TypeOf((*MockBeaconNode)(nil).SubmitSignedAggregateSelectionProof), msg)
}

// SubmitSignedContributionAndProof mocks base method.
func (m *MockBeaconNode) SubmitSignedContributionAndProof(contribution *altair.SignedContributionAndProof) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSignedContributionAndProof", contribution)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSignedContributionAndProof indicates an expected call of SubmitSignedContributionAndProof.
func (mr *MockBeaconNodeMockRecorder) SubmitSignedContributionAndProof(contribution interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSignedContributionAndProof", reflect.TypeOf((*MockBeaconNode)(nil).SubmitSignedContributionAndProof), contribution)
}

// SubmitSyncCommitteeSubscriptions mocks base method.
func (m *MockBeaconNode) SubmitSyncCommitteeSubscriptions(ctx context.Context, subscription []*v1.SyncCommitteeSubscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncCommitteeSubscriptions", ctx, subscription)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncCommitteeSubscriptions indicates an expected call of SubmitSyncCommitteeSubscriptions.
func (mr *MockBeaconNodeMockRecorder) SubmitSyncCommitteeSubscriptions(ctx, subscription interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncCommitteeSubscriptions", reflect.TypeOf((*MockBeaconNode)(nil).SubmitSyncCommitteeSubscriptions), ctx, subscription)
}

// SubmitSyncMessage mocks base method.
func (m *MockBeaconNode) SubmitSyncMessage(msg *altair.SyncCommitteeMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitSyncMessage", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitSyncMessage indicates an expected call of SubmitSyncMessage.
func (mr *MockBeaconNodeMockRecorder) SubmitSyncMessage(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitSyncMessage", reflect.TypeOf((*MockBeaconNode)(nil).SubmitSyncMessage), msg)
}

// SubmitValidatorRegistration mocks base method.
func (m *MockBeaconNode) SubmitValidatorRegistration(pubkey []byte, feeRecipient bellatrix.ExecutionAddress, sig phase0.BLSSignature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitValidatorRegistration", pubkey, feeRecipient, sig)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitValidatorRegistration indicates an expected call of SubmitValidatorRegistration.
func (mr *MockBeaconNodeMockRecorder) SubmitValidatorRegistration(pubkey, feeRecipient, sig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitValidatorRegistration", reflect.TypeOf((*MockBeaconNode)(nil).SubmitValidatorRegistration), pubkey, feeRecipient, sig)
}

// SubmitVoluntaryExit mocks base method.
func (m *MockBeaconNode) SubmitVoluntaryExit(voluntaryExit *phase0.SignedVoluntaryExit) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitVoluntaryExit", voluntaryExit)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubmitVoluntaryExit indicates an expected call of SubmitVoluntaryExit.
func (mr *MockBeaconNodeMockRecorder) SubmitVoluntaryExit(voluntaryExit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitVoluntaryExit", reflect.TypeOf((*MockBeaconNode)(nil).SubmitVoluntaryExit), voluntaryExit)
}

// SyncCommitteeDuties mocks base method.
func (m *MockBeaconNode) SyncCommitteeDuties(ctx context.Context, epoch phase0.Epoch, indices []phase0.ValidatorIndex) ([]*v1.SyncCommitteeDuty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncCommitteeDuties", ctx, epoch, indices)
	ret0, _ := ret[0].([]*v1.SyncCommitteeDuty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncCommitteeDuties indicates an expected call of SyncCommitteeDuties.
func (mr *MockBeaconNodeMockRecorder) SyncCommitteeDuties(ctx, epoch, indices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncCommitteeDuties", reflect.TypeOf((*MockBeaconNode)(nil).SyncCommitteeDuties), ctx, epoch, indices)
}

// SyncCommitteeSubnetID mocks base method.
func (m *MockBeaconNode) SyncCommitteeSubnetID(index phase0.CommitteeIndex) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncCommitteeSubnetID", index)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncCommitteeSubnetID indicates an expected call of SyncCommitteeSubnetID.
func (mr *MockBeaconNodeMockRecorder) SyncCommitteeSubnetID(index interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncCommitteeSubnetID", reflect.TypeOf((*MockBeaconNode)(nil).SyncCommitteeSubnetID), index)
}
