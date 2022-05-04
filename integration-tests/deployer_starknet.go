package integration_tests

import (
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
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
