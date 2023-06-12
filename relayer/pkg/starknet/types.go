package starknet

import (
	caigotypes "github.com/smartcontractkit/caigo/types"
)

type CallOps struct {
	ContractAddress caigotypes.Felt
	Selector        string
	Calldata        []string
}
