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

// TODO: ConfigDigest is byte[32] but what we really want here is a felt
func (d offchainConfigDigester) ConfigDigest(cfg types.ContractConfig) (types.ConfigDigest, error) {
	configDigest := types.ConfigDigest{}

	contract_address, valid := new(big.Int).SetString(string(d.contract), 16)
	if !valid {
		return configDigest, errors.New("invalid contract address")
	}

	if len(d.chainID) > 31 {
		return configDigest, errors.New("chainID exceeds max length")
	}

	if len(cfg.Signers) != len(cfg.Transmitters) {
		return configDigest, errors.New("must have equal number of signers and transmitters")
	}

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
