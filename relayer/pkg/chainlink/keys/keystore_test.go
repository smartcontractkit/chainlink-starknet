package keys_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/caigo"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyStoreAdapter(t *testing.T) {
	var (
		getter = newMemKeyGetter()

		starknetPK         = generateTestKey(t)
		starknetSenderAddr = "legit"
	)
	getter.Put(starknetSenderAddr, starknetPK)

	lk := keys.NewLooppKeystore(getter)
	adapter := keys.NewKeystoreAdapter(lk)
	// test that adapter implements the loopp spec. signing nil data should not error
	// on existing sender id
	signed, err := adapter.Loopp().Sign(context.Background(), starknetSenderAddr, nil)
	require.Nil(t, signed)
	require.NoError(t, err)

	signed, err = adapter.Loopp().Sign(context.Background(), "not an address", nil)
	require.Nil(t, signed)
	require.Error(t, err)

	hash, err := caigo.Curve.PedersenHash([]*big.Int{big.NewInt(42)})
	require.NoError(t, err)
	r, s, err := adapter.Sign(context.Background(), starknetSenderAddr, hash)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NotNil(t, s)

	pubx, puby, err := caigo.Curve.PrivateToPoint(starknetPK)
	require.NoError(t, err)
	require.True(t, caigo.Curve.Verify(hash, r, s, pubx, puby))
}

func generateTestKey(t *testing.T) *big.Int {
	// sadly generating a key can fail, but it should happen infrequently
	// best effort here to  avoid flaky tests
	var generatorDuration = 1 * time.Second
	d, exists := t.Deadline()
	if exists {
		generatorDuration = time.Until(d) / 2
	}
	timer := time.NewTicker(generatorDuration)
	defer timer.Stop()
	var key *big.Int
	var generationErr error

	generated := func() bool {
		select {
		case <-timer.C:
			key = nil
			generationErr = fmt.Errorf("failed to generate test key in allotted time")
			return true
		default:
			key, generationErr = caigo.Curve.GetRandomPrivateKey()
			if generationErr == nil {
				return true
			}
		}
		return false
	}

	for !generated() {
		// nolint
	}
	require.NoError(t, generationErr)
	return key
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
