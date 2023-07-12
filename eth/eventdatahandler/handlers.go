package eventdatahandler

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/herumi/bls-eth-go-binary/bls"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/eth/contract"
	"github.com/bloxapp/ssv/eth/eventdb"
	"github.com/bloxapp/ssv/logging/fields"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	ssvtypes "github.com/bloxapp/ssv/protocol/v2/types"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/utils/rsaencryption"
)

// b64 encrypted key length is 256
const encryptedKeyLength = 256

var (
	ErrAlreadyRegistered            = fmt.Errorf("operator registered with the same operator public key")
	ErrOperatorDataNotFound         = fmt.Errorf("operator data not found")
	ErrIncorrectSharesLength        = fmt.Errorf("shares length is not correct")
	ErrSignatureVerification        = fmt.Errorf("signature cannot be verified")
	ErrShareBelongsToDifferentOwner = fmt.Errorf("share already exists and belongs to different owner")
	ErrValidatorShareNotFound       = fmt.Errorf("validator share not found")
)

func (edh *EventDataHandler) handleOperatorAdded(txn eventdb.RW, event *contract.ContractOperatorAdded) error {
	logger := edh.logger.With(
		fields.OperatorID(event.OperatorId),
		zap.String("operator_pub_key", ethcommon.Bytes2Hex(event.PublicKey)),
		zap.String("owner_address", event.Owner.String()),
		zap.String("event_type", OperatorAdded),
	)

	logger.Debug("processing event")

	od := &registrystorage.OperatorData{
		PublicKey:    event.PublicKey,
		OwnerAddress: event.Owner,
		ID:           event.OperatorId,
	}

	// throw an error if there is an existing operator with the same public key and different operator id
	if edh.operatorData.ID != 0 && bytes.Equal(edh.operatorData.PublicKey, event.PublicKey) && edh.operatorData.ID != event.OperatorId {
		logger.Warn("malformed event: operator registered with the same operator public key",
			zap.Uint64("expected_operator_id", edh.operatorData.ID))
		return &MalformedEventError{Err: ErrAlreadyRegistered}
	}

	// TODO: consider saving other operators as well
	exists, err := txn.SaveOperatorData(od)
	if err != nil {
		return fmt.Errorf("save operator data: %w", err)
	}
	if exists {
		logger.Debug("operator data already exists")
		return nil
	}

	if bytes.Equal(event.PublicKey, edh.operatorData.PublicKey) {
		edh.operatorData = od
	}

	edh.metrics.OperatorPublicKey(od.ID, od.PublicKey)
	logger.Debug("report operator public key",
		fields.OperatorID(od.ID),
		fields.PubKey(od.PublicKey))

	logger.Debug("processed event")
	return nil
}

func (edh *EventDataHandler) handleOperatorRemoved(txn eventdb.RW, event *contract.ContractOperatorRemoved) error {
	logger := edh.logger.With(
		fields.OperatorID(event.OperatorId),
		zap.String("event_type", OperatorRemoved),
	)
	logger.Debug("processing event")

	od, err := txn.GetOperatorData(event.OperatorId)
	if err != nil {
		return fmt.Errorf("could not get operator data: %w", err)
	}
	if od == nil {
		logger.Warn("malformed event: could not find operator data")
		return &MalformedEventError{Err: ErrOperatorDataNotFound}
	}

	logger = logger.With(
		zap.String("operator_pub_key", ethcommon.Bytes2Hex(od.PublicKey)),
		zap.String("owner_address", od.OwnerAddress.String()),
	)

	// TODO: In original handler we didn't delete operator data, so this behavior was preserved. However we likely need to.
	// TODO: Delete operator from all the shares.
	//	var shares []Share
	//	for _, s := range nodeStorage.Shares().List() {
	//		// if operator in committee, delete him from it:
	//		//     shares = aopend(shares, s)
	//	}
	//	nodeStorage.Shares().Save(shares)

	logger.Debug("processed event")
	return nil
}

