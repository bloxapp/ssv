package sameduty

import (
	"fmt"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"

	"github.com/bloxapp/ssv/ekm"
)

type Wrapper struct {
	signer         spectypes.BeaconSigner
	sharePublicKey []byte
}

func New(signer spectypes.BeaconSigner, sharePublicKey []byte) *Wrapper {
	return &Wrapper{
		signer:         signer,
		sharePublicKey: sharePublicKey,
	}
}

func (saf *Wrapper) AttesterValueCheck(valueCheckF specqbft.ProposedValueCheckF) specqbft.ProposedValueCheckF {
	return func(data []byte) error {
		if valueCheckErr := valueCheckF(data); valueCheckErr != nil {
			if !strings.Contains(valueCheckErr.Error(), "slashable") {
				return valueCheckErr
			}

			sp, ok := saf.signer.(ekm.StorageProvider)
			if !ok {
				return valueCheckErr
			}

			highest, ok, err := sp.RetrieveHighestAttestation(saf.sharePublicKey)
			if err != nil || !ok {
				return valueCheckErr
			}

			cd := &spectypes.ConsensusData{}
			if err := cd.Decode(data); err != nil {
				return valueCheckErr
			}

			attestationData, err := cd.GetAttestationData()
			if err != nil {
				return valueCheckErr
			}

			if !equalAttestationData(attestationData, highest) {
				return valueCheckErr
			}

			return nil
		}

		return nil
	}
}

func (saf *Wrapper) ProposerValueCheck(valueCheckF specqbft.ProposedValueCheckF) specqbft.ProposedValueCheckF {
	return func(data []byte) error {
		if valueCheckErr := valueCheckF(data); valueCheckErr != nil {
			if !strings.Contains(valueCheckErr.Error(), "slashable") {
				return valueCheckErr
			}

			sp, ok := saf.signer.(ekm.StorageProvider)
			if !ok {
				return valueCheckErr
			}

			highest, ok, err := sp.RetrieveHighestProposal(saf.sharePublicKey)
			if err != nil || !ok {
				return valueCheckErr
			}

			cd := &spectypes.ConsensusData{}
			if err := cd.Decode(data); err != nil {
				return valueCheckErr
			}

			slot, err := getBlockSlot(cd)
			if err != nil {
				return valueCheckErr
			}

			if slot != highest {
				return valueCheckErr
			}

			return nil
		}

		return nil
	}
}

func equalAttestationData(a, b *phase0.AttestationData) bool {
	return a != nil && b != nil &&
		a.Slot == b.Slot &&
		a.Index == b.Index &&
		a.BeaconBlockRoot == b.BeaconBlockRoot &&
		equalCheckpoint(a.Source, b.Source) &&
		equalCheckpoint(a.Target, b.Target)
}

func equalCheckpoint(a, b *phase0.Checkpoint) bool {
	return a != nil && b != nil &&
		a.Epoch == b.Epoch &&
		a.Root == b.Root
}

func getBlockSlot(cd *spectypes.ConsensusData) (phase0.Slot, error) {
	blindedBlockData, _, err := cd.GetBlindedBlockData()
	if err != nil {
		blockData, _, err := cd.GetBlockData()
		if err != nil {
			return 0, fmt.Errorf("no block data")
		}

		slot, err := blockData.Slot()
		if err != nil {
			return 0, fmt.Errorf("get slot from block data: %w", err)
		}

		return slot, nil
	}

	slot, err := blindedBlockData.Slot()
	if err != nil {
		return 0, fmt.Errorf("get slot from blinded block data: %w", err)
	}

	return slot, nil
}
