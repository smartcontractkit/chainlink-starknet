package txm

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxStore(t *testing.T) {
	t.Parallel()

	t.Run("happypath", func(t *testing.T) {
		t.Parallel()

		call := starknetrpc.FunctionCall{
			ContractAddress:    new(felt.Felt).SetUint64(0),
			EntryPointSelector: new(felt.Felt).SetUint64(0),
		}

		nonce := new(felt.Felt).SetUint64(3)
		publicKey := new(felt.Felt).SetUint64(7)

		s := NewTxStore(nonce)
		assert.True(t, s.GetNextNonce().Cmp(nonce) == 0)
		assert.Equal(t, 0, s.InflightCount())
		require.NoError(t, s.AddUnconfirmed(nonce, "0x42", call, publicKey))
		assert.Equal(t, 1, s.InflightCount())
		assert.Equal(t, 1, len(s.GetUnconfirmed()))
		assert.Equal(t, "0x42", s.GetUnconfirmed()[0].Hash)
		require.NoError(t, s.Confirm(nonce, "0x42"))
		assert.Equal(t, 0, s.InflightCount())
		assert.Equal(t, 0, len(s.GetUnconfirmed()))
		assert.True(t, s.GetNextNonce().Cmp(new(felt.Felt).Add(nonce, new(felt.Felt).SetUint64(1))) == 0)
	})

	t.Run("save", func(t *testing.T) {
		t.Parallel()

		// create
		s := NewTxStore(new(felt.Felt).SetUint64(0))

		call := starknetrpc.FunctionCall{
			ContractAddress:    new(felt.Felt).SetUint64(0),
			EntryPointSelector: new(felt.Felt).SetUint64(0),
		}

		publicKey := new(felt.Felt).SetUint64(7)

		// accepts tx in order
		require.NoError(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(0), "0x0", call, publicKey))
		assert.Equal(t, 1, s.InflightCount())

		// reject tx that skips a nonce
		require.ErrorContains(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(2), "0x2", call, publicKey), "tried to add an unconfirmed tx at a future nonce")
		assert.Equal(t, 1, s.InflightCount())

		// accepts a subsequent tx
		require.NoError(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(1), "0x1", call, publicKey))
		assert.Equal(t, 2, s.InflightCount())

		// reject already in use nonce
		require.ErrorContains(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(1), "0xskip", call, publicKey), "tried to add an unconfirmed tx at an old nonce")
		assert.Equal(t, 2, s.InflightCount())

		// race save
		var err0 error
		var err1 error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			err0 = s.AddUnconfirmed(new(felt.Felt).SetUint64(2), "0x10", call, publicKey)
			wg.Done()
		}()
		go func() {
			err1 = s.AddUnconfirmed(new(felt.Felt).SetUint64(2), "0x10", call, publicKey)
			wg.Done()
		}()
		wg.Wait()
		assert.True(t, !errors.Is(err0, err1) && ((err0 != nil && err1 == nil) || (err0 == nil && err1 != nil)))
		assert.Equal(t, 3, s.InflightCount())

		// check that returned unconfirmed tx's are sorted
		unconfirmed := s.GetUnconfirmed()
		assert.Equal(t, 3, len(unconfirmed))
		assert.Equal(t, 0, unconfirmed[0].Nonce.Cmp(new(felt.Felt).SetUint64(0)))
		assert.Equal(t, 0, unconfirmed[1].Nonce.Cmp(new(felt.Felt).SetUint64(1)))
		assert.Equal(t, 0, unconfirmed[2].Nonce.Cmp(new(felt.Felt).SetUint64(2)))
	})

	t.Run("confirm", func(t *testing.T) {
		t.Parallel()

		call := starknetrpc.FunctionCall{
			ContractAddress:    new(felt.Felt).SetUint64(0),
			EntryPointSelector: new(felt.Felt).SetUint64(0),
		}

		publicKey := new(felt.Felt).SetUint64(7)

		// init store
		s := NewTxStore(new(felt.Felt).SetUint64(0))
		for i := 0; i < 6; i++ {
			require.NoError(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(uint64(i)), "0x"+fmt.Sprintf("%d", i), call, publicKey))
		}

		// confirm in order
		require.NoError(t, s.Confirm(new(felt.Felt).SetUint64(0), "0x0"))
		require.NoError(t, s.Confirm(new(felt.Felt).SetUint64(1), "0x1"))
		assert.Equal(t, 4, s.InflightCount())

		// confirm out of order
		require.NoError(t, s.Confirm(new(felt.Felt).SetUint64(4), "0x4"))
		require.NoError(t, s.Confirm(new(felt.Felt).SetUint64(3), "0x3"))
		require.NoError(t, s.Confirm(new(felt.Felt).SetUint64(2), "0x2"))
		assert.Equal(t, 1, s.InflightCount())

		// confirm unknown/duplicate
		require.ErrorContains(t, s.Confirm(new(felt.Felt).SetUint64(10), "0x10"), "no such unconfirmed nonce")
		// confirm with incorrect hash
		require.ErrorContains(t, s.Confirm(new(felt.Felt).SetUint64(5), "0x99"), "unexpected tx hash")

		// race confirm
		var err0 error
		var err1 error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			err0 = s.Confirm(new(felt.Felt).SetUint64(5), "0x5")
			wg.Done()
		}()
		go func() {
			err1 = s.Confirm(new(felt.Felt).SetUint64(5), "0x5")
			wg.Done()
		}()
		wg.Wait()
		assert.True(t, !errors.Is(err0, err1) && ((err0 != nil && err1 == nil) || (err0 == nil && err1 != nil)))
		assert.Equal(t, 0, s.InflightCount())
	})

	t.Run("resync", func(t *testing.T) {
		t.Parallel()

		call := starknetrpc.FunctionCall{
			ContractAddress:    new(felt.Felt).SetUint64(0),
			EntryPointSelector: new(felt.Felt).SetUint64(0),
		}

		publicKey := new(felt.Felt).SetUint64(7)
		txCount := 6

		// init store
		s := NewTxStore(new(felt.Felt).SetUint64(0))
		for i := 0; i < txCount; i++ {
			require.NoError(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(uint64(i)), "0x"+fmt.Sprintf("%d", i), call, publicKey))
		}
		assert.Equal(t, s.InflightCount(), txCount)

		staleTxs := s.SetNextNonce(new(felt.Felt).SetUint64(0))

		assert.Equal(t, len(staleTxs), txCount)
		for i := 0; i < txCount; i++ {
			staleTx := staleTxs[i]
			assert.Equal(t, staleTx.Nonce.Cmp(new(felt.Felt).SetUint64(uint64(i))), 0)
			assert.Equal(t, staleTx.Call, call)
			assert.Equal(t, staleTx.PublicKey.Cmp(publicKey), 0)
			assert.Equal(t, staleTx.Hash, "0x"+fmt.Sprintf("%d", i))
		}
		assert.Equal(t, s.InflightCount(), 0)

		for i := 0; i < txCount; i++ {
			require.NoError(t, s.AddUnconfirmed(new(felt.Felt).SetUint64(uint64(i)), "0x"+fmt.Sprintf("%d", i), call, publicKey))
		}

		newNextNonce := uint64(txCount - 1)
		staleTxs = s.SetNextNonce(new(felt.Felt).SetUint64(newNextNonce))
		assert.Equal(t, len(staleTxs), 1)
		assert.Equal(t, staleTxs[0].Nonce.Cmp(new(felt.Felt).SetUint64(newNextNonce)), 0)
	})
}

