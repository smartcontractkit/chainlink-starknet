package keys_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyStoreAdapter(t *testing.T) {
	var (
		getter       = newMemKeyGetter()
		expectedKey  = big.NewInt(12)
		expectedAddr = "addr1"
	)
	getter.Put(expectedAddr, expectedKey)

	lk := keys.NewLooppKeystore(getter)
	adapter := keys.NewKeystoreAdapter(lk)
	// test that adapter implements the loopp spec. signing nil data should not error
	// on existing sender id
	signed, err := adapter.Loopp().Sign(context.Background(), expectedAddr, nil)
	require.Nil(t, signed)
	require.NoError(t, err)

	signed, err = adapter.Loopp().Sign(context.Background(), "not an address", nil)
	require.Nil(t, signed)
	require.Error(t, err)

	x, y, err := adapter.Sign(context.Background(), expectedAddr, big.NewInt(37))
	require.NotNil(t, x)
	require.NotNil(t, y)
	require.NoError(t, err)
}

// memKeyGetter is an in-memory implementation of the KeyGetter interface to be used for testing.
type memKeyGetter struct {
	mu   sync.Mutex
	keys map[string]*big.Int
}

func newMemKeyGetter() *memKeyGetter {
	return &memKeyGetter{
		keys: make(map[string]*big.Int),
	}
}

func (ks *memKeyGetter) Put(senderAddress string, k *big.Int) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.keys[senderAddress] = k
}

var ErrSenderNoExist = errors.New("sender does not exist")

func (ks *memKeyGetter) Get(senderAddress string) (*big.Int, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, exists := ks.keys[senderAddress]
	if !exists {
		return nil, fmt.Errorf("error getting key for sender %s: %w", senderAddress, ErrSenderNoExist)
	}
	return k, nil
}
