package keys

import (
	"math/big"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

//go:generate mockery --name Keystore --output ./mocks/ --case=underscore --filename keystore.go

type Keystore interface {
	Get(id string) (Key, error)
	GetAll() ([]Key, error)
}

type NonceManager interface {
	types.Service

	NextSequence(address caigotypes.Hash, chainID string) (*big.Int, error)
	IncrementNextSequence(address caigotypes.Hash, chainID string, currentNonce *big.Int) error
}