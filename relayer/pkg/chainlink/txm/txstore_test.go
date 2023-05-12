package txm

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxStore(t *testing.T) {
	t.Parallel()

	t.Run("happypath", func(t *testing.T) {
		t.Parallel()

		s := NewTxStore(big.NewInt(0))
		assert.Equal(t, 0, s.InflightCount())
		require.NoError(t, s.Save(big.NewInt(0), "0x0"))
		assert.Equal(t, 1, s.InflightCount())
		assert.Equal(t, []string{"0x0"}, s.GetUnconfirmed())
		require.NoError(t, s.Confirm("0x0"))
		assert.Equal(t, 0, s.InflightCount())
		assert.Equal(t, []string{}, s.GetUnconfirmed())
	})

	t.Run("save", func(t *testing.T) {
		t.Parallel()

		// create
		s := NewTxStore(big.NewInt(0))

		// accepts tx in order
		require.NoError(t, s.Save(big.NewInt(0), "0x0"))
		assert.Equal(t, 1, s.InflightCount())
		assert.Equal(t, int64(1), s.currentNonce.Int64())

		// accepts tx that skips a nonce
		require.NoError(t, s.Save(big.NewInt(2), "0x2"))
		assert.Equal(t, 2, s.InflightCount())
		assert.Equal(t, int64(1), s.currentNonce.Int64())

		// accepts tx that fills in the missing nonce + fast forwards currentNonce
		require.NoError(t, s.Save(big.NewInt(1), "0x1"))
		assert.Equal(t, 3, s.InflightCount())
		assert.Equal(t, int64(3), s.currentNonce.Int64())

		// skip a nonce for later tests
		require.NoError(t, s.Save(big.NewInt(4), "0x4"))
		assert.Equal(t, 4, s.InflightCount())
		assert.Equal(t, int64(3), s.currentNonce.Int64())

		// rejects old nonce
		require.ErrorContains(t, s.Save(big.NewInt(0), "0xold"), "nonce too low: 0 < 3 (lowest)")
		assert.Equal(t, 4, s.InflightCount())

		// reject already in use nonce
		require.ErrorContains(t, s.Save(big.NewInt(4), "0xskip"), "nonce used: tried to use nonce (4) for tx (0xskip), already used by (0x4)")
		assert.Equal(t, 4, s.InflightCount())

		// reject already in use tx hash
		require.ErrorContains(t, s.Save(big.NewInt(5), "0x0"), "hash used: tried to use tx (0x0) for nonce (5), already used nonce (0)")
		assert.Equal(t, 4, s.InflightCount())

		// race save
		var err0 error
		var err1 error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			err0 = s.Save(big.NewInt(10), "0x10")
			wg.Done()
		}()
		go func() {
			err1 = s.Save(big.NewInt(10), "0x10")
			wg.Done()
		}()
		wg.Wait()
		assert.True(t, !errors.Is(err0, err1) && (err0 != nil || err1 != nil))
	})

	t.Run("confirm", func(t *testing.T) {
		t.Parallel()

		// init store
		s := NewTxStore(big.NewInt(0))
		for i := 0; i < 5; i++ {
			require.NoError(t, s.Save(big.NewInt(int64(i)), "0x"+fmt.Sprintf("%d", i)))
		}

		// confirm in order
		require.NoError(t, s.Confirm("0x0"))
		require.NoError(t, s.Confirm("0x1"))
		assert.Equal(t, 3, s.InflightCount())

		// confirm out of order
		require.NoError(t, s.Confirm("0x4"))
		require.NoError(t, s.Confirm("0x3"))
		require.NoError(t, s.Confirm("0x2"))
		assert.Equal(t, 0, s.InflightCount())

		// confirm unknown/duplicate
		require.ErrorContains(t, s.Confirm("0x2"), "tx hash does not exist - it may already be confirmed")
		require.ErrorContains(t, s.Confirm("0xNULL"), "tx hash does not exist - it may already be confirmed")

		// race confirm
		require.NoError(t, s.Save(big.NewInt(10), "0x10"))
		var err0 error
		var err1 error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			err0 = s.Confirm("0x10")
			wg.Done()
		}()
		go func() {
			err1 = s.Confirm("0x10")
			wg.Done()
		}()
		wg.Wait()
		assert.True(t, !errors.Is(err0, err1) && (err0 != nil || err1 != nil))
	})
}

func TestChainTxStore(t *testing.T) {
	t.Parallel()

	c := ChainTxStore{}

	// automatically save the from address
	require.NoError(t, c.Save(caigotypes.Hash{}, big.NewInt(0), "0x0"))

	// reject saving for existing address and reused hash & nonce
	// error messages are tested within TestTxStore
	assert.Error(t, c.Save(caigotypes.Hash{}, big.NewInt(0), "0x1"))
	assert.Error(t, c.Save(caigotypes.Hash{}, big.NewInt(1), "0x0"))

	// inflight count
	count, err := c.InflightCount(caigotypes.Hash{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	_, err = c.InflightCount(caigotypes.Hash{1})
	require.ErrorContains(t, err, "from address does not exist")

	// get unconfirmed
	list := c.GetUnconfirmed()
	assert.Equal(t, 1, len(list))
	hashes, ok := list[caigotypes.Hash{}]
	assert.True(t, ok)
	assert.Equal(t, []string{"0x0"}, hashes)

	// confirm
	assert.NoError(t, c.Confirm(caigotypes.Hash{}, "0x0"))
	assert.ErrorContains(t, c.Confirm(caigotypes.Hash{1}, "0x0"), "from address does not exist")
	assert.Error(t, c.Confirm(caigotypes.Hash{}, "0x1"))
	list = c.GetUnconfirmed()
	assert.Equal(t, 1, len(list))
	assert.Equal(t, 0, len(list[caigotypes.Hash{}]))
	count, err = c.InflightCount(caigotypes.Hash{})
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}