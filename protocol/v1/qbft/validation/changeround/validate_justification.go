package changeround

import (
	"bytes"
	"fmt"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"

	"github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/signedmsg"
	"github.com/bloxapp/ssv/protocol/v1/types"
)

// validateJustification validates change round justifications
type validateJustification struct {
	share *beacon.Share
	round *specqbft.Round
}

// Validate is the constructor of validateJustification
func Validate(share *beacon.Share) pipelines.SignedMessagePipeline {
	return &validateJustification{
		share: share,
	}
}

// ValidateWithRound is the constructor of validateJustification with round
func ValidateWithRound(share *beacon.Share, round specqbft.Round) pipelines.SignedMessagePipeline {
	return &validateJustification{
		share: share,
		round: &round,
	}
}

// Run implements pipeline.Pipeline interface
func (p *validateJustification) Run(signedMessage *specqbft.SignedMessage) error {
	if signedMessage.Message.Data == nil {
		return errors.New("change round justification msg is nil")
	}

	if len(signedMessage.GetSigners()) != 1 {
		return errors.New("round change msg allows 1 signer")
	}

	// TODO - change to normal prepare pipeline
	data, err := signedMessage.Message.GetRoundChangeData()
	if err != nil {
		return fmt.Errorf("could not get roundChange data : %w", err) // TODO(nkryuchkov): remove whitespace in ssv-spec
	}
	if data == nil {
		return errors.New("change round data is nil")
	}
	if err := data.Validate(); err != nil {
		return fmt.Errorf("roundChangeData invalid: %w", err)
	}

	// Addition to formal spec
	// We add this extra tests on the msg itself to filter round change msgs with invalid justifications, before they are inserted into msg containers
	if !data.Prepared() {
		return nil
	}

	// validate prepare message justifications
	for _, rcj := range data.RoundChangeJustification {
		if err := p.validateRoundChangeJustification(rcj, data, signedMessage); err != nil {
			return fmt.Errorf("round change justification invalid: %w", err)
		}
	}

	if quorum, _, _ := signedmsg.HasQuorum(p.share, data.RoundChangeJustification); !quorum {
		return fmt.Errorf("no justifications quorum")
	}

	round := signedMessage.Message.Round
	if p.round != nil {
		round = *p.round
	}

	if data.PreparedRound > round {
		return errors.New("prepared round > round")
	}

	return nil
}

func (p *validateJustification) validateRoundChangeJustification(rcj *specqbft.SignedMessage, roundChangeData *specqbft.RoundChangeData, signedMessage *specqbft.SignedMessage) error {
	if rcj.Message == nil {
		return errors.New("change round justification msg is nil")
	}
	if rcj.Message.MsgType != specqbft.PrepareMsgType {
		return errors.Errorf("change round justification msg type not Prepare (%d)", rcj.Message.MsgType)
	}
	if signedMessage.Message.Height != rcj.Message.Height {
		return errors.New("change round justification sequence is wrong")
	}
	if signedMessage.Message.Round <= rcj.Message.Round {
		return errors.New("change round justification round lower or equal to message round")
	}
	if roundChangeData.PreparedRound != rcj.Message.Round {
		return errors.New("change round prepared round not equal to justification msg round")
	}
	if !bytes.Equal(signedMessage.Message.Identifier, rcj.Message.Identifier) {
		return errors.New("change round justification msg identifier not equal to msg identifier not equal to instance identifier")
	}
	prepareMsg, err := rcj.Message.GetPrepareData()
	if err != nil {
		return errors.Wrap(err, "could not get prepare data")
	}
	if err := prepareMsg.Validate(); err != nil {
		return fmt.Errorf("prepareData invalid: %w", err)
	}
	if !bytes.Equal(prepareMsg.Data, roundChangeData.PreparedValue) {
		return errors.New("prepare data != proposed data")
	}
	if len(rcj.GetSigners()) != 1 {
		return errors.New("prepare msg allows 1 signer")
	}

	// validateJustification justification signature
	pksMap, err := p.share.PubKeysByID(rcj.GetSigners())
	var pks beacon.PubKeys
	for _, v := range pksMap {
		pks = append(pks, v)
	}

	if err != nil {
		return errors.Wrap(err, "change round could not get pubkey")
	}
	aggregated := pks.Aggregate()

	if err = rcj.Signature.Verify(rcj, types.GetDefaultDomain(), spectypes.QBFTSignatureType, aggregated.Serialize()); err != nil {
		return errors.Wrap(err, "invalid message signature")
	}
	return nil
}

// Name implements pipeline.Pipeline interface
func (p *validateJustification) Name() string {
	return "validateJustification msg"
}
