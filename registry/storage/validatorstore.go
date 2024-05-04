package storage

import (
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/bloxapp/ssv/protocol/v2/types"
)

type ValidatorStore interface {
	Validator(pubKey []byte) *types.SSVShare
	ValidatorByIndex(index phase0.ValidatorIndex) *types.SSVShare
	Validators() []*types.SSVShare
	ParticipatingValidators(epoch phase0.Epoch) []*types.SSVShare
	OperatorValidators(id spectypes.OperatorID) []*types.SSVShare

	Committee(id types.CommitteeID) *Committee
	Committees() []*Committee
	ParticipatingCommittees(epoch phase0.Epoch) []*Committee
	OperatorCommittees(id spectypes.OperatorID) []*Committee

	// TODO: save recipient address
}

type SelfValidatorStore interface {
	ValidatorStore

	SelfValidators() []*types.SSVShare
	SelfParticipatingValidators(phase0.Epoch) []*types.SSVShare

	SelfCommittees() []*Committee
	SelfParticipatingCommittees(phase0.Epoch) []*Committee
}

type Committee struct {
	ID         types.CommitteeID
	Operators  []spectypes.OperatorID
	Validators []*types.SSVShare
}

// IsParticipating returns whether any validator in the committee should participate in the given epoch.
func (c *Committee) IsParticipating(epoch phase0.Epoch) bool {
	for _, validator := range c.Validators {
		if validator.IsParticipating(epoch) {
			return true
		}
	}
	return false
}

type sharesAndCommittees struct {
	shares     []*types.SSVShare
	committees []*Committee
}

type validatorStore struct {
	operatorID func() spectypes.OperatorID
	shares     func() []*types.SSVShare
	byPubKey   func([]byte) *types.SSVShare

	byValidatorIndex map[phase0.ValidatorIndex]*types.SSVShare
	byCommitteeID    map[types.CommitteeID]*Committee
	byOperatorID     map[spectypes.OperatorID]*sharesAndCommittees

	mu sync.RWMutex
}

func newValidatorStore(
	operatorID func() spectypes.OperatorID,
	shares func() []*types.SSVShare,
	shareByPubKey func([]byte) *types.SSVShare,
) *validatorStore {
	return &validatorStore{
		operatorID:       operatorID,
		shares:           shares,
		byPubKey:         shareByPubKey,
		byValidatorIndex: make(map[phase0.ValidatorIndex]*types.SSVShare),
		byCommitteeID:    make(map[types.CommitteeID]*Committee),
		byOperatorID:     make(map[spectypes.OperatorID]*sharesAndCommittees),
	}
}

func (c *validatorStore) Validator(pubKey []byte) *types.SSVShare {
	return c.byPubKey(pubKey)
}

func (c *validatorStore) ValidatorByIndex(index phase0.ValidatorIndex) *types.SSVShare {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.byValidatorIndex[index]
}

func (c *validatorStore) Validators() []*types.SSVShare {
	return c.shares()
}

func (c *validatorStore) ParticipatingValidators(epoch phase0.Epoch) []*types.SSVShare {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var validators []*types.SSVShare
	for _, share := range c.shares() {
		if share.IsParticipating(epoch) {
			validators = append(validators, share)
		}
	}
	return validators
}

func (c *validatorStore) OperatorValidators(id spectypes.OperatorID) []*types.SSVShare {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if data, ok := c.byOperatorID[id]; ok {
		return data.shares
	}
	return nil
}

func (c *validatorStore) Committee(id types.CommitteeID) *Committee {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.byCommitteeID[id]
}

func (c *validatorStore) Committees() []*Committee {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return maps.Values(c.byCommitteeID)
}

func (c *validatorStore) ParticipatingCommittees(epoch phase0.Epoch) []*Committee {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var committees []*Committee
	for _, committee := range c.byCommitteeID {
		if committee.IsParticipating(epoch) {
			committees = append(committees, committee)
		}
	}
	return committees
}

func (c *validatorStore) OperatorCommittees(id spectypes.OperatorID) []*Committee {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if data, ok := c.byOperatorID[id]; ok {
		return data.committees
	}
	return nil
}

func (c *validatorStore) SelfValidators() []*types.SSVShare {
	return c.OperatorValidators(c.operatorID())
}

