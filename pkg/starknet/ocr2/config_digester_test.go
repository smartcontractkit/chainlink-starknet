package ocr2_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/NethermindEth/juno/pkg/rpc"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
)

// contract_address: 0x030cd58d6b04baafd0f4bb47652312c79335bd944f9bd8245448cd82e4f04f6b
// [ <BN: 1>, <BN: 1> ]
// {
//   oracles: [
//     {
//       signer: <BN: 330bd957f7d6dc7b0fc9b1c8d792eb85876ebafc7156f0c94ae4616609b1dde>,
//       transmitter: '0x035597ad15679bd62c26d2095bb7fbe9134d7d1420bda9c2a270f72ee2c9c222'
//     },
//     {
//       signer: <BN: 35b8d28cb168dad13b8f02aeb5b13450aa5e7552916095c72efb98bfcc7fa4f>,
//       transmitter: '0x01d794632b9ce1ff18d6c999932a8ab7521caf1de1c141f198c3ef8965ccaae0'
//     },
//     {
//       signer: <BN: 27048193a8696a4e691e1b3a2e91defc4a14931d3d56d79744bd8c0e5738201>,
//       transmitter: '0x03b8d21e7873041f19cbf4ad73b8d85ec4d0eccf1749c2fd107df09573d9af52'
//     },
//     {
//       signer: <BN: 639ed78f722682dcf52f4956a8fbe9605883ab45d20af096738c86ec0df6847>,
//       transmitter: '0x0758906a2332045ae6b410ee17533d1b302044b03f1862df771db3d154a97c0f'
//     }
//   ],
//   f: 1,
//   onchain_config: 1,
//   offchain_config_version: 2,
//   offchain_config: [ <BN: 1>, <BN: 1> ]
// }
// config_digest: 44d35fedccdab024f3611ff2dda5599dd5ba4da9a4501c10e7556bbf1e3e6

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
		mustHex("0330bd957f7d6dc7b0fc9b1c8d792eb85876ebafc7156f0c94ae4616609b1dde"),
		mustHex("035b8d28cb168dad13b8f02aeb5b13450aa5e7552916095c72efb98bfcc7fa4f"),
		mustHex("027048193a8696a4e691e1b3a2e91defc4a14931d3d56d79744bd8c0e5738201"),
		mustHex("0639ed78f722682dcf52f4956a8fbe9605883ab45d20af096738c86ec0df6847"),
	},
	Transmitters: []types.Account{
		"035597ad15679bd62c26d2095bb7fbe9134d7d1420bda9c2a270f72ee2c9c222",
		"01d794632b9ce1ff18d6c999932a8ab7521caf1de1c141f198c3ef8965ccaae0",
		"03b8d21e7873041f19cbf4ad73b8d85ec4d0eccf1749c2fd107df09573d9af52",
		"0758906a2332045ae6b410ee17533d1b302044b03f1862df771db3d154a97c0f",
	},
	F:                     1,
	OnchainConfig:         []byte{1},
	OffchainConfigVersion: 2,
	OffchainConfig:        []byte{1},
}

func TestConfigDigester(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		"SN_GOERLI", // DEFAULT_CHAIN_ID = StarknetChainId.TESTNET
		"030cd58d6b04baafd0f4bb47652312c79335bd944f9bd8245448cd82e4f04f6b",
	)

	digest, err := d.ConfigDigest(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "00044d35fedccdab024f3611ff2dda5599dd5ba4da9a4501c10e7556bbf1e3e6", digest.Hex())
}

func TestConfigDigester_InvalidChainID(t *testing.T) {
	d := ocr2.NewOffchainConfigDigester(
		rpc.ChainID(strings.Repeat("a", 256)), // chain ID is too long
		"42c59a00fd21bdc27c7be3e9cc272a9b684037e4a37417c2d5a920081e6e87c",
	)

	_, err := d.ConfigDigest(testConfig)
	assert.Error(t, err)
}

func FuzzEncoding(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Fuzz(func(t *testing.T, data []byte) {
		require.Equal(t, data, ocr2.DecodeBytes(ocr2.EncodeBytes(data)))
	})
}
