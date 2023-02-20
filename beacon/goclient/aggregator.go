package goclient

import (
	"encoding/binary"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

// SubmitAggregateSelectionProof returns an AggregateAndProof object
func (gc *goClient) SubmitAggregateSelectionProof(slot phase0.Slot, committeeIndex phase0.CommitteeIndex, committeeLength uint64, validatorIndex phase0.ValidatorIndex, slotSig []byte) (*phase0.AggregateAndProof, error) {
	// As specified in spec, an aggregator should wait until two thirds of the way through slot
	// to broadcast the best aggregate to the global aggregate channel.
	// https://github.com/ethereum/consensus-specs/blob/v0.9.3/specs/validator/0_beacon-chain-validator.md#broadcast-aggregate
	gc.waitToSlotTwoThirds(slot)

	// differ from spec because we need to subscribe to subnet
	isAggregator, err := isAggregator(committeeLength, slotSig)
	if err != nil {
		return nil, errors.Wrap(err, "could not get aggregator status")
	}
	if !isAggregator {
		return nil, errors.New("validator is not an aggregator")
	}

	attDataReqStart := time.Now()
	data, err := gc.client.AttestationData(gc.ctx, slot, committeeIndex)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("attestation data is nil")
	}

	metricsAttesterDataRequest.Observe(time.Since(attDataReqStart).Seconds())

	// Get aggregate attestation data.
	root, err := data.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "AttestationData.HashTreeRoot")
	}

	aggDataReqStart := time.Now()
	aggregateData, err := gc.client.AggregateAttestation(gc.ctx, slot, root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aggregate attestation")
	}
	if aggregateData == nil {
		return nil, errors.New("aggregation data is nil")
	}

	metricsAggregatorDataRequest.Observe(time.Since(aggDataReqStart).Seconds())

	var selectionProof phase0.BLSSignature
	copy(selectionProof[:], slotSig)

	return &phase0.AggregateAndProof{
		AggregatorIndex: validatorIndex,
		Aggregate:       aggregateData,
		SelectionProof:  selectionProof,
	}, nil
}

// SubmitSignedAggregateSelectionProof broadcasts a signed aggregator msg
func (gc *goClient) SubmitSignedAggregateSelectionProof(msg *phase0.SignedAggregateAndProof) error {
	return gc.client.SubmitAggregateAttestations(gc.ctx, []*phase0.SignedAggregateAndProof{msg})
}

// IsAggregator returns true if the signature is from the input validator. The committee
// count is provided as an argument rather than imported implementation from spec. Having
// committee count as an argument allows cheaper computation at run time.
//
// Spec pseudocode definition:
//
//	def is_aggregator(state: BeaconState, slot: Slot, index: CommitteeIndex, slot_signature: BLSSignature) -> bool:
//	 committee = get_beacon_committee(state, slot, index)
//	 modulo = max(1, len(committee) // TARGET_AGGREGATORS_PER_COMMITTEE)
//	 return bytes_to_uint64(hash(slot_signature)[0:8]) % modulo == 0
func isAggregator(committeeCount uint64, slotSig []byte) (bool, error) {
	modulo := committeeCount / TargetAggregatorsPerCommittee
	if modulo == 0 {
		// Modulo must be at least 1.
		modulo = 1
	}

	b := Hash(slotSig)
	return binary.LittleEndian.Uint64(b[:8])%modulo == 0, nil
}

// waitOneThirdOrValidBlock waits until one-third of the slot has transpired (SECONDS_PER_SLOT / 3 seconds after the start of slot)
func (gc *goClient) waitToSlotTwoThirds(slot phase0.Slot) {
	oneThird := gc.network.SlotDurationSec() / 3 /* one third of slot duration */

	finalTime := gc.slotStartTime(slot).Add(2 * oneThird)
	wait := time.Until(finalTime)
	if wait <= 0 {
		return
	}
	time.Sleep(wait)
}
