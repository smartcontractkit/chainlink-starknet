package ocr2

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	"github.com/NethermindEth/juno/pkg/rpc"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

// TODO: use libocr constant
const ConfigDigestPrefixStarknet types.ConfigDigestPrefix = 4

var _ types.OffchainConfigDigester = (*offchainConfigDigester)(nil)

type offchainConfigDigester struct {
	chainID  rpc.ChainID
	contract rpc.Address
}

func NewOffchainConfigDigester(chainID rpc.ChainID, contract rpc.Address) offchainConfigDigester {
	return offchainConfigDigester{
		chainID:  chainID,
		contract: contract,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const chunkSize = 31

// Encodes a byte slice as a bunch of felts. First felt indicates the total byte size.
func EncodeBytes(data []byte) (felts []*big.Int) {
	// prefix with len
	length := big.NewInt(int64(len(data)))
	felts = append(felts, length)

	// chunk every 31 bytes
	for i := 0; i < len(data); i += chunkSize {
		chunk := data[i:min(i+chunkSize, len(data))]
		// cast to int
		felt := new(big.Int).SetBytes(chunk)
		felts = append(felts, felt)
	}

	return felts
}

func DecodeBytes(felts []*big.Int) []byte {
	// TODO: validate len > 1

	data := []byte{}
	buf := make([]byte, chunkSize)
	// TODO: validate it fits into int64
	length := int(felts[0].Int64())

	for _, felt := range felts[1:] {
		buf := buf[:min(chunkSize, length)]

		felt.FillBytes(buf)
		data = append(data, buf...)

		length -= len(buf)
	}

	return data
}

// TODO: ConfigDigest is byte[32] but what we really want here is a felt
func (d offchainConfigDigester) ConfigDigest(cfg types.ContractConfig) (types.ConfigDigest, error) {
	configDigest := types.ConfigDigest{}

	// TODO: validate contract_address fits into a felt

	// TODO: probably everything needs to be constructed via felt.Felt so it's proper reduced

	contract_address, valid := new(big.Int).SetString(string(d.contract), 16)
	if !valid {
		return configDigest, errors.New("invalid contract address")
	}

	// TODO: assert len(signers) == len(transmitters)

	oracles := []*big.Int{}
	for i := range cfg.Signers {
		signer := new(big.Int).SetBytes(cfg.Signers[i])
		transmitter, valid := new(big.Int).SetString(string(cfg.Transmitters[i]), 16)
		if !valid {
			return configDigest, errors.New("invalid transmitter")
		}
		oracles = append(oracles, signer, transmitter)
	}

	offchainConfig := EncodeBytes(cfg.OffchainConfig)

	digest := pedersen.ArrayDigest(
		// golang... https://stackoverflow.com/questions/28625546/mixing-exploded-slices-and-regular-parameters-in-variadic-functions
		append(
			append(
				append(
					[]*big.Int{
						new(big.Int).SetBytes([]byte(d.chainID)),       // chain_id
						contract_address,                               // contract_address
						new(big.Int).SetUint64(cfg.ConfigCount),        // config_count
						new(big.Int).SetInt64(int64(len(cfg.Signers))), // oracles_len
					},
					oracles...,
				),
				big.NewInt(int64(cfg.F)), // f
				big.NewInt(1),            // TODO: onchain_config
				new(big.Int).SetUint64(cfg.OffchainConfigVersion), // offchain_config_version
				big.NewInt(int64(len(offchainConfig))),            // offchain_config_len
			),
			offchainConfig..., // offchain_config
		)...,
	)

	digest.FillBytes(configDigest[:])

	// set first two bytes to the digest prefix
	binary.BigEndian.PutUint16(configDigest[:2], uint16(d.ConfigDigestPrefix()))

	return configDigest, nil
}

func (offchainConfigDigester) ConfigDigestPrefix() types.ConfigDigestPrefix {
	return ConfigDigestPrefixStarknet
}