func (c *validatorStore) SelfParticipatingValidators(epoch phase0.Epoch) []*types.SSVShare {
	validators := c.OperatorValidators(c.operatorID())

	c.mu.RLock()
	defer c.mu.RUnlock()

	var participating []*types.SSVShare
	for _, validator := range validators {
		if validator.IsParticipating(epoch) {
			participating = append(participating, validator)
		}
	}
	return participating
}

func (c *validatorStore) SelfCommittees() []*Committee {
	return c.OperatorCommittees(c.operatorID())
}

func (c *validatorStore) SelfParticipatingCommittees(epoch phase0.Epoch) []*Committee {
	committees := c.OperatorCommittees(c.operatorID())

	c.mu.RLock()
	defer c.mu.RUnlock()

	var participating []*Committee
	for _, committee := range committees {
		if committee.IsParticipating(epoch) {
			participating = append(participating, committee)
		}
	}
	return participating
}

func (c *validatorStore) handleShareAdded(share *types.SSVShare) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update byValidatorIndex
	if share.HasBeaconMetadata() {
		c.byValidatorIndex[share.BeaconMetadata.Index] = share
	}

	// Update byCommitteeID
	committee := c.byCommitteeID[share.CommitteeID()]
	if committee == nil {
		committee = buildCommittee([]*types.SSVShare{share})
		c.byCommitteeID[committee.ID] = committee
	} else {
		committee = buildCommittee(append(committee.Validators, share))
	}
	c.byCommitteeID[committee.ID] = committee

	// Update byOperatorID
	for _, operator := range share.GetCommittee() {
		data := c.byOperatorID[operator.Signer]
		if data == nil {
			data = &sharesAndCommittees{
				shares:     []*types.SSVShare{share},
				committees: []*Committee{committee},
			}
		} else {
			data.shares = append(data.shares, share)
			data.committees = append(data.committees, committee)
		}
	}
}

func (c *validatorStore) handleShareRemoved(pk spectypes.ValidatorPK) {
	c.mu.Lock()
	defer c.mu.Unlock()

	share := c.byPubKey(pk[:])
	if share == nil {
		return
	}

	// Update byValidatorIndex
	delete(c.byValidatorIndex, share.BeaconMetadata.Index)

	// Update byCommitteeID
	committee := c.byCommitteeID[share.CommitteeID()]
	if committee == nil {
		return
	}
	validators := make([]*types.SSVShare, 0, len(committee.Validators)-1)
	for _, validator := range committee.Validators {
		if validator.ValidatorPubKey != pk {
			validators = append(validators, validator)
		}
	}
	if len(validators) == 0 {
		delete(c.byCommitteeID, committee.ID)
	} else {
		committee.Validators = validators
	}

	// Update byOperatorID
	for _, operator := range share.GetCommittee() {
		data := c.byOperatorID[operator.Signer]
		if data == nil {
			return
		}
		shares := make([]*types.SSVShare, 0, len(data.shares)-1)
		for _, s := range data.shares {
			if s.ValidatorPubKey != pk {
				shares = append(shares, s)
			}
		}
		if len(shares) == 0 {
			delete(c.byOperatorID, operator.Signer)
		} else {
			data.shares = shares
		}
	}
}

func (c *validatorStore) handleShareUpdated(share *types.SSVShare) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update byValidatorIndex
	if share.HasBeaconMetadata() {
		c.byValidatorIndex[share.BeaconMetadata.Index] = share
	}

	// Update byCommitteeID
	for _, committee := range c.byCommitteeID {
		for i, validator := range committee.Validators {
			if validator.ValidatorPubKey == share.ValidatorPubKey {
				committee.Validators[i] = share
				break
			}
		}
	}

	// Update byOperatorID
	for _, data := range c.byOperatorID {
		for i, s := range data.shares {
			if s.ValidatorPubKey == share.ValidatorPubKey {
				data.shares[i] = share
				break
			}
		}
	}
}

func buildCommittee(shares []*types.SSVShare) *Committee {
	committee := &Committee{
		ID:         shares[0].CommitteeID(),
		Operators:  make([]spectypes.OperatorID, 0, len(shares)),
		Validators: shares,
	}

	seenOperators := make(map[spectypes.OperatorID]struct{})

	for _, share := range shares {
		for _, shareMember := range share.GetCommittee() {
			seenOperators[shareMember.Signer] = struct{}{}
		}
	}

	committee.Operators = maps.Keys(seenOperators)
	slices.Sort(committee.Operators)

	return committee
}
