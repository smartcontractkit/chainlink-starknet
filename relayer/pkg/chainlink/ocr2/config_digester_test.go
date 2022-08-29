package ocr2_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		mustHex("03f5df103ef3ae4d8e5ec708abf48bfdbff08f9e8deacc1a0bedf3e93cffd2a3"),
		mustHex("0312c009b6d4cad9bd653450bb81eec18e81733052ee3f6cb3a2c182082db173"),
		mustHex("02fc4861ccb51c1548dee618ed1103fbc1c01be0144da8950ad1d80c1a7bc3ba"),
		mustHex("0331d8d682d098b685a929e0ad1d89a768ea8a1ca254c2dcbddf57521623729e"),
	},
	Transmitters: []types.Account{
		"01ccc16a80a22f0643b217d32798fe6994c823b7838262db80b4e2a867c61caa",
		"00578180df3312211b37f95ff74428270ea5fe1850908c1d6165b7242e614277",
		"0339c8b556e92cd6b7f7b6c0c69c942db62d0385209a387600077be35509d99e",
		"07ac7c65dc083c4552d53dc5b925de392e41255c6eada6a1ebdf1bdcb6b3ba54",
	},
	F: 1,
	OnchainConfig: []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, // version
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 246, // min (-1)
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 59, 154, 202, 0, // max (1000000000)
	},
	OffchainConfigVersion: 2,
	OffchainConfig:        []byte{1},
}

func TestConfigDigester(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		"SN_GOERLI", // DEFAULT_CHAIN_ID = StarknetChainId.TESTNET
		"01dfac180005c5a5efc88d2c37f880320e1764b83dd3a35006690e1ed7da68d7",
	)

	digest, err := d.ConfigDigest(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "0004d46ed94aa1a4bfa938170a3df74f5b286498a411f59fe5be4d00b6eef12d", digest.Hex())
}

func TestConfigDigester_InvalidChainID(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		strings.Repeat("a", 256), // chain ID is too long
		"42c59a00fd21bdc27c7be3e9cc272a9b684037e4a37417c2d5a920081e6e87c",
	)

	_, err := d.ConfigDigest(testConfig)
	assert.Error(t, err, "chainID")
}

func FuzzEncoding(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Fuzz(func(t *testing.T, data []byte) {
		result, err := starknet.DecodeFelts(starknet.EncodeFelts(data))
		require.NoError(t, err)
		require.Equal(t, data, result)
	})
}
