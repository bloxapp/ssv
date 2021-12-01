package eth1

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestParseOperatorAddedEvent(t *testing.T) {
	rawOperatorAdded := `{
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

	rawOperatorAdded = `{
   "address":"0x46e039f5feb38b05ca3c845a3492cd575125f220",
   "topics":[
      "0x39b34f12d0a1eb39d220d2acd5e293c894753a36ac66da43b832c9f1fdb8254e",
      "0x000000000000000000000000a7a7720499b7eb1f1408a8a319284bfd2db4a427"
   ],
   "data":"0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000761736431313233000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020ec281dc273f8e649cfdd2d7e4555b37105d0892bac5270f308c5a721787eb197",
   "blockNumber":"0x5a8a05",
   "transactionHash":"0xbcb69662b542baaa75052418cc92009f9fbaeab3dff80f02bd2b5bc3b1304ddc",
   "transactionIndex":"0x0",
   "blockHash":"0x95b6f453e87fe6c40bd3652208e03a407243ffa9315e315f43dca2092ace5218",
   "logIndex":"0x3",
   "removed":false
}`

	var vLogOperatorAdded types.Log
	err := json.Unmarshal([]byte(rawOperatorAdded), &vLogOperatorAdded)
	require.NoError(t, err)
	contractAbi, err := abi.JSON(strings.NewReader(ContractABI()))
	require.NoError(t, err)
	require.NotNil(t, contractAbi)
	parsed, isEventBelongsToOperator, err := ParseOperatorAddedEvent(zap.L(), nil, vLogOperatorAdded.Data, contractAbi)
	require.NoError(t, err)
	require.NotNil(t, contractAbi)
	require.False(t, isEventBelongsToOperator)
	require.NotNil(t, parsed)
	require.Equal(t, "asdas", parsed.Name)
}

func TestParseValidatorAddedEvent(t *testing.T) {
	rawValidatorAdded := `{
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

	rawValidatorAdded = `{
   "address":"0x46e039f5feb38b05ca3c845a3492cd575125f220",
   "topics":[
      "0x39b34f12d0a1eb39d220d2acd5e293c894753a36ac66da43b832c9f1fdb8254e",
      "0x000000000000000000000000a7a7720499b7eb1f1408a8a319284bfd2db4a427"
   ],
   "data":"0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000761736431313233000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020ec281dc273f8e649cfdd2d7e4555b37105d0892bac5270f308c5a721787eb197",
   "blockNumber":"0x5a8a05",
   "transactionHash":"0xbcb69662b542baaa75052418cc92009f9fbaeab3dff80f02bd2b5bc3b1304ddc",
   "transactionIndex":"0x0",
   "blockHash":"0x95b6f453e87fe6c40bd3652208e03a407243ffa9315e315f43dca2092ace5218",
   "logIndex":"0x3",
   "removed":false
}`

	var vLogValidatorAdded types.Log
	err := json.Unmarshal([]byte(rawValidatorAdded), &vLogValidatorAdded)
	require.NoError(t, err)
	contractAbi, err := abi.JSON(strings.NewReader(ContractABI()))
	require.NoError(t, err)
	require.NotNil(t, contractAbi)

	parsed, isEventBelongsToOperator, err := ParseValidatorAddedEvent(zap.L(), nil, vLogValidatorAdded.Data, contractAbi)
	require.NoError(t, err)
	require.NotNil(t, contractAbi)
	require.False(t, isEventBelongsToOperator)
	require.NotNil(t, parsed)
	require.Equal(t, "91db3a13ab428a6c9c20e7104488cb6961abeab60e56cf4ba199eed3b5f6e7ced670ecb066c9704dc2fa93133792381c",
		hex.EncodeToString(parsed.PublicKey))
}
