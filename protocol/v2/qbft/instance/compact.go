package instance

import (
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
)

// Compact trims the given qbft.State down to the minimum required
// for consensus to proceed.
//
// Compact always discards message from previous rounds.
// Compact discards non-commit messages from below their current round, only if the state is decided.
// Compact discards commit messages from below the current round, only if the whole committee has signed.
//
// This helps reduce the state's memory footprint.
func Compact(state *specqbft.State, decidedMessage *specqbft.SignedMessage) {
	compact(state, decidedMessage, compactContainerEdit)
}

// CompactCopy returns a compacted copy of the given qbft.State.
// CompactCopy is guaranteed to not modify the given state, but the returned state may be modified
// when the given state is modified.
//
// TODO: this is a temporary solution to not break spec-tests. Revert this once spec is aligned.
//
// See Compact for more details.
func CompactCopy(state *specqbft.State, decidedMessage *specqbft.SignedMessage) *specqbft.State {
	copy := *state
	compact(&copy, decidedMessage, compactContainerCopy)
	return &copy
}

func compact(state *specqbft.State, decidedMessage *specqbft.SignedMessage, compactContainer compactContainerFunc) {
	state.ProposeContainer = compactContainer(state.ProposeContainer, state.Round, state.Decided)
	state.PrepareContainer = compactContainer(state.PrepareContainer, state.LastPreparedRound, state.Decided)
	state.RoundChangeContainer = compactContainer(state.RoundChangeContainer, state.Round, state.Decided)

	// Only discard commit messages if the whole committee has signed,
	// otherwise just trim down to the current round and future rounds.
	var wholeCommitteeDecided bool
	if state.Share == nil {
		// Share may be missing in tests.
	} else {
		var signers []spectypes.OperatorID
		if decidedMessage != nil {
			signers = decidedMessage.Signers
		} else if state.Decided && len(state.CommitContainer.Msgs) >= len(state.Share.Committee) {
			// TODO: this is a temporary solution, we shouldn't hash here probably.
			root, err := specqbft.HashDataRoot(state.DecidedValue)
			if err != nil {
				panic(err)
			}
			signers, _ = state.CommitContainer.LongestUniqueSignersForRoundAndRoot(state.Round, root)
		}
		wholeCommitteeDecided = len(signers) == len(state.Share.Committee)
	}
	state.CommitContainer = compactContainer(state.CommitContainer, state.Round, wholeCommitteeDecided)
}

type compactContainerFunc func(container *specqbft.MsgContainer, currentRound specqbft.Round, clear bool) *specqbft.MsgContainer

func compactContainerEdit(container *specqbft.MsgContainer, currentRound specqbft.Round, clear bool) *specqbft.MsgContainer {
	switch {
	case container == nil || len(container.Msgs) == 0:
		// Empty already.
	case clear:
		// Discard all messages.
		container.Msgs = map[specqbft.Round][]*specqbft.SignedMessage{}
	default:
		// Trim down to the current and future rounds.
		for r := range container.Msgs {
			if r < currentRound {
				delete(container.Msgs, r)
			}
		}
	}
	return container
}

func compactContainerCopy(container *specqbft.MsgContainer, currentRound specqbft.Round, clear bool) *specqbft.MsgContainer {
	switch {
	case container == nil || len(container.Msgs) == 0:
		// Empty already.
		return container
	case clear:
		// Discard all messages.
		return &specqbft.MsgContainer{
			Msgs: map[specqbft.Round][]*specqbft.SignedMessage{},
		}
	default:
		// Trim down to the current and future rounds.
		compact := specqbft.MsgContainer{
			Msgs: map[specqbft.Round][]*specqbft.SignedMessage{},
		}
		for r, msgs := range container.Msgs {
			if r >= currentRound {
				compact.Msgs[r] = msgs
			}
		}
		return &compact
	}
}