func TestAccountStore(t *testing.T) {
	t.Parallel()

	c := NewAccountStore()

	felt0 := new(felt.Felt).SetUint64(0)
	felt1 := new(felt.Felt).SetUint64(1)

	store0, err := c.CreateTxStore(felt0, felt0)
	require.NoError(t, err)

	store1, err := c.CreateTxStore(felt1, felt1)
	require.NoError(t, err)

	_, err = c.CreateTxStore(felt0, felt0)
	require.ErrorContains(t, err, "TxStore already exists")

	assert.Equal(t, store0, c.GetTxStore(felt0))
	assert.Equal(t, store1, c.GetTxStore(felt1))

	assert.Equal(t, c.GetTotalInflightCount(), 0)

	publicKey := new(felt.Felt).SetUint64(2)

	call := starknetrpc.FunctionCall{
		ContractAddress:    new(felt.Felt).SetUint64(0),
		EntryPointSelector: new(felt.Felt).SetUint64(0),
	}

	// inflight count
	require.NoError(t, store0.AddUnconfirmed(felt0, "0x0", call, publicKey))
	require.NoError(t, store1.AddUnconfirmed(felt1, "0x1", call, publicKey))
	assert.Equal(t, c.GetTotalInflightCount(), 2)

	// get unconfirmed
	m := c.GetAllUnconfirmed()
	assert.Equal(t, 2, len(m))
	hashes0, ok := m[felt0.String()]
	assert.True(t, ok)
	assert.Equal(t, len(hashes0), 1)
	assert.Equal(t, hashes0[0].Hash, "0x0")
	hashes1, ok := m[felt1.String()]
	assert.True(t, ok)
	assert.Equal(t, len(hashes1), 1)
	assert.Equal(t, hashes1[0].Hash, "0x1")
}
