package logs_catcher

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/qbft"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv/protocol/v2/message"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/e2e/logs_catcher/docker"
)

const (
	targetContainer = "ssv-node-1"

	verifySignatureErr           = "failed processing consensus message: could not process msg: invalid signed message: msg signature invalid: failed to verify signature"
	reconstructSignatureErr      = "could not reconstruct post consensus signature: could not reconstruct beacon sig: failed to verify reconstruct signature: could not reconstruct a valid signature"
	pastRoundErr                 = "failed processing consensus message: could not process msg: invalid signed message: past round"
	reconstructSignaturesSuccess = "reconstructed partial signatures"
	submittedAttSuccess          = "✅ successfully submitted attestation"
	gotDutiesSuccess             = "🗂 got duties"

	msgHeightField        = "\"msg_height\":%d"
	msgRoundField         = "\"msg_round\":%d"
	msgTypeField          = "\"msg_type\":\"%s\""
	consensusMsgTypeField = "\"consensus_msg_type\":%d"
	signersField          = "\"signers\":[%d]"
	errorField            = "\"error\":\"%s\""
	dutyIDField           = "\"duty_id\":\"%s\""
	roleField             = "\"role\":\"%s\""
	slotField             = "\"slot\":%d"
)

type logCondition struct {
	role             string
	slot             phase0.Slot
	round            int
	msgType          types.MsgType
	consensusMsgType qbft.MessageType
	signer           types.OperatorID
	error            string
}

func VerifyBLSSignature(pctx context.Context, logger *zap.Logger, cli DockerCLI) error {
	startctx, startc := context.WithTimeout(pctx, time.Minute*6*4) // wait max 4 epochs
	defer startc()

	// TODO: pass corrupted operator & validator info from outside
	corruptedOperator := types.OperatorID(4)
	//corruptedValidatorIndex := fmt.Sprintf("v%d", 1476356) // leader 1
	//corruptedValidatorPubKey := "8c5801d7a18e27fae47dfdd99c0ac67fbc6a5a56bb1fc52d0309626d805861e04eaaf67948c18ad50c96d63e44328ab0"
	corruptedValidatorIndex := fmt.Sprintf("v%d", 1476359) // leader 4
	corruptedValidatorPubKey := "81bde622abeb6fb98be8e6d281944b11867c6ddb23b2af582b2af459a0316f766fdb97e56a6c69f66d85e411361c0b8a"

	conditionLog, err := StartCondition(startctx, logger, []string{gotDutiesSuccess, corruptedValidatorIndex}, targetContainer, cli)
	if err != nil {
		return fmt.Errorf("failed to start condition: %w", err)
	}

	dutyID, dutySlot, err := ParseAndExtractDutyInfo(conditionLog, corruptedValidatorIndex)
	if err != nil {
		return fmt.Errorf("failed to parse and extract duty info: %w", err)
	}
	fmt.Println("Duty ID: ", dutyID)

	leader, committee := DetermineLeaderAndCommittee(dutySlot)
	fmt.Println("Leader: ", leader)

	_, err = StartCondition(startctx, logger, []string{submittedAttSuccess, corruptedValidatorPubKey}, targetContainer, cli)
	if err != nil {
		return fmt.Errorf("failed to start condition: %w", err)
	}

	ctx, c := context.WithCancel(pctx)
	defer c()

	return ProcessLogs(ctx, logger, cli, committee, leader, dutyID, dutySlot, corruptedOperator)
}

func ParseAndExtractDutyInfo(conditionLog string, corruptedValidatorIndex string) (string, phase0.Slot, error) {
	parsedData, err := parseLogString(conditionLog)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse log string: %w", err)
	}

	dutyID, err := extractDutyID(parsedData, corruptedValidatorIndex)
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract duty id: %w", err)
	}

	dutySlot, err := extractDutySlot(dutyID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract duty slot: %w", err)
	}

	return dutyID, dutySlot, nil
}

func DetermineLeaderAndCommittee(dutySlot phase0.Slot) (types.OperatorID, []*types.Operator) {
	committee := []*types.Operator{
		{OperatorID: 1},
		{OperatorID: 2},
		{OperatorID: 3},
		{OperatorID: 4},
	}

	leader := qbft.RoundRobinProposer(&qbft.State{
		Share: &types.Share{
			Committee: committee,
		},
		Height: qbft.Height(dutySlot),
	}, qbft.FirstRound)

	return leader, committee
}

