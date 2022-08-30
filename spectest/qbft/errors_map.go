package qbft

var errorsMap = map[string]string{
	"invalid prepare message: message data is different from proposed data":                                                                                                              "invalid prepare msg: prepare data != proposed data",
	"invalid round change message: round change msg allows 1 signer":                                                                                                                     "round change msg invalid: round change msg allows 1 signer",
	"invalid proposal message: proposal not justified: proposal value invalid: invalid value":                                                                                            "proposal invalid: proposal not justified: proposal value invalid: invalid value",
	"invalid proposal message: proposal not justified: change round has not quorum":                                                                                                      "proposal invalid: proposal not justified: change round has not quorum",
	"invalid round change message: roundChangeData invalid: round change prepared value invalid":                                                                                         "round change msg invalid: roundChangeData invalid: round change prepared value invalid",
	"invalid commit message: could not get msg commit data: could not decode commit data from message: invalid character '\\x01' looking for beginning of value":                         "commit msg invalid: could not get msg commit data: could not decode commit data from message: invalid character '\\x01' looking for beginning of value",
	"invalid proposal message: proposal is not valid with current state, proposal round 1, state round 1, proposal acceptance true":                                                      "proposal invalid: proposal is not valid with current state",
	"invalid proposal message: proposal is not valid with current state, proposal round 2, state round 1, proposal acceptance false":                                                     "proposal invalid: proposal is not valid with current state",
	"invalid proposal message: proposal is not valid with current state, proposal round 2, state round 5, proposal acceptance true":                                                      "proposal invalid: proposal is not valid with current state",
	"invalid round change message: message height is wrong":                                                                                                                              "round change msg invalid: round change Height is wrong",
	"invalid proposal message: proposal not justified: change round msg not valid: round change justification invalid: round is wrong":                                                   "proposal invalid: proposal not justified: change round msg not valid: round change justification invalid: msg round wrong",
	"invalid round change message: round change justification invalid: round is wrong":                                                                                                   "round change msg invalid: round change justification invalid: msg round wrong",
	"invalid round change message: no justifications quorum":                                                                                                                             "round change msg invalid: no justifications quorum",
	"invalid proposal message: proposal not justified: signed prepare not valid":                                                                                                         "proposal invalid: proposal not justified: signed prepare not valid",
	"invalid proposal message: message height is wrong":                                                                                                                                  "proposal invalid: proposal Height is wrong",
	"invalid proposal message: proposal not justified: change round msg not valid: no justifications quorum":                                                                             "proposal invalid: proposal not justified: change round msg not valid: no justifications quorum",
	"invalid round change message: round change justification invalid: invalid message signature: failed to verify signature":                                                            "round change msg invalid: round change justification invalid: prepare msg signature invalid: failed to verify signature",
	"invalid proposal message: proposal not justified: change round msg not valid: invalid message signature: failed to verify signature":                                                "proposal invalid: proposal not justified: change round msg not valid: round change msg signature invalid: failed to verify signature",
	"invalid proposal message: proposal not justified: proposed data doesn't match highest prepared":                                                                                     "proposal invalid: proposal not justified: proposed data doesn't match highest prepared",
	"round change msg invalid: could not get roundChange data : could not decode round change data from message: invalid character '\\x01' looking for beginning of value":               "invalid round change message: could not get roundChange data : could not decode round change data from message: invalid character '\\x01' looking for beginning of value",
	"invalid commit message: message data is different from proposed data":                                                                                                               "commit msg invalid: proposed data different than commit msg data",
	"invalid commit message: round is wrong":                                                                                                                                             "commit msg invalid: commit round is wrong",
	"invalid proposal message: could not get proposal data: could not decode proposal data from message: invalid character '\\x01' looking for beginning of value":                       "proposal invalid: could not get proposal data: could not decode proposal data from message: invalid character '\\x01' looking for beginning of value",
	"invalid prepare message: prepare msg allows 1 signer":                                                                                                                               "invalid prepare msg: prepare msg allows 1 signer",
	"invalid proposal message: proposal msg allows 1 signer":                                                                                                                             "proposal invalid: proposal msg allows 1 signer",
	"invalid proposal message: proposal not justified: change round msg not valid: round change justification invalid: prepare data != proposed data":                                    "proposal invalid: proposal not justified: change round msg not valid: round change justification invalid: prepare data != proposed data",
	"invalid prepare message: round is wrong":                                                                                                                                            "invalid prepare msg: msg round wrong",
	"invalid proposal message: proposal leader invalid":                                                                                                                                  "proposal invalid: proposal leader invalid",
	"invalid prepare message: invalid message signature: failed to verify signature":                                                                                                     "invalid prepare msg: prepare msg signature invalid: failed to verify signature",
	"invalid commit message: message height is wrong":                                                                                                                                    "commit msg invalid: commit Height is wrong",
	"invalid proposal message: invalid message signature: failed to verify signature":                                                                                                    "proposal invalid: proposal msg signature invalid: failed to verify signature",
	"invalid prepare message: could not get prepare data: could not decode prepare data from message: invalid character '\\x01' looking for beginning of value":                          "invalid prepare msg: could not get prepare data: could not decode prepare data from message: invalid character '\\x01' looking for beginning of value",
	"invalid round change message: round change justification invalid: prepare data != proposed data":                                                                                    "round change msg invalid: round change justification invalid: prepare data != proposed data",
	"invalid commit message: invalid message signature: failed to verify signature":                                                                                                      "commit msg invalid: commit msg signature invalid: failed to verify signature",
	"invalid round change message: round change justification invalid: prepareData invalid: PrepareData data is invalid":                                                                 "round change msg invalid: round change justification invalid: prepareData invalid: PrepareData data is invalid",
	"invalid round change message: invalid message signature: failed to verify signature":                                                                                                "round change msg invalid: round change msg signature invalid: failed to verify signature",
	"invalid round change message: round change justification invalid: prepare msg allows 1 signer":                                                                                      "round change msg invalid: round change justification invalid: prepare msg allows 1 signer",
	"invalid prepare message: message height is wrong":                                                                                                                                   "invalid prepare msg: msg Height wrong",
	"invalid round change message: round change justification invalid: change round justification round lower or equal to message round":                                                 "round change msg invalid: round change justification invalid: msg round wrong",
	"invalid proposal message: proposal not justified: change round msg not valid: round change justification invalid: change round justification round lower or equal to message round": "proposal invalid: proposal not justified: change round msg not valid: round change justification invalid: msg round wrong",
	"invalid round change message: round change justification invalid: change round prepared round not equal to justification msg round":                                                 "round change msg invalid: round change justification invalid: msg round wrong",
	"invalid round change message: could not get roundChange data : could not decode round change data from message: invalid character '\\x01' looking for beginning of value":           "round change msg invalid: could not get roundChange data : could not decode round change data from message: invalid character '\\x01' looking for beginning of value",
	"invalid commit message: did not receive proposal for this round":                                                                                                                    "did not receive proposal for this round",
	"invalid prepare message: no proposal accepted for prepare":                                                                                                                          "no proposal accepted for prepare",
}