func (edh *EventDataHandler) handleValidatorAdded(txn eventdb.RW, event *contract.ContractValidatorAdded) error {
	logger := edh.logger.With(
		zap.String("owner_address", event.Owner.String()),
		zap.Uint64s("operator_ids", event.OperatorIds),
		zap.String("operator_pub_key", ethcommon.Bytes2Hex(event.PublicKey)),
		zap.String("event_type", ValidatorAdded),
	)
	logger.Debug("processing event")

	// get nonce
	nonce, nonceErr := txn.GetNextNonce(event.Owner)
	if nonceErr != nil {
		return fmt.Errorf("failed to get next nonce: %w", nonceErr)
	}

	// Calculate the expected length of constructed shares based on the number of operator IDs,
	// signature length, public key length, and encrypted key length.
	operatorCount := len(event.OperatorIds)
	signatureOffset := phase0.SignatureLength
	pubKeysOffset := phase0.PublicKeyLength*operatorCount + signatureOffset
	sharesExpectedLength := encryptedKeyLength*operatorCount + pubKeysOffset

	if sharesExpectedLength != len(event.Shares) {
		logger.Warn("malformed event: event shares length is not correct",
			zap.Int("expected", sharesExpectedLength),
			zap.Int("got", len(event.Shares)))

		if err := txn.BumpNonce(event.Owner); err != nil {
			return fmt.Errorf("bump nonce: %w", err)
		}

		return &MalformedEventError{Err: ErrIncorrectSharesLength}
	}

	signature := event.Shares[:signatureOffset]
	sharePublicKeys := splitBytes(event.Shares[signatureOffset:pubKeysOffset], phase0.PublicKeyLength)
	encryptedKeys := splitBytes(event.Shares[pubKeysOffset:], len(event.Shares[pubKeysOffset:])/operatorCount)

	// verify sig
	if err := verifySignature(signature, event.Owner, event.PublicKey, nonce); err != nil {
		logger.Warn("malformed event: failed to verify signature",
			zap.Error(err))

		if err := txn.BumpNonce(event.Owner); err != nil {
			return fmt.Errorf("bump nonce: %w", err)
		}

		return &MalformedEventError{Err: ErrSignatureVerification}
	}

	validatorShare := edh.shareMap.Get(event.PublicKey)

	if validatorShare == nil {
		createdShare, err := edh.handleShareCreation(txn, event, sharePublicKeys, encryptedKeys)
		if err != nil {

			var malformedEventError *MalformedEventError
			if errors.As(err, &malformedEventError) {
				logger.Warn("malformed event", zap.Error(err))
				return err
			}

			edh.metrics.ValidatorError(event.PublicKey)
			return err
		}

		validatorShare = createdShare

		if err := txn.BumpNonce(event.Owner); err != nil {
			return fmt.Errorf("bump nonce: %w", err)
		}
	} else if event.Owner != validatorShare.OwnerAddress {
		// Prevent multiple registration of the same validator with different owner address
		// owner A registers validator with public key X (OK)
		// owner B registers validator with public key X (NOT OK)
		logger.Warn("malformed event: validator share already exists with different owner address",
			zap.String("expected", validatorShare.OwnerAddress.String()),
			zap.String("got", event.Owner.String()))

		if err := txn.BumpNonce(event.Owner); err != nil {
			return fmt.Errorf("bump nonce: %w", err)
		}

		return &MalformedEventError{Err: ErrShareBelongsToDifferentOwner}
	}

	isOperatorShare := validatorShare.BelongsToOperator(edh.operatorData.ID)
	if isOperatorShare {
		edh.metrics.ValidatorInactive(event.PublicKey)
	}

	logger.Debug("processed event")
	return nil
}

