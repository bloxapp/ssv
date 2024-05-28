package dutystore

import (
	eth2apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"sync/atomic"
)

type Store struct {
	ready         atomic.Bool
	Attester      *Duties[eth2apiv1.AttesterDuty]
	Proposer      *Duties[eth2apiv1.ProposerDuty]
	SyncCommittee *SyncCommitteeDuties
}

func New() *Store {
	return &Store{
		Attester:      NewDuties[eth2apiv1.AttesterDuty](),
		Proposer:      NewDuties[eth2apiv1.ProposerDuty](),
		SyncCommittee: NewSyncCommitteeDuties(),
	}
}
