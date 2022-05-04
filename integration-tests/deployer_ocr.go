package integration_tests

import (
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
)

func NewOCRDeployer(c blockchain.EVMClient) (*OCRDeployer, error) {
	return &OCRDeployer{c}, nil
}

type OCRDeployer struct {
	client blockchain.EVMClient
}

func (e *OCRDeployer) Deploy() error {
	//TODO implement me
	return nil
}
