package txm

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	caigotypes "github.com/smartcontractkit/caigo/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

//go:generate mockery --name NonceManagerClient --output ./mocks/ --case=underscore --filename nonce_manager_client.go

type NonceManagerClient interface {
	AccountNonce(context.Context, caigotypes.Felt) (*big.Int, error)
}

type NonceManager interface {
	services.Service

	Register(ctx context.Context, address caigotypes.Felt, chainId string, client NonceManagerClient) error

	NextSequence(address caigotypes.Felt, chainID string) (*big.Int, error)
	IncrementNextSequence(address caigotypes.Felt, chainID string, currentNonce *big.Int) error
}

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger

	n    map[string]map[string]*big.Int // map address + chain ID to nonce
	lock sync.RWMutex
}

func NewNonceManager(lggr logger.Logger) *nonceManager {
	return &nonceManager{
		lggr: logger.Named(lggr, "NonceManager"),
		n:    map[string]map[string]*big.Int{},
	}
}

func (nm *nonceManager) Start(ctx context.Context) error {
	return nm.starter.StartOnce(nm.Name(), func() error { return nil })
}

func (nm *nonceManager) Ready() error {
	return nm.starter.Ready()
}

func (nm *nonceManager) Name() string {
	return nm.lggr.Name()
}

func (nm *nonceManager) Close() error {
	return nm.starter.StopOnce(nm.Name(), func() error { return nil })
}

func (nm *nonceManager) HealthReport() map[string]error {
	return map[string]error{nm.Name(): nm.starter.Healthy()}
}

// Register is used because we cannot pre-fetch nonces. the pubkey is known before hand, but the account address is not known until a job is started and sends a tx
func (nm *nonceManager) Register(ctx context.Context, addr caigotypes.Felt, chainId string, client NonceManagerClient) error {
	nm.lock.Lock()
	defer nm.lock.Unlock()
	addressNonces, exists := nm.n[addr.String()]
	if !exists {
		nm.n[addr.String()] = map[string]*big.Int{}
	}
	_, exists = addressNonces[chainId]
	if !exists {
		n, err := client.AccountNonce(ctx, addr)
		if err != nil {
			return err
		}
		nm.n[addr.String()][chainId] = n
	}

	return nil
}

func (nm *nonceManager) NextSequence(addr caigotypes.Felt, chainId string) (*big.Int, error) {
	if err := nm.validate(addr, chainId); err != nil {
		return nil, err
	}

	nm.lock.RLock()
	defer nm.lock.RUnlock()
	return nm.n[addr.String()][chainId], nil
}

func (nm *nonceManager) IncrementNextSequence(addr caigotypes.Felt, chainId string, currentNonce *big.Int) error {
	if err := nm.validate(addr, chainId); err != nil {
		return err
	}

	nm.lock.Lock()
	defer nm.lock.Unlock()
	n := nm.n[addr.String()][chainId]
	if n.Cmp(currentNonce) != 0 {
		return fmt.Errorf("mismatched nonce for %s: %s (expected) != %s (got)", addr, n, currentNonce)
	}
	nm.n[addr.String()][chainId] = big.NewInt(n.Int64() + 1)
	return nil
}

func (nm *nonceManager) validate(addr caigotypes.Felt, id string) error {
	nm.lock.RLock()
	defer nm.lock.RUnlock()
	if _, exists := nm.n[addr.String()]; !exists {
		return fmt.Errorf("nonce tracking does not exist for key: %s", addr.String())
	}
	if _, exists := nm.n[addr.String()][id]; !exists {
		return fmt.Errorf("nonce does not exist for key: %s and chain: %s", addr.String(), id)
	}
	return nil
}
