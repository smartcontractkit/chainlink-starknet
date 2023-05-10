package keys

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	caigotypes "github.com/smartcontractkit/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

//go:generate mockery --name NonceManagerClient --output ./mocks/ --case=underscore --filename nonce_manager_client.go

type NonceManagerClient interface {
	ChainID(context.Context) (string, error)
	AccountNonce(context.Context, caigotypes.Hash) (*big.Int, error)
}

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	client  NonceManagerClient
	ks      Keystore

	chainId string
	n       map[string]*big.Int
	lock    sync.RWMutex
}

func NewNonceManager(lggr logger.Logger, client NonceManagerClient, ks Keystore) *nonceManager {
	return &nonceManager{
		lggr:   logger.Named(lggr, "NonceManager"),
		client: client,
		ks:     ks,
		n:      map[string]*big.Int{},
	}
}

func (nm *nonceManager) Start(ctx context.Context) error {
	return nm.starter.StartOnce(nm.Name(), func() error {
		// get chain ID
		id, err := nm.client.ChainID(ctx)
		if err != nil {
			return err
		}
		nm.chainId = id
		return nil
	})
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
func (nm *nonceManager) Register(ctx context.Context, addr caigotypes.Hash, chainId string) error {
	if chainId != nm.chainId {
		return fmt.Errorf("chain id does not match: %s (expected) != %s (got)", nm.chainId, chainId)
	}

	nm.lock.Lock()
	defer nm.lock.Unlock()
	if _, exists := nm.n[addr.String()]; !exists {
		n, err := nm.client.AccountNonce(ctx, addr)
		if err != nil {
			return err
		}
		nm.n[addr.String()] = n
	}
	return nil
}

func (nm *nonceManager) NextSequence(addr caigotypes.Hash, chainId string) (*big.Int, error) {
	if err := nm.validate(addr, chainId); err != nil {
		return nil, err
	}

	nm.lock.RLock()
	defer nm.lock.RUnlock()
	return nm.n[addr.String()], nil
}

func (nm *nonceManager) IncrementNextSequence(addr caigotypes.Hash, chainId string, currentNonce *big.Int) error {
	if err := nm.validate(addr, chainId); err != nil {
		return err
	}

	nm.lock.Lock()
	defer nm.lock.Unlock()
	n := nm.n[addr.String()]
	if n.Cmp(currentNonce) != 0 {
		return fmt.Errorf("mismatched nonce for %s: %s (expected) != %s (got)", addr, n, currentNonce)
	}
	nm.n[addr.String()] = n.Add(n, big.NewInt(1))
	return nil
}

func (nm *nonceManager) validate(addr caigotypes.Hash, id string) error {
	if id != nm.chainId {
		return fmt.Errorf("chain id does not match: %s (expected) != %s (got)", nm.chainId, id)
	}

	nm.lock.RLock()
	defer nm.lock.RUnlock()
	if _, exists := nm.n[addr.String()]; !exists {
		return fmt.Errorf("nonce does not exist for key: %s", addr.String())
	}
	return nil
}