// handleShareCreation is called when a validator was added/updated during registry sync
func (edh *EventDataHandler) handleShareCreation(
	txn eventdb.RW,
	validatorEvent *contract.ContractValidatorAdded,
	sharePublicKeys [][]byte,
	encryptedKeys [][]byte,
) (*ssvtypes.SSVShare, error) {
	share, shareSecret, err := validatorAddedEventToShare(
		validatorEvent,
		edh.shareEncryptionKeyProvider,
		edh.operatorData,
		sharePublicKeys,
		encryptedKeys,
	)
	if err != nil {
		return nil, fmt.Errorf("could not extract validator share from event: %w", err)
	}

	if share.BelongsToOperator(edh.operatorData.ID) {
		if shareSecret == nil {
			return nil, errors.New("could not decode shareSecret")
		}

		logger := edh.logger.With(fields.PubKey(share.ValidatorPubKey))

		// get metadata
		if updated, err := updateShareMetadata(share, edh.beacon); err != nil {
			logger.Warn("could not add validator metadata", zap.Error(err))
		} else if !updated {
			logger.Warn("could not find validator metadata")
		}

		// save secret key
		if err := edh.keyManager.AddShare(shareSecret); err != nil {
			return nil, fmt.Errorf("could not add share secret to key manager: %w", err)
		}
	}

	edh.shareMap.Save(share)

	// save validator data
	if err := txn.SaveShares(share); err != nil {
		return nil, fmt.Errorf("could not save validator share: %w", err)
	}

	return share, nil
}

func updateShareMetadata(share *ssvtypes.SSVShare, bc beaconprotocol.BeaconNode) (bool, error) {
	pk := hex.EncodeToString(share.ValidatorPubKey)
	results, err := beaconprotocol.FetchValidatorsMetadata(bc, [][]byte{share.ValidatorPubKey})
	if err != nil {
		return false, fmt.Errorf("failed to fetch metadata for share: %w", err)
	}
	meta, ok := results[pk]
	if !ok {
		return false, nil
	}
	share.BeaconMetadata = meta
	return true, nil
}

func validatorAddedEventToShare(
	event *contract.ContractValidatorAdded,
	shareEncryptionKeyProvider ShareEncryptionKeyProvider,
	operatorData *registrystorage.OperatorData,
	sharePublicKeys [][]byte,
	encryptedKeys [][]byte,
) (*ssvtypes.SSVShare, *bls.SecretKey, error) {
	validatorShare := ssvtypes.SSVShare{}

	publicKey, err := ssvtypes.DeserializeBLSPublicKey(event.PublicKey)
	if err != nil {
		return nil, nil, &MalformedEventError{
			Err: fmt.Errorf("failed to deserialize validator public key: %w", err),
		}
	}
	validatorShare.ValidatorPubKey = publicKey.Serialize()
	validatorShare.OwnerAddress = event.Owner
	var shareSecret *bls.SecretKey

	committee := make([]*spectypes.Operator, 0)
	for i := range event.OperatorIds {
		operatorID := event.OperatorIds[i]
		committee = append(committee, &spectypes.Operator{
			OperatorID: operatorID,
			PubKey:     sharePublicKeys[i],
		})

		if operatorID != operatorData.ID {
			continue
		}

		validatorShare.OperatorID = operatorID
		validatorShare.SharePubKey = sharePublicKeys[i]

		operatorPrivateKey, found, err := shareEncryptionKeyProvider()
		if err != nil {
			return nil, nil, fmt.Errorf("could not get operator private key: %w", err)
		}
		if !found {
			return nil, nil, errors.New("could not find operator private key")
		}

		shareSecret = &bls.SecretKey{}
		decryptedSharePrivateKey, err := rsaencryption.DecodeKey(operatorPrivateKey, encryptedKeys[i])
		if err != nil {
			return nil, nil, &MalformedEventError{
				Err: fmt.Errorf("could not decrypt share private key: %w", err),
			}
		}
		if err = shareSecret.SetHexString(string(decryptedSharePrivateKey)); err != nil {
			return nil, nil, &MalformedEventError{
				Err: fmt.Errorf("could not set decrypted share private key: %w", err),
			}
		}
		if !bytes.Equal(shareSecret.GetPublicKey().Serialize(), validatorShare.SharePubKey) {
			return nil, nil, &MalformedEventError{
				Err: errors.New("share private key does not match public key"),
			}
		}
	}

	validatorShare.Quorum, validatorShare.PartialQuorum = ssvtypes.ComputeQuorumAndPartialQuorum(len(committee))
	validatorShare.DomainType = ssvtypes.GetDefaultDomain()
	validatorShare.Committee = committee
	validatorShare.Graffiti = []byte("ssv.network")

	return &validatorShare, shareSecret, nil
}

