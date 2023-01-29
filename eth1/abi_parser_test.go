package eth1

import (
	"encoding/hex"
	"strings"
	"testing"

	json "github.com/bytedance/sonic"

	"github.com/bloxapp/ssv/eth1/abiparser"
	"github.com/bloxapp/ssv/utils/logex"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseOperatorRegistrationEvent(t *testing.T) {
	var rawOperatorRegistration = `{
  "address": "0x2EAD684aa2E10E31370830F00E0812bE6205F5f9",
  "topics": [
	"0x26a77904793977b23eb8b2d412c486276510e0dc1966a4a2936d4bea0ff86e9d",
	"0x0000000000000000000000000000000000000000000000000000000000000001",
	"0x00000000000000000000000097a6c1f3aab5427b901fb135ed492749191c0f1f"
  ],
  "data": "0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000001e5755170500000000000000000000000000000000000000000000000000000000000000005474e532d3100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424e32705863457872656d643254586476527a684e64455679556a494b524768554d6b313164456c6d59556430566d784d654456574b326734616d7772646e6c7854315976636d784b5245566c517939484d7a567056304d3057455533526e464b55566331516d707651575a315458685165677052517a5a364d45453162314933656e52755748553263305633546b684a534668335245464954486c54645664514d334247596c6f30516e63356231465a54554a6d62564e734c33685852307379566e4e336156686b436b4e4663555a4b526d644e55466b334e6c4a5159306f325232646b545763725756525257565646616d6c52546a4670646d4a4b5a6a5257615570435254637262564e7465465a4e4e54417a566d6c7951575a6e646b494b656e426e64544e7a64485a496448705256315a3265484a304e545230526d39444d48526d5745315252584e53553056745456526f566b686f63566f725a544a434f43396b545751325231466f646e45355a58523152517068516b786f536c704655586c704d6b6c7055553032556c6732613031765a476447556d6376656d747454465a5851305649547a457a61465635526b6f78616e67314c304d3562454979553256454e57396a64316834436d4a525355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b00000000000000000000000000000000000000000000000000000000",
  "blockNumber": "0x6E1070",
  "transactionHash": "0x79478d46847aca9aa93f351c4b9c2126739a746b916da6445c0e64ab227fd016"
}`

	logger := logex.Build("test", zap.DebugLevel, nil)
	t.Run("v2 operator added", func(t *testing.T) {
		LogOperatorRegistration, contractAbi := unmarshalLog(t, rawOperatorRegistration, V2)
		abiParser := NewParser(logger, V2)
		parsed, err := abiParser.ParseOperatorRegistrationEvent(*LogOperatorRegistration, contractAbi)
		var malformedEventErr *abiparser.MalformedEventError
		require.NoError(t, err)
		require.False(t, errors.As(err, &malformedEventErr))
		require.NotNil(t, contractAbi)
		require.NotNil(t, parsed)
		require.Equal(t, "GNS-1", parsed.Name)
		require.Equal(t, "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBN2pXcExremd2TXdvRzhNdEVyUjIKRGhUMk11dElmYUd0VmxMeDVWK2g4amwrdnlxT1YvcmxKREVlQy9HMzVpV0M0WEU3RnFKUVc1QmpvQWZ1TXhQegpRQzZ6MEE1b1I3enRuWHU2c0V3TkhJSFh3REFITHlTdVdQM3BGYlo0Qnc5b1FZTUJmbVNsL3hXR0syVnN3aVhkCkNFcUZKRmdNUFk3NlJQY0o2R2dkTWcrWVRRWVVFamlRTjFpdmJKZjRWaUpCRTcrbVNteFZNNTAzVmlyQWZndkIKenBndTNzdHZIdHpRV1Z2eHJ0NTR0Rm9DMHRmWE1RRXNSU0VtTVRoVkhocVorZTJCOC9kTWQ2R1FodnE5ZXR1RQphQkxoSlpFUXlpMklpUU02Ulg2a01vZGdGUmcvemttTFZXQ0VITzEzaFV5Rkoxang1L0M5bEIyU2VENW9jd1h4CmJRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K",
			string(parsed.PublicKey))
		require.Equal(t, "0x97a6C1f3aaB5427B901fb135ED492749191C0f1F", parsed.OwnerAddress.Hex())
		require.Equal(t, uint32(1), parsed.Id)
	})
}

func TestParseValidatorRegistrationEvent(t *testing.T) {
	var rawValidatorRegistration = `{
  "address": "0x2EAD684aa2E10E31370830F00E0812bE6205F5f9",
  "topics": [
	"0x888b4bb563730efc1c420fb22b503c3551134948a3a3dce4ffab6380e9ce5025",
	"0x000000000000000000000000ceefd323dd28a8d9514eddfec45a6c81800a7d49"
  ],
  "data": "0x000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000003a00000000000000000000000000000000000000000000000000000000000000030b954f437ea50c52ebb1d4962e614d50d3c64773614204c79fef7cf6bf190a9db53ed92b92f24a6f073e874ba99f4da5800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000007000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000001a00000000000000000000000000000000000000000000000000000000000000030abcdcbb92038af19dc83f9a3f9889ddcd52fc2d943a51e523a4a0a30e9caf6463ce665d77361310ddbb2229cca72b48f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030a3f0ba91a5c9fbf87d4bfb8dae0cc2e44588bbc5e9e9159722d5d6f21f644454a71707c33690b0119816e48cf53850e4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030a5b761866a5dd4abf444fdafe4be1adba00c64317083296c7ab5b524dbf2e0711a81c457cecd3717207632fc2fbda38d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030b4661cb225d4d3987c7cc3c48ae67863140f10ac964207812dbf713830a63091ae11ded575879ff4470c158fbb84e38d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000240000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000005c000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158534e766878314a51436c7a33427a6f596b784b333266306d43635177647948776d3377426c4e4c507050715846625755474f7a6b57474468664b5a3454793931665333686e4f4c496b34736e4d303074446b37734c4963384a396a73564b314f45562b36323652746a4b4e734b6a5173384759367a64587171485a516476533962626e2f67787264796a4330543238324a655862396b4d616e61373348466c62414d306f637a2f4874744a3165307a7a2b4b33554a6566494935743433304d61436f694c4d644662564b31476d477558666852416d7962324d4c725a4b56346f58697567756f46664a43714a694e66315666425869793158736f7445783743596c3258424d49624f4d43763871435052744a42564e58434f35525776362f6e62736a55536654556b775057342b3349773852417279467436664d6a7079464b5167764a796a49394a2f7035663636594d427464676c673d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158522f314a6559564643423745734a746c31773147522f4a6c4145353670487444586d59543772537a38344d6d4a4f7947735775693263537472316745506343356c38306c7670367a336d6341335332416b695961444b7772786765374b736b424b6c64347767454e5a4c536362523169566e55312f4d6b444f747443325079556731507579616b42686577466a384c3071375a4e687532463947576e423573476b464266326937364152714c74496d5576732b664b636b686842706e58617375556e7934524156555959634a32457637556a367770753972683858466563314a4254486d4f644b785776724448445541454746464b5830316665342f465a2b36616a536659545752794b635a55616d72754a796957327353474a736845717a44445671556e507369746349555649437946574d4e34497032587935356f47427336524f47456246662f5a2b366b446e573641366f75513d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000015853347355425542484268454b5343524245307662596f624d62527049545450646c6e514669587851545138784d6b6f386d625a75544c392b37526b5957597045362f7242526573317735525954355334563934547152585054417a7a6930544b7570747744665934344d4e30716b5139444c4e2f6e333959434346524e716d6c7a30595a5657555441685061535646744f39526b6c65694242665532626d363462414144464d5a482b64684c7137354e692b6e58766c306f6535563352382b6f7369637653546437367839354962426f5237364c65785857626b52392b6f756665344c62692f4b2b4f356d6167775573344c4d38734f614249797a6773734734324b37733278777763625046702b6c39616a7538377276624f41694564786a2f593868637669637a74506e69743250646443556b6a7a416a31666b426c3168663774696a75415a684c4432426535735a7134703348773d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158753172755779454c4e682b536b6174347a584175322f73726578374d49744553484e69492f785333762b7042454d357548763474707a622b45503763664b336f4e6f59356452313251493576706a71614a6458322f68744d6c4c78514835384170436172666f772f345445454871336376536770535252717a6c75374e586c6f4379344d6b424a2f5043446d6134376f334c70784c73574e59567079697953664f454435545a6744534f316a597550333375326f6179587830316144745142516f336e2b2f46677458654d7176732b6f52576c7a6b6e586f6c49656459775a4a416235386c714e5a4a616363416543455850717877306f414a44734169396471376a574f774b7376666d615a4139356478315434544a7a6f466e44697750596f7442546d387a31747148784956454255435a486a366f794371522f734b70495271645845716578333363314c757174643139443932413d3d0000000000000000",
  "blockNumber": "0x6E10A0",
  "transactionHash": "0x836169107c9e68eb9372daf220281b73552a6fcd99f188ca4335029d2513439d"
}`

	t.Run("v2 validator added", func(t *testing.T) {
		vLogValidatorRegistration, contractAbi := unmarshalLog(t, rawValidatorRegistration, V2)
		abiParser := NewParser(logex.Build("test", zap.InfoLevel, nil), V2)
		parsed, err := abiParser.ParseValidatorRegistrationEvent(*vLogValidatorRegistration, contractAbi)
		var malformedEventErr *abiparser.MalformedEventError
		require.NoError(t, err)
		require.NotNil(t, contractAbi)
		require.False(t, errors.As(err, &malformedEventErr))
		require.NotNil(t, parsed)
		require.Equal(t, "b954f437ea50c52ebb1d4962e614d50d3c64773614204c79fef7cf6bf190a9db53ed92b92f24a6f073e874ba99f4da58", hex.EncodeToString(parsed.PublicKey))
		require.Equal(t, "0xcEEfd323DD28a8d9514EDDfeC45a6c81800A7D49", parsed.OwnerAddress.Hex())
		operators := []string{"LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBN2pXcExremd2TXdvRzhNdEVyUjIKRGhUMk11dElmYUd0VmxMeDVWK2g4amwrdnlxT1YvcmxKREVlQy9HMzVpV0M0WEU3RnFKUVc1QmpvQWZ1TXhQegpRQzZ6MEE1b1I3enRuWHU2c0V3TkhJSFh3REFITHlTdVdQM3BGYlo0Qnc5b1FZTUJmbVNsL3hXR0syVnN3aVhkCkNFcUZKRmdNUFk3NlJQY0o2R2dkTWcrWVRRWVVFamlRTjFpdmJKZjRWaUpCRTcrbVNteFZNNTAzVmlyQWZndkIKenBndTNzdHZIdHpRV1Z2eHJ0NTR0Rm9DMHRmWE1RRXNSU0VtTVRoVkhocVorZTJCOC9kTWQ2R1FodnE5ZXR1RQphQkxoSlpFUXlpMklpUU02Ulg2a01vZGdGUmcvemttTFZXQ0VITzEzaFV5Rkoxang1L0M5bEIyU2VENW9jd1h4CmJRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBb3pVaGFzSG9HeE1YS3pUbzgrSHcKWGV0enJtWENYOTdteXBpaHhjL2wxSEllSVVwV2V3NkFNMzlPd1JQZ2VVMFZ3QmQ2NHZhbzZsTTNaQWxTdVZlMgpablN0T01JckJTWGVsYkc0b1BrRG5xZkNNbGJma1RNRlhXVFowdE1IdGJwVkU3N2o0aEpxaUI3ZU13YitwNXUxClovNmVxWjZmRWRnODI5MzN3ZUhhVWNzd2ZJQmhYNlNaUjNlMkJvRUJ2bHljNE5ENEFoNVFaZjMrRWpxSit5dHYKc3hiRm5MNUpLWWhjSlc4YmtCdzNoM2VreUYyY2I2eUE3M3dsTzZhWklaRWJ4QkE0WDl4WjhMSFBaNHJYWG9GbwpoMVFCd1IxOUVhemF5b0h1TmJkWGpBbU9hc1ViT0ttNFJBdk9ya1FwZ1I4S0J4NGMzczk0OFlidTBJRktQb0NICkJ3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBcjZXc09kMzJZVStPeVowVVZtUlYKQkhEREtLM2U1OTRpUzV2dHRLMVJiMlVYd3YwNGZKcGd4L1NQWmlqUmE0eFdmc3ZsaTMxeHg1c2srMlh6OTJ1VQo5TlE4OGRlL0YxemJtanQwM25wWjhaS253cm1LOXZURE9PZFY4M1RiMUNYTzFhb3J2eVM1MERiZTlSbHE2SGNDCnVuTTRaQnk0SHdvZ2pBZjY2YTFCc085eGx2Rjc0UEgrRTJ0Q1k0ZVYwL1M4VFdHbjh4R0dITW5GT0l1UmRMUTAKemMvQ0pPVjBIK1daSEVEZTcyNU8wR1AwTXV0QmNHZWE1R3A4ckZwWHkvMDFBdmlXajBnMDdqMFR1M0hZN0dlSwovZVNTL1hWOGJURG44M0ZQbE54WHdyVml3czl0cGxzTFMxeUxSN0xxT2NYYVl4NHRLY3FrVTQ0UFhmem9UeC9BCmh3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdVIzV0hmU1lhWlE0NjkxenR0aTYKZlBBcExqa29LcysrQS90QWdSVXdHbEhXYm5iNjJPVU4ra0tTUU53VWlNMFRwWGdOVHVSNGpjdWdKa1NTRlRSRAp5WEwvSXRpbzlFZHE3aEhRQ3BEQ0xCVFNYRlNtMjJrNlNRbllGeWs3UVNndnoyQW9mOXJ6YVdBQmVmUkZPdUs5CnFWT00rbzhnRnFwcXlQRnRJRy9CVS9Fb1l2M0FNU1A5UWJCTXRXSkIvcTd2QStZMUFrZEJiYUNuaGFkK1FUWGwKY1VkSzRabHZ1NVdFWkxLdC9OMlU1RGQwaFh4RXBuRlo3L01SNVRnRVl2NFl3aUpHeWNyRTFKWGVSU2MrM21DWQpKekVzYjJPWTBTZE83YjBMcWdqM2hVa0RtcEdVS2NoQlQyaGw0NWJ5ak4valZjUW1rb29lYUgzSCt2R2IvNzhVCkV3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K"}
		for i, pk := range parsed.OperatorPublicKeys {
			require.Equal(t, operators[i], string(pk))
		}
		shares := []string{"abcdcbb92038af19dc83f9a3f9889ddcd52fc2d943a51e523a4a0a30e9caf6463ce665d77361310ddbb2229cca72b48f", "a3f0ba91a5c9fbf87d4bfb8dae0cc2e44588bbc5e9e9159722d5d6f21f644454a71707c33690b0119816e48cf53850e4", "a5b761866a5dd4abf444fdafe4be1adba00c64317083296c7ab5b524dbf2e0711a81c457cecd3717207632fc2fbda38d", "b4661cb225d4d3987c7cc3c48ae67863140f10ac964207812dbf713830a63091ae11ded575879ff4470c158fbb84e38d"}
		for i, pk := range parsed.SharesPublicKeys {
			require.Equal(t, shares[i], hex.EncodeToString(pk))
		}
	})
}

func unmarshalLog(t *testing.T, rawOperatorRegistration string, abiVersion Version) (*types.Log, abi.ABI) {
	var vLogOperatorRegistration types.Log
	err := json.Unmarshal([]byte(rawOperatorRegistration), &vLogOperatorRegistration)
	require.NoError(t, err)
	contractAbi, err := abi.JSON(strings.NewReader(ContractABI(abiVersion)))
	require.NoError(t, err)
	require.NotNil(t, contractAbi)
	return &vLogOperatorRegistration, contractAbi
}
