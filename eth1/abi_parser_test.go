package eth1

import (
	"encoding/hex"
	"encoding/json"
	"github.com/bloxapp/ssv/utils/logex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestParseOperatorAddedEvent(t *testing.T) {
	OldRawOperatorAdded := `{
  "address": "0x9573c41f0ed8b72f3bd6a9ba6e3e15426a0aa65b",
  "topics": [
	"0x39b34f12d0a1eb39d220d2acd5e293c894753a36ac66da43b832c9f1fdb8254e"
  ],
  "data": "0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000067ce5c69260bd819b4e0ad13f4b873074d47981100000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000005617364617300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e4255555642623364464e303946596e643554477432636c6f7756465530616d6f4b6232393553555a34546e5a6e636c6b34526d6f7256334e736556705562486c714f4656455a6b5a7957576731565734796454525a545752425a53746a5547597857457372515339514f5668594e3039434e47356d4d51705062306457516a5a33636b4d76616d684d596e5a50534459314d484a3556566c766347565a6147785457486848626b5130646d4e3256485a6a6355784d516974315a54497661586c546546464d634670534c7a5a57436e4e554d325a47636b5676626e704756484675526b4e33513059794f476c51626b7057516d70594e6c517653474e55536a553153555272596e52766447467956545a6a6433644f543068755347743656334a324e326b4b64486c5161314930523255784d576874566b633555577053543351314e6d566f57475a4763305a764e55317855335a7863466c776246687253533936565535744f476f76624846465a465577556c6856636a517854416f7961486c4c57533977566d707a5a32316c56484e4f4e79396163554644613068355a546c47596d74574f565976566d4a556144646f56315a4d5648464855326733516c6b765244646e643039335a6e564c61584579436c52335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330414c53304b00000000000000000000000000000000000000000000000000000000",
  "blockNumber": "0x49f59c",
  "transactionHash": "0x097d9a621ace2ca0c78d115d833edc1901bfe75f107a7b3f427663ea308c12ca",
  "transactionIndex": "0xf",
  "blockHash": "0x9542ecebe9d541e2575cb5577dfd4b73c9b0c3ab634fcac4ce0ff319249c90e4",
  "logIndex": "0xf",
  "removed": false
}`
	rawOperatorAdded := `{
   "address":"0xf1e8f1b98d6b05dcc27e93b1b501c89e286eeba9",
   "topics":[
      "0x39b34f12d0a1eb39d220d2acd5e293c894753a36ac66da43b832c9f1fdb8254e",
      "0x000000000000000000000000a7a7720499b7eb1f1408a8a319284bfd2db4a427"
   ],
   "data":"0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000007617364313132350000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e4255555642646c566d556a684954326c73643364695a6d56344f45707559564d4b623068535655565164585a35615442345954564c52466454575455345632746d557a4a324e6a644a56444d795232526863555176616b746e5131704d624849354e6d6461636d49306247394b615535555348686f62677046633031784d4531365331466b5a56525953584a48644555305747644751546b31557a42726447354657446c69636c497a4e4652454d6b6c4753544a45615467354f465633613270524e306469626a5a735645646e436b686f61324e69626a5654536d644252533876523152444d6c524b62697378626b6458516b64524e6a68355356687057575a4f534846495430707a6347704a546d704a516c563651304e535a336c4c5648704b4e47734b61486c74553370714f555270566c684b636a4a4b555646755430567556564a4c4f47787a543364706155564455456855533278506545526f645446355a6c4d335a565a4d656d566c565773725a45564a556d31504b776f79634539315345706b57575a496230566a61533932637a6459656e677661336430524856484e46705262464a4b5a6b68315a556c6e4d454a76553046324f544e536444684c5347315152586c3661464653526d4e54436d78335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b00000000000000000000000000000000000000000000000000000022",
   "blockNumber":"0x5b5560",
   "transactionHash":"0x7f6d1152566ec73a68a7eb51720530236f149035746bcf955fd943e709f8927b",
   "transactionIndex":"0xa",
   "blockHash":"0x0618f84d1fa9de55f266fdc54d6f0174f3a7ab3d9b0813b0a3195b2bb142e5d0",
   "logIndex":"0x24",
   "removed":false
}`

	logger := logex.Build("test", zap.DebugLevel, nil)
	t.Run("legacy operator added", func(t *testing.T) {
		legacyLogOperatorAdded, legacyContractAbi := unmarshalLog(t, OldRawOperatorAdded, Legacy)
		abiParser := NewParser(logger, Legacy)
		parsed, isEventBelongsToOperator, err := abiParser.ParseOperatorAddedEvent(nil, legacyLogOperatorAdded.Data, legacyContractAbi)
		require.NoError(t, err)
		require.NotNil(t, legacyContractAbi)
		require.False(t, isEventBelongsToOperator)
		require.NotNil(t, parsed)
		require.Equal(t, "asdas", parsed.Name)
	})

	t.Run("v2 operator added", func(t *testing.T) {
		LogOperatorAdded, contractAbi := unmarshalLog(t, rawOperatorAdded, V2)
		abiParser := NewParser(logger, V2)
		parsed, isEventBelongsToOperator, err := abiParser.ParseOperatorAddedEvent(nil, LogOperatorAdded.Data, contractAbi)
		require.NoError(t, err)
		require.NotNil(t, contractAbi)
		require.False(t, isEventBelongsToOperator)
		require.NotNil(t, parsed)
		require.Equal(t, "asd1125", parsed.Name)
		require.Equal(t, "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdlVmUjhIT2lsd3diZmV4OEpuYVMKb0hSVUVQdXZ5aTB4YTVLRFdTWTU4V2tmUzJ2NjdJVDMyR2RhcUQvaktnQ1pMbHI5NmdacmI0bG9KaU5USHhobgpFc01xME16S1FkZVRYSXJHdEU0WGdGQTk1UzBrdG5FWDliclIzNFREMklGSTJEaTg5OFV3a2pRN0dibjZsVEdnCkhoa2NibjVTSmdBRS8vR1RDMlRKbisxbkdXQkdRNjh5SVhpWWZOSHFIT0pzcGpJTmpJQlV6Q0NSZ3lLVHpKNGsKaHltU3pqOURpVlhKcjJKUVFuT0VuVVJLOGxzT3dpaUVDUEhUS2xPeERodTF5ZlM3ZVZMemVlVWsrZEVJUm1PKwoycE91SEpkWWZIb0VjaS92czdYengva3d0RHVHNFpRbFJKZkh1ZUlnMEJvU0F2OTNSdDhLSG1QRXl6aFFSRmNTCmx3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K",
			string(parsed.PublicKey))
		require.Equal(t, "ec281dc273f8e649cfdd2d7e4555b37105d0892bac5270f308c5a721787eb197", parsed.OwnerAddress.Hex())
	})
}

