package txm

import (
	"context"
	"fmt"
	"sync"

	"github.com/NethermindEth/juno/core/felt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

//go:generate mockery --name NonceManagerClient --output ./mocks/ --case=underscore --filename nonce_manager_client.go

type NonceManagerClient interface {
	AccountNonce(context.Context, *felt.Felt) (*felt.Felt, error)
}

type NonceManager interface {
	services.Service
	Register(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, chainId string, client NonceManagerClient) error
	NextSequence(address *felt.Felt, chainID string) (*felt.Felt, error)
	IncrementNextSequence(address *felt.Felt, chainID string, currentNonce *felt.Felt) error
	// Resets local account nonce to on-chain account nonce
	Sync(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, chainId string, client NonceManagerClient) error
}

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger

	n    map[string]map[string]*felt.Felt // map address + chain ID to nonce
	lock sync.RWMutex
}

func NewNonceManager(lggr logger.Logger) *nonceManager {
	return &nonceManager{
		lggr: logger.Named(lggr, "NonceManager"),
		n:    map[string]map[string]*felt.Felt{},
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

func (nm *nonceManager) Sync(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, chainId string, client NonceManagerClient) error {
	if err := nm.validate(address, chainId); err != nil {
		return err
	}
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, err := client.AccountNonce(ctx, address)
	if err != nil {
		return err
	}

	nm.n[publicKey.String()][chainId] = n

	return nil
}

// Register is used because we cannot pre-fetch nonces. the pubkey is known before hand, but the account address is not known until a job is started and sends a tx
func (nm *nonceManager) Register(ctx context.Context, addr *felt.Felt, publicKey *felt.Felt, chainId string, client NonceManagerClient) error {
	nm.lock.Lock()
	defer nm.lock.Unlock()
	addressNonces, exists := nm.n[publicKey.String()]
	if !exists {
		nm.n[publicKey.String()] = map[string]*felt.Felt{}
	}
	_, exists = addressNonces[chainId]
	if !exists {
		n, err := client.AccountNonce(ctx, addr)
		if err != nil {
			return err
		}
		nm.n[publicKey.String()][chainId] = n
	}

	return nil
}

func (nm *nonceManager) NextSequence(publicKey *felt.Felt, chainId string) (*felt.Felt, error) {
	if err := nm.validate(publicKey, chainId); err != nil {
		return nil, err
	}

	nm.lock.RLock()
	defer nm.lock.RUnlock()
	return nm.n[publicKey.String()][chainId], nil
}

func (nm *nonceManager) IncrementNextSequence(publicKey *felt.Felt, chainId string, currentNonce *felt.Felt) error {
	if err := nm.validate(publicKey, chainId); err != nil {
		return err
	}

	nm.lock.Lock()
	defer nm.lock.Unlock()
	n := nm.n[publicKey.String()][chainId]
	if n.Cmp(currentNonce) != 0 {
		return fmt.Errorf("mismatched nonce for %s: %s (expected) != %s (got)", publicKey, n, currentNonce)
	}
	one := new(felt.Felt).SetUint64(1)
	nm.n[publicKey.String()][chainId] = new(felt.Felt).Add(n, one)
	return nil
}

func (nm *nonceManager) validate(publicKey *felt.Felt, chainId string) error {
	nm.lock.RLock()
	defer nm.lock.RUnlock()
	if _, exists := nm.n[publicKey.String()]; !exists {
		return fmt.Errorf("nonce tracking does not exist for key: %s", publicKey.String())
	}
	if _, exists := nm.n[publicKey.String()][chainId]; !exists {
		return fmt.Errorf("nonce does not exist for key: %s and chain: %s", publicKey.String(), chainId)
	}
	return nil
}