func (edh *EventDataHandler) handleValidatorRemoved(txn eventdb.RW, event *contract.ContractValidatorRemoved) error {
	logger := edh.logger.With(
		zap.String("owner_address", event.Owner.String()),
		zap.Uint64s("operator_ids", event.OperatorIds),
		fields.PubKey(event.PublicKey),
		zap.String("event_type", ValidatorRemoved),
	)
	logger.Debug("processing event")

	// TODO: handle metrics
	share := edh.shareMap.Get(event.PublicKey)
	if share == nil {
		logger.Warn("malformed event: could not find validator share")
		return &MalformedEventError{Err: ErrValidatorShareNotFound}
	}

	// Prevent removal of the validator registered with different owner address
	// owner A registers validator with public key X (OK)
	// owner B registers validator with public key X (NOT OK)
	// owner A removes validator with public key X (OK)
	// owner B removes validator with public key X (NOT OK)
	if event.Owner != share.OwnerAddress {
		logger.Warn("malformed event: validator share already exists with different owner address",
			zap.String("expected", share.OwnerAddress.String()),
			zap.String("got", event.Owner.String()))

		return &MalformedEventError{Err: ErrShareBelongsToDifferentOwner}
	}

	// remove decided messages
	messageID := spectypes.NewMsgID(ssvtypes.GetDefaultDomain(), share.ValidatorPubKey, spectypes.BNRoleAttester)
	store := edh.storageMap.Get(messageID.GetRoleType())
	if store != nil {
		if err := store.CleanAllInstances(logger, messageID[:]); err != nil { // TODO need to delete for multi duty as well
			return fmt.Errorf("could not clean all decided messages: %w", err)
		}
	}

	edh.shareMap.Delete(share.ValidatorPubKey)

	// remove from storage
	if err := txn.DeleteShare(share.ValidatorPubKey); err != nil {
		return fmt.Errorf("could not remove validator share: %w", err)
	}

	isOperatorShare := share.BelongsToOperator(edh.operatorData.ID)
	if isOperatorShare {
		edh.metrics.ValidatorRemoved(event.PublicKey)
	}

	if isOperatorShare || edh.fullNode {
		logger = logger.With(zap.String("validator_pubkey", hex.EncodeToString(share.ValidatorPubKey)))
	}

	logger.Debug("processed event")
	return nil
}

func (edh *EventDataHandler) handleClusterLiquidated(txn eventdb.RW, event *contract.ContractClusterLiquidated) ([]*ssvtypes.SSVShare, error) {
	logger := edh.logger.With(
		zap.String("owner_address", event.Owner.String()),
		zap.Uint64s("operator_ids", event.OperatorIds),
		zap.String("event_type", ClusterLiquidated),
	)
	logger.Debug("processing event")

	toLiquidate, liquidatedPubKeys, err := edh.processClusterEvent(txn, event.Owner, event.OperatorIds, true)
	if err != nil {
		return nil, fmt.Errorf("could not process cluster event: %w", err)
	}

	if len(liquidatedPubKeys) > 0 {
		logger = logger.With(zap.Strings("liquidated_validators", liquidatedPubKeys))
	}

	logger.Debug("processed event")
	return toLiquidate, nil
}

