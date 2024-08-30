package ocr2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const blockOutput = `{"result": {"events": [ {"from_address": "0xd43963a4e875a361f5d164b2e70953598eb4f45fde86924082d51b4d78e489", "keys": ["0x9a144bf4a6a8fd083c93211e163e59221578efcc86b93f8c97c620e7b9608a", "0x0", "0x4b791b801cf0d7b6a2f9e59daf15ec2dd7d9cdc3bc5e037bada9c86e4821c"], "data": ["0x1", "0x4", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737", "0x1", "0x3", "0x1", "0x0", "0xf4240", "0x2", "0x15", "0x263", "0x880a0d9e61d1080d88ee16f1880bcc1960b2080cab5ee01288090dfc04a30", "0x53a0201024220af400004fa5d02cd5170b5261032e71f2847ead36159cf8d", "0xee68affc3c8520904220af400004fa5d02cd5170b5261032e71f2847ead361", "0x59cf8dee68affc3c8520914220af400004fa5d02cd5170b5261032e71f2847", "0xead36159cf8dee68affc3c8520924220af400004fa5d02cd5170b5261032e7", "0x1f2847ead36159cf8dee68affc3c8520934a42307830346363316266613939", "0x65323832653433346165663238313563613137333337613932336364326336", "0x31636630633764653562333236643761383630333733304a42307830346363", "0x31626661393965323832653433346165663238313563613137333337613932", "0x33636432633631636630633764653562333236643761383630333733314a42", "0x30783034636331626661393965323832653433346165663238313563613137", "0x33333761393233636432633631636630633764653562333236643761383630", "0x333733324a4230783034636331626661393965323832653433346165663238", "0x31356361313733333761393233636432633631636630633764653562333236", "0x643761383630333733335200608094ebdc03688084af5f708084af5f788084", "0xaf5f82018c010a202ac49e648a1f84da5a143eeab68c8402c65a1567e63971", "0x7f5732d5e6310c2c761220a6c1ae85186dc981dc61cd14d7511ee5ab70258a", "0x10ac4e03e4d4991761b2c0a61a1090696dc7afed7f61a26887e78e683a1c1a", "0x10a29e5fa535f2edea7afa9acb4fd349b31a10d1b88713982955d79fa0e422", "0x685a748b1a10a07e0118cc38a71d2a9d60bf52938b4a"]}]}}`
const ocr2ContractAddress = "0xd43963a4e875a361f5d164b2e70953598eb4f45fde86924082d51b4d78e489" // matches blockOutput event
// 11 events will trigger pagination logic
const newTransmissionEvents = `{"result": {
	"events": [
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b2465a459",
							"0x66b23867",
							"0x100020301000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b20e5c6c0",
							"0x3b22d7fbef",
							"0x3b2465a459",
							"0x3b28b1692d",
							"0xd4e8dde018993e970",
							"0xbb42b2ce90a2",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a604",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd0",
							"0x573ea9a8602e03417a4a31d55d115748f37a08bbb23adf6347cb699743a998d"
					],
					"transaction_hash": "0x4adcee6da21143dc987a08b77fdf1be9bde5531786ed54a5f8f045f7f1518d5"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b0f955fa9",
							"0x66b2387b",
							"0x103000102000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b0f955fa9",
							"0x3b0f955fa9",
							"0x3b0f955fa9",
							"0x3b1a52d630",
							"0xd4e8dde018993e970",
							"0xbb42b2ce90a2",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a605",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd1",
							"0x23a4d7f2cdf202ea916bbb07814f5bc32ae50e9cdf1fde114d8e6e808b1e965"
					],
					"transaction_hash": "0x2899099a4b35c2d1cd4506177c0fb2ae7b725069cc4a85d5afb297b069636b"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b122cdb00",
							"0x66b2388f",
							"0x101030002000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b10027e01",
							"0x3b10c2e26f",
							"0x3b122cdb00",
							"0x3b1e375618",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a606",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd2",
							"0x143fe26927dd6a302522ea1cd6a821ab06b3753194acee38d88a85c93b3cbc6"
					],
					"transaction_hash": "0x25757f06d80d9d114d66d574711f0b7d5e6db431aac3298d71a74d41346480b"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b0fe5793b",
							"0x66b238a3",
							"0x101020300000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b0f547a79",
							"0x3b0fdb02b3",
							"0x3b0fe5793b",
							"0x3b15b11fc0",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a701",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd3",
							"0x573ea9a8602e03417a4a31d55d115748f37a08bbb23adf6347cb699743a998d"
					],
					"transaction_hash": "0x7d7fc93b94b2e35ad1ce58332ed21dc053d583f65c7e529d8e3d004a0bd8ae1"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b0d25b529",
							"0x66b238b7",
							"0x102030001000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b0cb32b74",
							"0x3b0d24df9f",
							"0x3b0d25b529",
							"0x3b0d25b529",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a702",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd4",
							"0x143fe26927dd6a302522ea1cd6a821ab06b3753194acee38d88a85c93b3cbc6"
					],
					"transaction_hash": "0x79faf4432ebc36288bbf95de04ca960f12a467c164ed1e2f250ddc001fd17e4"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b084e2980",
							"0x66b238cb",
							"0x100030201000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b06505b40",
							"0x3b06cc23d7",
							"0x3b084e2980",
							"0x3b08b9a4de",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a703",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd5",
							"0x23a4d7f2cdf202ea916bbb07814f5bc32ae50e9cdf1fde114d8e6e808b1e965"
					],
					"transaction_hash": "0xf8ebb0d9dd263e310225114909c8b9befcade0efaffcb7b8078a1f761c8bb4"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b0d5e86a4",
							"0x66b238df",
							"0x101000203000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b041e0a05",
							"0x3b09e3e240",
							"0x3b0d5e86a4",
							"0x3b0dba65b0",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a704",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd6",
							"0x1d091b30a2d20ca2509579f8beae26934bfdc3725c0b497f50b353b7a3c636f"
					],
					"transaction_hash": "0x4d792e87657b051a06f56e739a6779d0fe9df091ae6f66bb9f7029f9315e2c3"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3b0ab981c0",
							"0x66b238f3",
							"0x102030001000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3b06f4995d",
							"0x3b0a4ce764",
							"0x3b0ab981c0",
							"0x3b13bc8a26",
							"0xd4e8dde018993e970",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a705",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd7",
							"0x143fe26927dd6a302522ea1cd6a821ab06b3753194acee38d88a85c93b3cbc6"
					],
					"transaction_hash": "0x645ee25aae5a476997349c4d1543d0982678ed8d44c84f3dd7a7f77f412c71c"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3afcb55963",
							"0x66b23907",
							"0x103000201000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3af36a4483",
							"0x3af942ae80",
							"0x3afcb55963",
							"0x3b00dab871",
							"0xd4db3c15baa524ce2",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a706",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd8",
							"0x143fe26927dd6a302522ea1cd6a821ab06b3753194acee38d88a85c93b3cbc6"
					],
					"transaction_hash": "0x45e4733d6a001ce90a4862e0b31cf6993540feb094c9d31149838ae4f530865"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3aeb13458c",
							"0x66b2391b",
							"0x102000103000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3ae7ddbbb7",
							"0x3aeb13458c",
							"0x3aeb13458c",
							"0x3aeb13458c",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a801",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cd9",
							"0x573ea9a8602e03417a4a31d55d115748f37a08bbb23adf6347cb699743a998d"
					],
					"transaction_hash": "0x7dec2e2d75a5990a6d20f42724aeae28695b840199b0cb911535828a00ac704"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3af06fe86a",
							"0x66b2392f",
							"0x101020300000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3ae8d58830",
							"0x3aed511116",
							"0x3af06fe86a",
							"0x3af389d680",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a802",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cda",
							"0x1d091b30a2d20ca2509579f8beae26934bfdc3725c0b497f50b353b7a3c636f"
					],
					"transaction_hash": "0x3fe8250da4794708b523f75a57ea7a34196828d8b31d8cb27b03bd798486462"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3af135a645",
							"0x66b23943",
							"0x102030001000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3aede0fa8d",
							"0x3af0e645aa",
							"0x3af135a645",
							"0x3af135a645",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a803",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cdb",
							"0x23a4d7f2cdf202ea916bbb07814f5bc32ae50e9cdf1fde114d8e6e808b1e965"
					],
					"transaction_hash": "0x63cf059d9e2c153b8b600c7bf256c1d46eac35147bf03f5bb748ab4b2fa8cd0"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3ae5e39340",
							"0x66b23957",
							"0x103020001000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3adf456450",
							"0x3ae56a9e68",
							"0x3ae5e39340",
							"0x3af68fa5c0",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a804",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cdc",
							"0x573ea9a8602e03417a4a31d55d115748f37a08bbb23adf6347cb699743a998d"
					],
					"transaction_hash": "0x6b41146f8a1f5e749bb7cc5cafd325c94a92fb084f0a251986a8d0e2354adf2"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3af2dc4d1f",
							"0x66b2396b",
							"0x102010300000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3aeeba4089",
							"0x3aef7dc08c",
							"0x3af2dc4d1f",
							"0x3af4226d00",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a805",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cdd",
							"0x1d091b30a2d20ca2509579f8beae26934bfdc3725c0b497f50b353b7a3c636f"
					],
					"transaction_hash": "0xce398eb37e530ac2b3ff36b8b4825b52ca494358b4c8c34b7f073055e7fb7f"
			},
			{
					"block_hash": "0x20acc9de1b7a76b76ce4596ebbcbba9fceb25bd438cfb577e45d55589bb8848",
					"block_number": 86153,
					"data": [
							"0x3aeec2b180",
							"0x66b2397f",
							"0x102010300000000000000000000000000000000000000000000000000000000",
							"0x4",
							"0x3aecc28a2c",
							"0x3aee3c12d4",
							"0x3aeec2b180",
							"0x3aef022b80",
							"0xd4bd480255eb72602",
							"0xa292d535f5c0",
							"0x454e13523e484df9a580cea129056608a4531135a900bd72adbeaa8a2c8be",
							"0xa1a806",
							"0x0"
					],
					"from_address": "0x132303a40ae2f271f4e1b707596a63f6f2921c4d400b38822548ed1bb0cbe0",
					"keys": [
							"0x19e22f866f4c5aead2809bf160d2b29e921e335d899979732101c6f3c38ff81",
							"0x10cde",
							"0x143fe26927dd6a302522ea1cd6a821ab06b3753194acee38d88a85c93b3cbc6"
					],
					"transaction_hash": "0x360a78756e8b72c937e45e71d7128889cf0a50c678b6ea8b3e15ccdbeb5c60a"
			}
	]
}}`

