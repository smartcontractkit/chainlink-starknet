package starknet

import (
	caigotypes "github.com/smartcontractkit/caigo/types"
)

type CallOps struct {
	ContractAddress caigotypes.Hash
	Selector        string
	Calldata        []string
}