func ProcessLogs(ctx context.Context, logger *zap.Logger, cli DockerCLI, committee []*types.Operator, leader types.OperatorID, dutyID string, dutySlot phase0.Slot, corruptedOperator types.OperatorID) error {
	for _, operator := range committee {
		target := fmt.Sprintf("ssv-node-%d", operator.OperatorID)
		if operator.OperatorID == corruptedOperator {
			err := processCorruptedOperatorLogs(ctx, logger, cli, dutyID, dutySlot, corruptedOperator, target)
			if err != nil {
				return err
			}
		} else {
			err := processNonCorruptedOperatorLogs(ctx, logger, cli, leader, dutySlot, corruptedOperator, target)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func processCorruptedOperatorLogs(ctx context.Context, logger *zap.Logger, cli DockerCLI, dutyID string, dutySlot phase0.Slot, corruptedOperator types.OperatorID, target string) error {
	successConditions := []string{
		reconstructSignaturesSuccess,
		fmt.Sprintf(dutyIDField, dutyID),
	}
	failConditions := []string{
		fmt.Sprintf(roleField, types.BNRoleAttester.String()),
		fmt.Sprintf(slotField, dutySlot),
		fmt.Sprintf(msgTypeField, message.MsgTypeToString(types.SSVPartialSignatureMsgType)),
		fmt.Sprintf(errorField, reconstructSignatureErr),
	}
	return matchDualConditionLog(ctx, logger, cli, corruptedOperator, successConditions, failConditions, target)
}

func processNonCorruptedOperatorLogs(ctx context.Context, logger *zap.Logger, cli DockerCLI, leader types.OperatorID, dutySlot phase0.Slot, corruptedOperator types.OperatorID, target string) error {
	var conditions []logCondition
	if leader == corruptedOperator {
		conditions = []logCondition{
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            1,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.ProposalMsgType,
				signer:           corruptedOperator,
				error:            verifySignatureErr,
			},
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            1,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.PrepareMsgType,
				signer:           corruptedOperator,
				error:            pastRoundErr,
			},
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            2,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.RoundChangeMsgType,
				signer:           corruptedOperator,
				error:            verifySignatureErr,
			},
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            2,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.PrepareMsgType,
				signer:           corruptedOperator,
				error:            verifySignatureErr,
			},
			// TODO: should we handle decided failed signature?
		}
	} else {
		conditions = []logCondition{
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            1,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.PrepareMsgType,
				signer:           corruptedOperator,
				error:            verifySignatureErr,
			},
			{
				role:             types.BNRoleAttester.String(),
				slot:             dutySlot,
				round:            1,
				msgType:          types.SSVConsensusMsgType,
				consensusMsgType: qbft.CommitMsgType,
				signer:           corruptedOperator,
				error:            verifySignatureErr,
			},

			// TODO: should we handle decided failed signature?
			//if err := matchSingleConditionLog(ctx, logger, cli, []string{
			//	fmt.Sprintf(msgHeightField, dutySlot),
			//	fmt.Sprintf(msgTypeField, message.MsgTypeToString(types.SSVConsensusMsgType)),
			//	"\"consensus_msg_type\":2", // decided
			//	"\"signers\":[1,3,4]",
			//	"\"error\":\"failed processing consensus message: invalid decided msg: invalid decided msg: msg signature invalid: failed to verify signature\"",
			//}, target); err != nil {
			//	return err
			//}
			//
			//// TODO: 1,2,4 signers ??? no log for this
			//// TODO: 2,3,4 signers ??? no log for this
			//
			//if err := matchSingleConditionLog(ctx, logger, cli, []string{
			//	fmt.Sprintf(msgHeightField, dutySlot),
			//	fmt.Sprintf(msgTypeField, message.MsgTypeToString(types.SSVConsensusMsgType)),
			//	"\"consensus_msg_type\":2", // decided
			//	"\"signers\":[1,2,3,4]",
			//	"\"error\":\"failed processing consensus message: invalid decided msg: invalid decided msg: msg signature invalid: failed to verify signature\"",
			//}, target); err != nil {
			//	return err
			//}
			//
			//// post consensus
			//if err := matchSingleConditionLog(ctx, logger, cli, []string{
			//	fmt.Sprintf(msgHeightField, dutySlot),
			//	"\"msg_type\":\"partial_signature\"",
			//	"\"signer\":4",
			//	"\"error\":\"failed processing post consensus message: invalid post-consensus message: failed to verify PartialSignature: failed to verify signature\"",
			//}, target); err != nil {
			//	return err
			//}
		}
	}

	for _, condition := range conditions {
		if err := matchCondition(ctx, logger, cli, condition, target); err != nil {
			return fmt.Errorf("failed to match condition: %w", err)
		}
	}
	return nil
}