func TestOCR2Client(t *testing.T) {
	chainID := "SN_SEPOLIA"
	lggr := logger.Test(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		switch {
		case r.RequestURI == "/":
			type Request struct {
				Selector string `json:"entry_point_selector"`
			}
			type Call struct {
				Method string            `json:"method"`
				Params []json.RawMessage `json:"params"`
			}

			call := Call{}
			require.NoError(t, json.Unmarshal(req, &call))

			switch call.Method {
			case "starknet_blockNumber":
				out = []byte(`{"result":777}`)
			case "starknet_chainId":
				out = []byte(`{"result":"0x534e5f4d41494e"}`)
			case "starknet_call":
				raw := call.Params[0]
				reqdata := Request{}
				err := json.Unmarshal([]byte(raw), &reqdata)
				require.NoError(t, err)

				fmt.Printf("%v %v\n", reqdata.Selector, starknetutils.GetSelectorFromNameFelt("latest_transmission_details").String())
				switch reqdata.Selector {
				case starknetutils.GetSelectorFromNameFelt("billing").String():
					// billing response
					out = []byte(`{"result":["0x0","0x0","0x0","0x0"]}`)
				case starknetutils.GetSelectorFromNameFelt("latest_config_details").String():
					// latest config details response
					out = []byte(`{"result":["0x1","0x2","0x4b791b801cf0d7b6a2f9e59daf15ec2dd7d9cdc3bc5e037bada9c86e4821c"]}`)
				case starknetutils.GetSelectorFromNameFelt("latest_transmission_details").String():
					// latest transmission details response
					out = []byte(`{"result":["0x4cfc96325fa7d72e4854420e2d7b0abda72de17d45e4c3c0d9f626016d669","0x0","0x0","0x0"]}`)
				case starknetutils.GetSelectorFromNameFelt("latest_round_data").String():
					// latest transmission details response
					out = []byte(`{"result":["0x0","0x0","0x0","0x0","0x0"]}`)
				case starknetutils.GetSelectorFromNameFelt("link_available_for_payment").String():
					// latest transmission details response
					out = []byte(`{"result":["0x0","0x0"]}`)
				default:
					require.False(t, true, "unsupported contract method %s", reqdata.Selector)
				}
			case "starknet_getEvents":
				eventsReq := starknetrpc.EventsInput{}
				require.NoError(t, json.Unmarshal(call.Params[0], &eventsReq))

				configSetSelector := starknetutils.GetSelectorFromNameFelt("ConfigSet")

				if eventsReq.EventFilter.Keys[0][0].Equal(configSetSelector) {
					out = []byte(blockOutput)
				} else {
					// for new transmission event
					out = []byte(newTransmissionEvents)
				}

			default:
				require.False(t, true, "unsupported RPC method")
			}
		case strings.Contains(r.RequestURI, "/feeder_gateway/get_block"):
			out = []byte(blockOutput)
		default:
			require.False(t, true, "unsupported endpoint")
		}

		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	url := mockServer.URL
	duration := 10 * time.Second
	reader, err := starknet.NewClient(chainID, url, "", lggr, &duration)
	require.NoError(t, err)
	client, err := NewClient(reader, lggr)
	assert.NoError(t, err)

	contractAddress, err := starknetutils.HexToFelt(ocr2ContractAddress)
	require.NoError(t, err)

	t.Run("get billing details", func(t *testing.T) {
		billing, err := client.BillingDetails(context.Background(), contractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", billing)
	})

	t.Run("get latest config details", func(t *testing.T) {
		details, err := client.LatestConfigDetails(context.Background(), contractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", details)

		config, err := client.ConfigFromEventAt(context.Background(), contractAddress, details.Block)
		require.NoError(t, err)
		fmt.Printf("%+v\n", config)
	})

	t.Run("get latest transmission details", func(t *testing.T) {
		transmissions, err := client.LatestTransmissionDetails(context.Background(), contractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", transmissions)
	})

	t.Run("get new transmission event", func(t *testing.T) {
		events, err := client.NewTransmissionsFromEventsAt(context.Background(), contractAddress, 123)
		require.NoError(t, err)
		assert.Len(t, events, 15)
	})

	t.Run("get latest round data", func(t *testing.T) {
		round, err := client.LatestRoundData(context.Background(), contractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", round)
	})

	t.Run("get link available for payment", func(t *testing.T) {
		available, err := client.LinkAvailableForPayment(context.Background(), contractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", available)
	})

	t.Run("get latest transmission", func(t *testing.T) {
		round, err := client.LatestRoundData(context.Background(), contractAddress)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", round)
	})
}
