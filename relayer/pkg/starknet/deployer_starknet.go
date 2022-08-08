package starknet

import (
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"

	// unused module to keep it go.mod and prevent ambiguous import
	_ "github.com/btcsuite/btcd/chaincfg/chainhash"
)

func NewStarkNetContractDeployer(c blockchain.EVMClient) (*StarkNetContractDeployer, error) {
	return &StarkNetContractDeployer{c}, nil
}

type StarkNetContractDeployer struct {
	client blockchain.EVMClient
}

func (e *StarkNetContractDeployer) Deploy() error {
	//TODO implement me
	return nil
}
