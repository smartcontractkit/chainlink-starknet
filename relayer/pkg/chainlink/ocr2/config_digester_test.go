package ocr2_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/NethermindEth/juno/pkg/rpc"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

func mustHex(data string) []byte {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	return bytes
}

var testConfig = types.ContractConfig{
	ConfigCount: 1,
	Signers: []types.OnchainPublicKey{
		mustHex("04bb6dab7f0defb1a4e7495c8b0462a3601b6471828430e8149c6c4ce4fbb939"),
		mustHex("06d95501c78b021f8b3ba2010495b63c9fdd7792c80f6db63614de0cfe49279f"),
		mustHex("075f974b8699bb3f29ea4ced5586512a7b31814ba9fac9b9ae14c6da8736d868"),
		mustHex("013cfd7d081ba6dd4fd088b543769e8201b9b06dab8b147569ccf9c4a25e11e5"),
	},
	Transmitters: []types.Account{
		"029f74aa5ec305f12307556d1739c181eb71c4e4f9b6b3adbbb12d66f9c837a7",
		"07e72346466d373b3a52abc3e753e56a9fa372eb2d7aefa79e990e6d65720cb8",
		"05ed89dbf3a22af52e7960479faa9e69bbdd17ac752bae9b5aa8ecea67992cd2",
		"04ca4c7fa6e7423219ec2bd64a074473fd61ed80d81af335bff79a1c3bc61178",
	},
	F:                     1,
	OnchainConfig:         []byte{1},
	OffchainConfigVersion: 2,
	OffchainConfig:        []byte{1},
}

func TestConfigDigester(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		"SN_GOERLI", // DEFAULT_CHAIN_ID = StarknetChainId.TESTNET
		"01cb7acedf79ffdd598dbe57b9dfd67e16900323da6af617944f1b763a39503a",
	)

	digest, err := d.ConfigDigest(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "00044e5d4f35325e464c87374b13c512f60e09d1236dd902f4bef4c9aedd7300", digest.Hex())
}

func TestConfigDigester_InvalidChainID(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		rpc.ChainID(strings.Repeat("a", 256)), // chain ID is too long
		"42c59a00fd21bdc27c7be3e9cc272a9b684037e4a37417c2d5a920081e6e87c",
	)

	_, err := d.ConfigDigest(testConfig)
	assert.Error(t, err, "chainID")
}

func FuzzEncoding(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Fuzz(func(t *testing.T, data []byte) {
		result, err := ocr2.DecodeBytes(ocr2.EncodeBytes(data))
		require.NoError(t, err)
		require.Equal(t, data, result)
	})
}