func (edh *EventDataHandler) handleClusterReactivated(txn eventdb.RW, event *contract.ContractClusterReactivated) ([]*ssvtypes.SSVShare, error) {
	logger := edh.logger.With(
		zap.String("owner_address", event.Owner.String()),
		zap.Uint64s("operator_ids", event.OperatorIds),
		zap.String("event_type", ClusterReactivated),
	)
	logger.Debug("processing event")

	toEnable, enabledPubKeys, err := edh.processClusterEvent(txn, event.Owner, event.OperatorIds, false)
	if err != nil {
		return nil, fmt.Errorf("could not process cluster event: %w", err)
	}

	if len(enabledPubKeys) > 0 {
		logger = logger.With(zap.Strings("enabled_validators", enabledPubKeys))
	}

	logger.Debug("processed event")
	return toEnable, nil
}

func (edh *EventDataHandler) handleFeeRecipientAddressUpdated(txn eventdb.RW, event *contract.ContractFeeRecipientAddressUpdated) (bool, error) {
	logger := edh.logger.With(
		zap.String("owner_address", event.Owner.String()),
		fields.FeeRecipient(event.RecipientAddress.Bytes()),
		zap.String("event_type", FeeRecipientAddressUpdated),
	)
	logger.Debug("processing event")

	recipientData := &registrystorage.RecipientData{
		Owner: event.Owner,
	}
	copy(recipientData.FeeRecipient[:], event.RecipientAddress.Bytes())

	r, err := txn.SaveRecipientData(recipientData)
	if err != nil {
		return false, fmt.Errorf("could not save recipient data: %w", err)
	}

	logger.Debug("processed event")
	return r != nil, nil
}

func splitBytes(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

// verify signature of the ValidatorAddedEvent shares data
// todo(align-contract-v0.3.1-rc.0): move to crypto package in ssv protocol?
func verifySignature(sig []byte, owner ethcommon.Address, pubKey []byte, nonce registrystorage.Nonce) error {
	data := fmt.Sprintf("%s:%d", owner.String(), nonce)
	hash := crypto.Keccak256([]byte(data))

	sign := &bls.Sign{}
	if err := sign.Deserialize(sig); err != nil {
		return fmt.Errorf("failed to deserialize signature: %w", err)
	}

	pk := &bls.PublicKey{}
	if err := pk.Deserialize(pubKey); err != nil {
		return fmt.Errorf("failed to deserialize public key: %w", err)
	}

	if res := sign.VerifyByte(pk, hash); !res {
		return errors.New("failed to verify signature")
	}

	return nil
}

// processClusterEvent handles registry contract event for cluster
func (edh *EventDataHandler) processClusterEvent(
	txn eventdb.RW,
	owner ethcommon.Address,
	operatorIDs []uint64,
	toLiquidate bool,
) ([]*ssvtypes.SSVShare, []string, error) {
	clusterID, err := ssvtypes.ComputeClusterIDHash(owner.Bytes(), operatorIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("could not compute share cluster id: %w", err)
	}

	shares := edh.shareMap.List(registrystorage.ByClusterID(clusterID))
	toUpdate := make([]*ssvtypes.SSVShare, 0)
	updatedPubKeys := make([]string, 0)

	for _, share := range shares {
		isOperatorShare := share.BelongsToOperator(edh.operatorData.ID)
		if isOperatorShare || edh.fullNode {
			updatedPubKeys = append(updatedPubKeys, hex.EncodeToString(share.ValidatorPubKey))
		}
		if isOperatorShare {
			share.Liquidated = toLiquidate
			toUpdate = append(toUpdate, share)
		}
	}

	if len(toUpdate) > 0 {
		edh.shareMap.Save(toUpdate...)
		if err = txn.SaveShares(toUpdate...); err != nil {
			return nil, nil, fmt.Errorf("could not save validator shares: %w", err)
		}
	}

	return toUpdate, updatedPubKeys, nil
}

// MalformedEventError is returned when event is malformed
type MalformedEventError struct {
	Err error
}

func (e *MalformedEventError) Error() string {
	return e.Err.Error()
}
