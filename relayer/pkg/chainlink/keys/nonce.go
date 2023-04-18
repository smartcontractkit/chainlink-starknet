package keys

import (
	"context"
	"math/big"
	"sync"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger

	chainId string
	n       map[string]*big.Int
	lock    sync.RWMutex
}

func NewNonceManager(lggr logger.Logger) *nonceManager {
	return &nonceManager{
		lggr: logger.Named(lggr, "NonceManager"),
		n:    map[string]*big.Int{},
	}
}

func (nm *nonceManager) Start(ctx context.Context) error {
	// TODO: initial sync

	return nil
}

func (nm *nonceManager) Ready() error {
	return nm.starter.Ready()
}

func (nm *nonceManager) Name() string {
	return nm.lggr.Name()
}

func (nm *nonceManager) Close() error {
	return nil
}

func (nm *nonceManager) HealthReport() map[string]error {
	return map[string]error{nm.Name(): nm.starter.Healthy()}
}

func (nm *nonceManager) NextNonce(addr caigotypes.Hash, chainID string) (*big.Int, error) {
	// TODO
	return nil, nil
}

func (nm *nonceManager) IncrementNextNonce(address caigotypes.Hash, chainID string, currentNonce *big.Int) error {
	// TODO
	return nil
}