func matchCondition(ctx context.Context, logger *zap.Logger, cli DockerCLI, condition logCondition, target string) error {
	conditionStrings := []string{
		fmt.Sprintf(roleField, condition.role),
		fmt.Sprintf(msgHeightField, condition.slot),
		fmt.Sprintf(msgRoundField, condition.round),
		fmt.Sprintf(msgTypeField, message.MsgTypeToString(condition.msgType)),
		fmt.Sprintf(consensusMsgTypeField, condition.consensusMsgType),
		fmt.Sprintf(signersField, condition.signer),
		fmt.Sprintf(errorField, condition.error),
	}
	return matchSingleConditionLog(ctx, logger, cli, conditionStrings, target)
}

func matchSingleConditionLog(ctx context.Context, logger *zap.Logger, cli DockerCLI, first []string, target string) error {
	res, err := docker.DockerLogs(ctx, cli, target, "")
	if err != nil {
		return err
	}

	filteredLogs := res.Grep(first)

	logger.Info("matched", zap.Int("count", len(filteredLogs)), zap.String("target", target), zap.Strings("match_string", first))

	if len(filteredLogs) != 1 {
		return fmt.Errorf("found non matching messages on %v, want %v got %v", target, 1, len(filteredLogs))
	}

	logger.Info("SUCCESS matched ", zap.Int("matched", len(filteredLogs)))

	return nil
}

func matchDualConditionLog(ctx context.Context, logger *zap.Logger, cli DockerCLI, corruptedOperator types.OperatorID, success []string, fail []string, target string) error {
	res, err := docker.DockerLogs(ctx, cli, target, "")
	if err != nil {
		return err
	}

	filteredLogs := res.Grep(success)

	if len(filteredLogs) == 1 {
		logger.Info("matched", zap.Int("count", len(filteredLogs)), zap.String("target", target), zap.Strings("match_string", success), zap.String("RAW", filteredLogs[0]))
		parsedData, err := parseLogString(filteredLogs[0])
		if err != nil {
			return fmt.Errorf("error parsing log string: %v", err)
		}

		signers, err := extractSigners(parsedData)
		if err != nil {
			return fmt.Errorf("error extracting signers: %v", err)
		}

		for _, signer := range signers {
			if signer == corruptedOperator {
				return fmt.Errorf("found corrupted signer %v on successful signers %v", corruptedOperator, signers)
			}
		}
	} else {
		filteredLogs = res.Grep(fail)
		logger.Info("matched", zap.Int("count", len(filteredLogs)), zap.String("target", target), zap.Strings("match_string", fail))

		if len(filteredLogs) != 1 {
			return fmt.Errorf("found non matching messages on %v, want %v got %v", target, 1, len(filteredLogs))
		}
	}

	logger.Info("SUCCESS matched ", zap.Int("matched", len(filteredLogs)))
	return nil
}

func cleanLogString(logString string) string {
	// Regular expression to match ANSI color codes
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanedLog := ansiRegex.ReplaceAllString(logString, "")

	// Remove non-JSON characters before the first '{'
	startIndex := strings.Index(cleanedLog, "{")
	if startIndex > -1 {
		return cleanedLog[startIndex:]
	}
	return cleanedLog
}

func parseLogString(logString string) (map[string]interface{}, error) {
	cleanedLog := cleanLogString(logString)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedLog), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling log string: %v", err)
	}
	return result, nil
}

func extractDutyID(parsedData map[string]interface{}, searchPart string) (string, error) {
	if duties, ok := parsedData["duties"].(string); ok {
		dutyList := strings.Split(duties, ", ")
		for _, duty := range dutyList {
			if strings.Contains(duty, searchPart) {
				return duty, nil
			}
		}
	}
	return "", fmt.Errorf("no duty id found for %v", searchPart)
}

func extractDutySlot(dutyID string) (phase0.Slot, error) {
	// Extracting the part after "s" and before the next "-"
	parts := strings.Split(dutyID, "-")
	for _, part := range parts {
		if strings.HasPrefix(part, "s") {
			slotStr := strings.TrimPrefix(part, "s")
			slotInt, err := strconv.Atoi(slotStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse duty slot to int: %w", err)
			}
			return phase0.Slot(slotInt), nil
		}
	}
	return 0, fmt.Errorf("no duty slot found for %v", dutyID)
}

func extractSigners(parsedData map[string]interface{}) ([]types.OperatorID, error) {
	if signers, ok := parsedData["signers"].([]interface{}); ok {
		signerIDs := make([]types.OperatorID, len(signers))
		for i, signer := range signers {
			if signerNum, ok := signer.(float64); ok { // JSON numbers are parsed as float64
				signerIDs[i] = types.OperatorID(signerNum)
			} else {
				return nil, fmt.Errorf("failed to parse signer to int: %v", signer)
			}
		}
		return signerIDs, nil
	}
	return nil, fmt.Errorf("no signers found")
}
