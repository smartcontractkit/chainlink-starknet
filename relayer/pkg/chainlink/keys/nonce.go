package keys

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

var _ NonceManager = (*nonceManager)(nil)

type nonceManager struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	client  *utils.LazyLoad[*starknet.Client]
	ks      Keystore

	chainId string
	n       map[string]*big.Int
	lock    sync.RWMutex
}

func NewNonceManager(lggr logger.Logger, client *utils.LazyLoad[*starknet.Client], ks Keystore) *nonceManager {
	return &nonceManager{
		lggr:   logger.Named(lggr, "NonceManager"),
		client: client,
		ks:     ks,
		n:      map[string]*big.Int{},
	}
}

func (nm *nonceManager) Start(ctx context.Context) error {
	client, err := nm.client.Get()
	if err != nil {
		return err
	}

	// get chain ID
	id, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	nm.chainId = id

	// get keys + nonces
	keys, err := nm.ks.GetAll()
	if err != nil {
		return err
	}
	nm.lock.Lock()
	defer nm.lock.Unlock()
	for i := range keys {
		addr := keys[i].AccountAddressStr()
		n, err := client.AccountNonce(ctx, caigotypes.HexToHash(addr))
		if err != nil {
			return err
		}
		nm.n[addr] = n
	}
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

func (nm *nonceManager) NextNonce(addr caigotypes.Hash, chainId string) (*big.Int, error) {
	if err := nm.validate(addr, chainId); err != nil {
		return nil, err
	}

	n, exists := nm.n[addr.String()]
	if !exists {
		return nil, errors.New("nonce does not exist for key")
	}
	return n, nil
}

func (nm *nonceManager) IncrementNextNonce(addr caigotypes.Hash, chainId string, currentNonce *big.Int) error {
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