package ocr2

import (
	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	"github.com/NethermindEth/juno/pkg/rpc"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

// TODO: use libocr constant
const ConfigDigestPrefixStarknet types.ConfigDigestPrefix = 4

var _ types.OffchainConfigDigester = (*offchainConfigDigester)(nil)

type offchainConfigDigester struct {
	chainID  rpc.ChainID // TODO: encode as bytes https://docs.starknet.io/docs/Blocks/transactions/#chain-id
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
	// FIELDS:
	// chain_id
	// contract_address
	// config_count
	// oracles_len
	// hash each oracle
	// f
	// onchain_config
	// offchain_config_version
	// offchain_config_len

	digest := pedersen.ArrayDigest()

	// TODO: add prefix

	configDigest := types.ConfigDigest{}
	digest.FillBytes(configDigest[:])
	return configDigest, nil
}

func (offchainConfigDigester) ConfigDigestPrefix() types.ConfigDigestPrefix {
	return ConfigDigestPrefixStarknet
}