func TestParseValidatorAddedEvent(t *testing.T) {
	legacyRawValidatorAdded := `{
  "address": "0x9573c41f0ed8b72f3bd6a9ba6e3e15426a0aa65b",
  "topics": [
	"0x8674c0b4bd63a0814bf1ae6d64d71cf4886880a8bdbd3d7c1eca89a37d1e9271"
  ],
  "data": "0x000000000000000000000000feedb14d8b2c76fdf808c29818b06b830e8c2c0e000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000003091db3a13ab428a6c9c20e7104488cb6961abeab60e56cf4ba199eed3b5f6e7ced670ecb066c9704dc2fa93133792381c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000b80000000000000000000000000000000000000000000000000000000000000110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000036000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424d455a336545684c647a4a355a7a466a555731486446464b4d6d344b52437474646d64475331564c536e6c34515468365543394e4f446379596c70735233684b62305a4a624735684d44564d634868436545466e543035705a4556584d335a3165485647533245344f4467355558425957517070565751354d316848636b4e5a574546794f557074576a686964456c5554304e3464445634613346336557746855455a784d585647646e46774f4852574f54426f596b3536536c70354d6d786f553156794f485268436b6c755130467963304670616e7058533267314e6d705751314a4d51305253623363324d324a436456424c6545646d5a5735425a5564586448466b595842325531645861485a4a63584278636b7050566b4e576356594b63334e594d3278304f57517a656b453254325a4c4d54425455323478645545764d5539584e6a4e32596e4670534446775a475a73527a427952466b7262554e566269397453566f7a4d533876556a4e5a5a7a6856595170706246685054545a32555642685a44466c527a525253544e4761465a4d61456f77616a4d3462555a6855306446616a46435257687a52314e61536e5258526c6c684f555a4b566c4e4464587035516e705654556876436c68525355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003092aa9de41ee36d746a37c2696816103a052bdcf03af3a9d0bf517fb9ef3c30501fcf34a73a57c78f3d3ca23da1aec4580000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158782b6f4b3052687365507254432f4e77486766625870365a6b343679676375412b6551313733546343303771427239484a714261574636536d552f6c3833663067445474554b35434a364c7032445867437033574b476c6c513775704b72467a77517a46725252503132425968416c746a5a645a347068504f6e4343687637716e7362706d633976494f78654b4963416433696c4e51594857484c302b5758546c4c392f556343565461514d38685268656c6636504a67654e7a5145384f66413555624131667735595a51423865647932446d34734c4642617542386a4548554d4e526458413974356868633345617a7972695845736d48613047494a736e3537747a4c324a375455363566626f6d43766e57517743635a534639344142494477324b54324e6755524c34705932674870756a452f514a53573733537478593343444162654f39754979566448524b735941644238773d3d000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000036000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424d465135526d63724d324a6f5a6b394d63546849524445324d6b4d4b5a4468704d6c5a44576a64475430647756484e4f4d6a52344f4578345754524e526d354b5a56464a656b3030633152594f546c4756324e4c5232644b62307842537a566b65545272546e465a57554675533051765451706a4e3255316348684b53465a4965484a595456567a5955744e55304a6d5a33706f556a5a6e63444e584d47704e53305a3354477836564670686231526a616c56455557356d4d336c514f584577626a426f64556879436e6c465444526862586b354f47313363453158576c70564f57644e576a4a4952544673634846305557396d516a417959574d79533146764f544e46546a4252546d6330656d6857526d466a645656355955703662326f4b6448646c574652435a574a7157453173645749315a584e5855334a45626a64555a465646654745725530566b56444261576d39526347704b596b5275646c526a643039766333557a545374505a44524e5555564f63517078556a52594e5751306133427956557377626d395864455658626d7733555445314d486c6b4f55355254305a5961304e6d57537472646c5259567a4277597a46776232466c5a336c4c64555a43566d704a52446478436b4e335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030a1000cd6aaffd95ebe6baff24fe4d9544b7e0ad6727b54ea36cf8117bd34b19fbb586e1ba16ce84084c0ba0a3fa8f3620000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158435a6d6c4d4b42612f43527568346c4d57764d5a342b5a476476675373556a33416e386b573062544f77626862574d626b656d6667644651326e50344458304d374861703446584c3251774d464c6865474353534f6b7649784b42474c4c5439523253737453706358446e7955634d51464244764b5871365673537167644f61375048485a6874566b6775457037414b59566452554d345258657234536a4e6b4968474d6f5a506f775935665067784c33355864473537736c412b35745869596a4b326b4f6663685562306f6e33312f66356c365a386c31377536427679484d5a756658313932554b597a45566241505a64657a674e6e69506c35495452784f2b49423644415145323233754a6a2b48357857377570524d6a58656943796a34567769437354512b483933787151594c62553051316e4c566a3571525567534f674d52434a6133536432536537777531497633694d773d3d000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000036000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424e32705863457872656d643254586476527a684e64455679556a494b524768554d6b313164456c6d59556430566d784d654456574b326734616d7772646e6c7854315976636d784b5245566c517939484d7a567056304d3057455533526e464b55566331516d707651575a315458685165677052517a5a364d45453162314933656e52755748553263305633546b684a534668335245464954486c54645664514d334247596c6f30516e63356231465a54554a6d62564e734c33685852307379566e4e336156686b436b4e4663555a4b526d644e55466b334e6c4a5159306f325232646b545763725756525257565646616d6c52546a4670646d4a4b5a6a5257615570435254637262564e7465465a4e4e54417a566d6c7951575a6e646b494b656e426e64544e7a64485a496448705256315a3265484a304e545230526d39444d48526d5745315252584e53553056745456526f566b686f63566f725a544a434f43396b545751325231466f646e45355a58523152517068516b786f536c704655586c704d6b6c7055553032556c6732613031765a476447556d6376656d747454465a5851305649547a457a61465635526b6f78616e67314c304d3562454979553256454e57396a64316834436d4a525355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030887c4fbcaf5dafaa60ea73f1944be4a3eaaf55f15fac8d2e717e10d6fcdb4c82cf000305acd3fe6eb43b51cd615df2670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001587468336178485a73434f75687268724c4d4c4e394874717977697533494b454479394c4d77504f6b32494d3066622f576744465671744b4d2b626b6c69327163585273346b6c39382f506e6d7378725630664a786871415343376746424955536739674a356864757a68676d6c7672443639554c376c627847545037676f337a32345741335372646254434f32675553496537675a437349796f4b617a466c4d46477342464c436647676264585555324f4e6f334937703564634271707072324e516a6f745475496146507568684a4545706436685a6b2b49766179325246655448796674704a79394569583241397478364459386b456b4269305841442f796d553230707053307a477463674b7a2b312f645a695963315347444a7a6f42424431777a737848642f6c353752766f6533372b52764d424d3853744c42436a3053324f46485673634b347545375070654368623758773d3d000000000000000000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000036000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e4255555642623364464e303946596e643554477432636c6f7756465530616d6f4b6232393553555a34546e5a6e636c6b34526d6f7256334e736556705562486c714f4656455a6b5a7957576731565734796454525a545752425a53746a5547597857457372515339514f5668594e3039434e47356d4d51705062306457516a5a33636b4d76616d684d596e5a50534459314d484a3556566c766347565a6147785457486848626b5130646d4e3256485a6a6355784d516974315a54497661586c546546464d634670534c7a5a57436e4e554d325a47636b5676626e704756484675526b4e33513059794f476c51626b7057516d70594e6c513153474e55536a553153555272596e52766447467956545a6a6433644f543068755347743656334a324e326b4b64486c5161314930523255784d576874566b633555577053543351314e6d566f57475a4763305a764e55317855335a7863466c776246687253533936565535744f476f76624846465a465577556c6856636a517854416f7961486c4c57533977566d707a5a32316c56484e4f4e79396163554644613068355a546c47596d74574f565976566d4a556144646f56315a4d5648464855326733516c6b765244646e643039335a6e564c61584579436c52335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030b860fe4b61a3f9295a08b53f6676443fb3ba19ed4502540a18c9d30268c4f7018b3b039d6528f06f08a63e4cb0ca67d30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001586859562b4378525131614d4f4d686b4d5174706d4650316c524732433454545a51414135654866746a63484a6b6c745872364b7877614c6f4d704f367170313333592f637a325a362b5553565554702b797655312b4e49306152504835534362557058376e7057497148306f4756656f674a4550514650374e65336c376a493761583876556a6b46596a7a636d52496e696e5a7370723455593344626953582f476a327a795644462f2b5941335a2f2f49366a77746561386c6c31576c3978412b794f70427a356b2b4b726134714a787452367a7668714d646532594836305a424a515342796934722b4c5035642f3279496b4f39344f7451326f4c41576d2f756778436f4d747152314e37444e4e633467614b444869722f7658676b514f475178424e396265353352345179386b69354869655953594e6e4b2f6d6e387374304e307250756442394a334b346278796657574f41413d3d0000000000000000",
  "blockNumber": "0x4a3a2e",
  "transactionHash": "0x20b673d0be280a38daa4f636ec6ad1108c0635dcb35c603f8e401a4120a2b506",
  "transactionIndex": "0x3",
  "blockHash": "0x579a98700bc9f9b1dc6ea3d00f9fd43bf28bd795f615210fd138fe724b8654d4",
  "logIndex": "0x2",
  "removed": false
}`
	rawValidatorAdded := `{
   "address":"0xfdff361ed2b094730fdd8fa9658b370ce6cd4b10",
   "topics":[
      "0x088097840a21a2c763dd9bd97cc2b0b27628bb6a42124a398260fac7f31ff571"
   ],
   "data":"0x0000000000000000000000004e409db090a71d14d32adbfbc0a22b1b06dde7de00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000d200000000000000000000000000000000000000000000000000000000000000f4000000000000000000000000000000000000000000000000000000000000000308095b9398e1a3a58bae632db4d9d981f10c69172911e4a4c04b4c9bfc64b13bd2eeb8f49e280db211186c3c7ca5f3233000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000003600000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000092000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424d6b4e5152455256624846574e32357a4f4746444e575a4f4e33634b6255686955305a51525449325a56564355553971516c453354445a4c576c56735133464c4e32466f636b4e474e474a5561316c6164554a5762554a34626d737a596e453254586853556b354557564e4a556c5a48626770346257746f6257465a566b566c5a7a6c7263564a446157597a6447314f574670524d306c56576d7468625770316448644462335a546546425a4e57564c6347465354334244525841764e30354e54555234636d687a436b5a305a4777324f47493261564636546a6c7559303876595446346544526a6546423661564532656d78735a304a4b4d58413052306c6e566c526f576d6f7a654842314d6c42505457786b51574e764b3039566155594b55316378556e5a775a446857636d564d5a57685462477074616e566b536e466b595746424d314a34656b355964545a5853314d3165546b3352544e72596a6c61656b78714d484a4753554677656c5135616e4a544f51706b5432464e5432785a5a544652647a497a515756764e485248624468695a337059626d464a575374366146646e5446564b5a4468305a7a4e35626b4d355758644757545a6c59566845616d465452576f7a55555a6c436b4e335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e4255555642645464795a6a5644567a52345548464d4e444679654735504e58594b5333517a5a4870325454646d5645645a524641355530526b526e4e6a55574a59616e6868544670684d474e764b334e4f52305278536c5a5a6556465559334a745a30737a4d475a4c4f486c6a5246423655464a696151707654467033535578504e315a4b636c684365466448633073766547394354475646656c4a78556b4d78636c46505444597a59565a35546e466b57577334654856514e5774754b3367344e456c4b6132457851336c68436c4e7965476c455a6e4e5054475579527a63796430786a53336c71556b6c46646c42574d326476535555776257314e64586c6d62586c71516d354561454e7162564e5653315a4e625751724e484678536c63766379384b646c464365474e4a6345464a4d6c42704d3074345a6e52744e31467356575a334b324e5a63484a50596a6c4d5257357362305a6e5354684d543255725533466e52465251564535514f437474523064525a3156744d51707053554633626e5a756355394561446850525845764b3038325254525952464a4762315a7164334a56517a564b516d703161544a724d545a52563268454d45313663546847576e56725a5539756269744b4e45314e436e68335355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e4255555642646d3170576c465162464268546c42755445704e546e52584e32634b5133687663575659545646524d324e695457393353475a6c4e6e646e4f5752325a545649616e68494d3235554e7a6b7a64455a4c61444a776455557a5355557a524756775245317056576c305a6d4645644731425667706f616b6c7563577033526a5a70626b785054306c355332526859573030546e6442576a5a525369746a646e4a684c31646863325a76636d704b5357705464474652613078435a6d68754f4464485357513551586468436d383257555659556e4e684d6c4131566b6c4e5a56464554305273633352494f555135626d4a596444417753304a49646b39694d46424c61464a5555585255596b646f534752425446493161454a586545564b596a674b526e42305a54557a546b3179564774774d304e71556a4257625442705457464d64474e784f57704757546c61546e4a3352464a74656e6c6b5555744d576974304b7a52704d6e5a78627a4269654339525257314861776f76556a46316346684a536b78474e57783461573952516a4a72626b5a724e7a42464e58706c5632526e4c3164434b30784c5369744654336c77626d39714d314e68556e42704c33424a4f5645325632356951584a78436b46525355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002c0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002644c5330744c5331435255644a54694253553045675546564354456c4449457446575330744c533074436b314a53554a4a616b464f516d64726357687261556335647a424351564646526b464254304e425554684254556c4a516b4e6e53304e42555556424d47566c626b685663533957565442724b324e325631464a5756674b6231423363336c566554467a6144464d616d5a6b546b687965573161563070355a6b30306445647564573431547a644561584a75546a644562467068626a526953314659544568324d69745959553173593341766141704a5447566a637a68364e5730724e56426d63454d72615749324d33517859325a6a4b33527351324d30516d353365444250596c427353326c50516c4130523142485546413462466c574e577458566b49344e474a78436b3551533352476331684b59584976556b354e556b7454524556324d57785954445a43576d745a64577872566c707853336c734d6d64476133467563445a6c5a303478643142584f476f335a30396a5657704c5a31494b5630526e4d316f355a5846515a3145335a6b3531543078554b33464e59336b345657317a4e7974436233683561336732543149344c334e4d626c4a34525552334e556c7265576c6e596c4a4d53305a705a6e466b65416f77627a426d556a4e334d3073796231463053453932553239485631464e64454d7765553934567a6845644564544b7a4a515445744d5930316e6347704f65445177534563765a6d70354d31566c566d56564f457051436b74525355524255554643436930744c5330745255354549464a545153425156554a4d53554d675330565a4c5330744c53304b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000003080cfc6d3338484432d20deb9ac95583ae507bd2a12a1df1a57e490800784d163e5d61d165fc811d88ed31a01bc7fb82f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003081f08f30cfd8a9abfd88ac7a9f7980e98e7a05a59b941988d72a519cf039e2c51398e6280e435b667b1e7cfca6643cef000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030b75818cb40c685043ff75a99ff25c5d92a18e77ee31ba463c2238d85e1c6f7c147f2c3a328a3042c5426664b8a951eea000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030b819d24f5cfaf7aa4456ef929b82ef96766516a06bc30fadc5e405c7f3a4a8d2b685819ebfc103b9f59f99502dabe09300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000240000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000005c000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001586c69776847497663534f556f6e3730327049356b4c6f6a6c44416769527954515a664b2f2b687377564c364b6d56644e635a627932466556466e4341773470666d47596a2b35424d53445555692f33436e516945537374503858722f64757737704c696b563555367347504855565a633374594a48307339522f46472b4d483847764f666c755574506b5637734264523862585a4f527865646a66547149634e49444d716543792b3165676d37575133356b6765554939565935367638446a51514d6649445967645564306c53596f494e6354766b30385a634451626e31304a65794734426b2f5937453654614e63775733723233726244322b444f41693555474668504f38712b43714d2f4259744d6b39516d49362f4a534e752b454f3157496971494b394b4c716d42525554526b49434e657a465553662f74787461654364504f4c434154786553347748512b445943306551773d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001584b65515370506a3837554d724c2b53574748374467655a6b596a77487646354f6838506f415952677644456d7870426d38755179367547376f7961384a705647686357354d34366676564a6a4b435470522f653652512b545630453858387438683164496d6e4d5476774b726e434a5177705a533157715845336851457738653546465856354d5459624d3634464139714e652b7235382f7834716c683251354c54456f6837437a5938426246465455664f6d747a35597057365477544e52724d304e482f545971574b4c727171686d314337566332514256706e353930456c516134504e766c565955775645426e64414b53306874697339736b645043383578423836543431304d424b46317a71746c61786346616d6664516169727476754d506b714169396647304a3030556342504e466a4376443776736f4e486a7537436546564f6653725a353355786148716230764671413d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001586730526f3379444a67733666754b476e7036426d5a664f527851425877707a596e714857465a4d413273583339784b733372743469377353393143326b697a4d42466b6b4d6c394d6548476578674f6b4b546f707647413878416269432b4f4d584e4e65435835387847456a58773937344a715248524442316e7371364b685a6877564f4f6e52524a556f444b657064614c4d4f482b455368796a4739737558542b744b4b6833434e716f35696a5432594541796f3675597070346c63372f664d4b4c7237434b7531646e4f4f36462b586a76706155434e75364f6f35397a4d777974386e347169654f58366d78784b54764173697a4e5663476974546873796e4355544b376c493947576955714a5852644437467054457a564b514e652b42644c3532674c4672464873346453374c58784674526f417831523437734468526b706431535647386366597938724a564630427a74513d3d000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000158576e7075482f6a4e365675794b767055546a5972484e53597a746a68675a53436452566f58517836322f4a7a51674357496b776b626f6c504e316a396f53744b7557614b53462b7649722b356e465745646c424778734845584f3446732b6e6959704d536648486c617a744e4455362b5a5a70307a446142597a2b6d434359626a324755502f5855432b464c636a656834757661425461336d586c5239704770796d7679436b6e4e61397459666c4c39653931733558316d4f764a4d327732327342694463317951766b5367695847755262616c31644a51747143394d4567675a6e555a3764517a5058795559706444444d4867534a59654f4b614856657150542f506676377644626554574e367a30496a5a6f386d62503451304e36455a7764392b7963416c7042694b6f5542684e67394d6236444275397a435a676e6256644d64514a4c504f694f3342423230694d4f58556f673d3d0000000000000000",
   "blockNumber":"0x5aab39",
   "transactionHash":"0x44c139ec31a10f53561834a44b597edd1f38ca49233853133d5e5d583f19a1dd",
   "transactionIndex":"0x0",
   "blockHash":"0x24f497fcb229f4592e958cb3102eda7d31ef5ecbb47cb104719e3c316717a780",
   "logIndex":"0x2",
   "removed":false
}`

	t.Run("legacy validator added", func(t *testing.T) {
		vLogValidatorAdded, contractAbi := unmarshalLog(t, legacyRawValidatorAdded, Legacy)
		abiParser := NewParser(logex.Build("test", zap.InfoLevel, nil), Legacy)
		parsed, isEventBelongsToOperator, err := abiParser.ParseValidatorAddedEvent(nil, vLogValidatorAdded.Data, contractAbi)
		require.NoError(t, err)
		require.NotNil(t, contractAbi)
		require.False(t, isEventBelongsToOperator)
		require.NotNil(t, parsed)
		require.Equal(t, "91db3a13ab428a6c9c20e7104488cb6961abeab60e56cf4ba199eed3b5f6e7ced670ecb066c9704dc2fa93133792381c", hex.EncodeToString(parsed.PublicKey))
	})

	t.Run("v2 validator added", func(t *testing.T) {
		vLogValidatorAdded, contractAbi := unmarshalLog(t, rawValidatorAdded, V2)
		abiParser := NewParser(logex.Build("test", zap.InfoLevel, nil), V2)
		parsed, isEventBelongsToOperator, err := abiParser.ParseValidatorAddedEvent(nil, vLogValidatorAdded.Data, contractAbi)
		require.NoError(t, err)
		require.NotNil(t, contractAbi)
		require.False(t, isEventBelongsToOperator)
		require.NotNil(t, parsed)
		require.Equal(t, "8095b9398e1a3a58bae632db4d9d981f10c69172911e4a4c04b4c9bfc64b13bd2eeb8f49e280db211186c3c7ca5f3233", hex.EncodeToString(parsed.PublicKey))
		require.Equal(t, "0x4e409dB090a71D14d32AdBFbC0A22B1B06dde7dE", parsed.OwnerAddress.Hex())
		operators := []string{"LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMkNQRERVbHFWN25zOGFDNWZON3cKbUhiU0ZQRTI2ZVVCUU9qQlE3TDZLWlVsQ3FLN2FockNGNGJUa1ladUJWbUJ4bmszYnE2TXhSUk5EWVNJUlZHbgp4bWtobWFZVkVlZzlrcVJDaWYzdG1OWFpRM0lVWmthbWp1dHdDb3ZTeFBZNWVLcGFST3BDRXAvN05NTUR4cmhzCkZ0ZGw2OGI2aVF6TjluY08vYTF4eDRjeFB6aVE2emxsZ0JKMXA0R0lnVlRoWmozeHB1MlBPTWxkQWNvK09VaUYKU1cxUnZwZDhWcmVMZWhTbGptanVkSnFkYWFBM1J4ek5YdTZXS1M1eTk3RTNrYjlaekxqMHJGSUFwelQ5anJTOQpkT2FNT2xZZTFRdzIzQWVvNHRHbDhiZ3pYbmFJWSt6aFdnTFVKZDh0ZzN5bkM5WXdGWTZlYVhEamFTRWozUUZlCkN3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdTdyZjVDVzR4UHFMNDFyeG5PNXYKS3QzZHp2TTdmVEdZRFA5U0RkRnNjUWJYanhhTFphMGNvK3NOR0RxSlZZeVFUY3JtZ0szMGZLOHljRFB6UFJiaQpvTFp3SUxPN1ZKclhCeFdHc0sveG9CTGVFelJxUkMxclFPTDYzYVZ5TnFkWWs4eHVQNWtuK3g4NElKa2ExQ3lhClNyeGlEZnNPTGUyRzcyd0xjS3lqUklFdlBWM2dvSUUwbW1NdXlmbXlqQm5EaENqbVNVS1ZNbWQrNHFxSlcvcy8KdlFCeGNJcEFJMlBpM0t4ZnRtN1FsVWZ3K2NZcHJPYjlMRW5sb0ZnSThMT2UrU3FnRFRQVE5QOCttR0dRZ1VtMQppSUF3bnZucU9EaDhPRXEvK082RTRYRFJGb1Zqd3JVQzVKQmp1aTJrMTZRV2hEME16cThGWnVrZU9ubitKNE1NCnh3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdm1pWlFQbFBhTlBuTEpNTnRXN2cKQ3hvcWVYTVFRM2NiTW93SGZlNndnOWR2ZTVIanhIM25UNzkzdEZLaDJwdUUzSUUzRGVwRE1pVWl0ZmFEdG1BVgpoaklucWp3RjZpbkxPT0l5S2RhYW00TndBWjZRSitjdnJhL1dhc2ZvcmpKSWpTdGFRa0xCZmhuODdHSWQ5QXdhCm82WUVYUnNhMlA1VklNZVFET0Rsc3RIOUQ5bmJYdDAwS0JIdk9iMFBLaFJUUXRUYkdoSGRBTFI1aEJXeEVKYjgKRnB0ZTUzTk1yVGtwM0NqUjBWbTBpTWFMdGNxOWpGWTlaTnJ3RFJtenlkUUtMWit0KzRpMnZxbzBieC9RRW1HawovUjF1cFhJSkxGNWx4aW9RQjJrbkZrNzBFNXplV2RnL1dCK0xLSitFT3lwbm9qM1NhUnBpL3BJOVE2V25iQXJxCkFRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K", "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMGVlbkhVcS9WVTBrK2N2V1FJWVgKb1B3c3lVeTFzaDFMamZkTkhyeW1aV0p5Zk00dEdudW41TzdEaXJuTjdEbFphbjRiS1FYTEh2MitYYU1sY3AvaApJTGVjczh6NW0rNVBmcEMraWI2M3QxY2ZjK3RsQ2M0Qm53eDBPYlBsS2lPQlA0R1BHUFA4bFlWNWtXVkI4NGJxCk5QS3RGc1hKYXIvUk5NUktTREV2MWxYTDZCWmtZdWxrVlpxS3lsMmdGa3FucDZlZ04xd1BXOGo3Z09jVWpLZ1IKV0RnM1o5ZXFQZ1E3Zk51T0xUK3FNY3k4VW1zNytCb3h5a3g2T1I4L3NMblJ4RUR3NUlreWlnYlJMS0ZpZnFkeAowbzBmUjN3M0syb1F0SE92U29HV1FNdEMweU94VzhEdEdTKzJQTEtMY01ncGpOeDQwSEcvZmp5M1VlVmVVOEpQCktRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K"}
		for i, pk := range parsed.OperatorPublicKeys {
			require.Equal(t, operators[i], string(pk))
		}
		shares := []string{"80cfc6d3338484432d20deb9ac95583ae507bd2a12a1df1a57e490800784d163e5d61d165fc811d88ed31a01bc7fb82f", "81f08f30cfd8a9abfd88ac7a9f7980e98e7a05a59b941988d72a519cf039e2c51398e6280e435b667b1e7cfca6643cef", "b75818cb40c685043ff75a99ff25c5d92a18e77ee31ba463c2238d85e1c6f7c147f2c3a328a3042c5426664b8a951eea", "b819d24f5cfaf7aa4456ef929b82ef96766516a06bc30fadc5e405c7f3a4a8d2b685819ebfc103b9f59f99502dabe093"}
		for i, pk := range parsed.SharesPublicKeys {
			require.Equal(t, shares[i], hex.EncodeToString(pk))
		}
	})
}

func unmarshalLog(t *testing.T, rawOperatorAdded string, abiVersion Version) (*types.Log, abi.ABI) {
	var vLogOperatorAdded types.Log
	err := json.Unmarshal([]byte(rawOperatorAdded), &vLogOperatorAdded)
	require.NoError(t, err)
	contractAbi, err := abi.JSON(strings.NewReader(ContractABI(abiVersion)))
	require.NoError(t, err)
	require.NotNil(t, contractAbi)
	return &vLogOperatorAdded, contractAbi
}
