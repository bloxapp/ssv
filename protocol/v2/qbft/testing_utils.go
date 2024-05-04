package qbft

// TODO: uncomment when fixing tests
//
//var TestingMessage = &specqbft.Message{
//	MsgType:    specqbft.ProposalMsgType,
//	Height:     specqbft.FirstHeight,
//	Round:      specqbft.FirstRound,
//	Identifier: []byte{1, 2, 3, 4},
//	Root:       [32]byte{1, 2, 3, 4},
//}
//
//var TestingSignedMsg = func() *spectypes.SignedSSVMessage {
//	return SignMsg(TestingSK, 1, TestingMessage)
//}()
//
//var SignMsg = func(sk *bls.SecretKey, id types.OperatorID, msg *specqbft.Message) *spectypes.SignedSSVMessage {
//	domain := testingutils.TestingSSVDomainType
//	sigType := types.QBFTSignatureType
//
//	r, _ := types.ComputeSigningRoot(msg, types.ComputeSignatureDomain(domain, sigType))
//	sig := sk.SignByte(r[:])
//
//	return &spectypes.SignedSSVMessage{
//		Message:   *msg,
//		Signers:   []types.OperatorID{id},
//		Signature: sig.Serialize(),
//	}
//}
//
//var TestingSK = func() *bls.SecretKey {
//	types.InitBLS()
//	ret := &bls.SecretKey{}
//	ret.SetByCSPRNG()
//	return ret
//}()
//
//var testingValidatorPK = phase0.BLSPubKey{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
//
//var testingShare = &types.Share{
//	OperatorID:      1,
//	ValidatorPubKey: testingValidatorPK[:],
//	SharePubKey:     TestingSK.GetPublicKey().Serialize(),
//	DomainType:      testingutils.TestingSSVDomainType,
//	Quorum:          3,
//	PartialQuorum:   2,
//	Committee: []*types.Operator{
//		{
//			OperatorID: 1,
//			PubKey:     TestingSK.GetPublicKey().Serialize(),
//		},
//	},
//}
//
//var TestingInstanceStruct = &specqbft.Instance{
//	State: &specqbft.State{
//		Share:                           testingShare,
//		ID:                              []byte{1, 2, 3, 4},
//		Round:                           1,
//		Height:                          1,
//		LastPreparedRound:               1,
//		LastPreparedValue:               []byte{1, 2, 3, 4},
//		ProposalAcceptedForCurrentRound: TestingSignedMsg,
//		Decided:                         false,
//		DecidedValue:                    []byte{1, 2, 3, 4},
//
//		ProposeContainer: &specqbft.MsgContainer{
//			Msgs: map[specqbft.Round][]*spectypes.SignedSSVMessage{
//				1: {
//					TestingSignedMsg,
//				},
//			},
//		},
//		PrepareContainer: &specqbft.MsgContainer{
//			Msgs: map[specqbft.Round][]*spectypes.SignedSSVMessage{
//				1: {
//					TestingSignedMsg,
//				},
//			},
//		},
//		CommitContainer: &specqbft.MsgContainer{
//			Msgs: map[specqbft.Round][]*spectypes.SignedSSVMessage{
//				1: {
//					TestingSignedMsg,
//				},
//			},
//		},
//		RoundChangeContainer: &specqbft.MsgContainer{
//			Msgs: map[specqbft.Round][]*spectypes.SignedSSVMessage{
//				1: {
//					TestingSignedMsg,
//				},
//			},
//		},
//	},
//}
//
//var TestingControllerStruct = &specqbft.Controller{
//	Identifier:      []byte{1, 2, 3, 4},
//	Height:          specqbft.Height(1),
//	Share:           testingShare,
//	StoredInstances: specqbft.InstanceContainer{TestingInstanceStruct},
//}
