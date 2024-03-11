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
	Register(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, client NonceManagerClient) error
	NextSequence(address *felt.Felt) (*felt.Felt, error)
	IncrementNextSequence(address *felt.Felt, currentNonce *felt.Felt) error
	// Resets local account nonce to on-chain account nonce
	Sync(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, client NonceManagerClient) error
}

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger

	n    map[string]*felt.Felt // map public key to nonce
	lock sync.RWMutex
}

func NewNonceManager(lggr logger.Logger) *nonceManager {
	return &nonceManager{
		lggr: logger.Named(lggr, "NonceManager"),
		n:    map[string]*felt.Felt{},
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

func (nm *nonceManager) Sync(ctx context.Context, address *felt.Felt, publicKey *felt.Felt, client NonceManagerClient) error {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	if err := nm.validate(address); err != nil {
		return err
	}

	n, err := client.AccountNonce(ctx, address)
	if err != nil {
		return err
	}

	nm.n[publicKey.String()] = n

	return nil
}

// Register is used because we cannot pre-fetch nonces. the pubkey is known before hand, but the account address is not known until a job is started and sends a tx
func (nm *nonceManager) Register(ctx context.Context, addr *felt.Felt, publicKey *felt.Felt, client NonceManagerClient) error {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	_, exists := nm.n[publicKey.String()]
	if !exists {
		n, err := client.AccountNonce(ctx, addr)
		if err != nil {
			return err
		}
		nm.n[publicKey.String()] = n
	}

	return nil
}

func (nm *nonceManager) NextSequence(publicKey *felt.Felt) (*felt.Felt, error) {
	nm.lock.RLock()
	defer nm.lock.RUnlock()

	if err := nm.validate(publicKey); err != nil {
		return nil, err
	}

	return nm.n[publicKey.String()], nil
}

func (nm *nonceManager) IncrementNextSequence(publicKey *felt.Felt, currentNonce *felt.Felt) error {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	if err := nm.validate(publicKey); err != nil {
		return err
	}

	n := nm.n[publicKey.String()]
	if n.Cmp(currentNonce) != 0 {
		return fmt.Errorf("mismatched nonce for %s: %s (expected) != %s (got)", publicKey, n, currentNonce)
	}
	one := new(felt.Felt).SetUint64(1)
	nm.n[publicKey.String()] = new(felt.Felt).Add(n, one)
	return nil
}

func (nm *nonceManager) validate(publicKey *felt.Felt) error {
	if _, exists := nm.n[publicKey.String()]; !exists {
		return fmt.Errorf("nonce tracking does not exist for key: %s", publicKey.String())
	}
	return nil
}
