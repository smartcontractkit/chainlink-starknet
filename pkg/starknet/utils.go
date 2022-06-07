package starknet

import (
	caigo "github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

func isSetConfigEventFromContract(event *caigotypes.Event, address string) bool {
	if event.FromAddress != address {
		return false
	}

	var isSameEventSelector bool
	eventKey := caigo.GetSelectorFromName("config_set")
	for _, key := range event.Keys {
		if key.Cmp(eventKey) == 0 {
			isSameEventSelector = true
			break
		}
	}

	return isSameEventSelector
}

func parseConfigEventData(eventData []*caigotypes.Felt) (types.ContractConfig, error) {
	index := 0
	// previous_config_block_number - skip

	// latest_config_digest
	index += 1
	digest, err := types.BytesToConfigDigest(eventData[index].Bytes())
	if err != nil {
		return types.ContractConfig{}, err
	}

	// config_count
	index += 1
	configCount := eventData[index].Uint64()

	// oracles_len
	index += 1
	oraclesLen := eventData[index].Uint64()

	// oracles
	index += 1
	oracleMembers := eventData[index:(index + int(oraclesLen)*2)]
	var signers []types.OnchainPublicKey
	var transmitters []types.Account
	for i, member := range oracleMembers {
		if i%2 == 0 {
			signers = append(signers, member.Bytes())
		} else {
			transmitters = append(transmitters, types.Account(member.Hex()))
		}
	}

	// f
	index = index + int(oraclesLen)*2
	f := eventData[index].Uint64()

	// onchain_config
	index += 1
	onchainConfig := eventData[index].Bytes()

	// offchain_config_version
	index += 1
	offchainConfigVersion := eventData[index].Uint64()

	// offchain_config_len
	index += 1
	offchainConfigLen := eventData[index].Uint64()

	// offchain_config
	index += 1
	offchainConfigFelts := eventData[index:(index + int(offchainConfigLen))]
	var offchainConfig []byte
	for _, ocFelt := range offchainConfigFelts {
		offchainConfig = append(offchainConfig, ocFelt.Bytes()...)
	}

	return types.ContractConfig{
		ConfigDigest:          digest,
		ConfigCount:           configCount,
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint8(f),
		OnchainConfig:         onchainConfig,
		OffchainConfigVersion: offchainConfigVersion,
		OffchainConfig:        offchainConfig,
	}, nil
}
