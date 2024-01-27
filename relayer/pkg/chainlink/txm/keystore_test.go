package txm_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	adapters "github.com/smartcontractkit/chainlink-common/pkg/loop/adapters/starknet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm"
)

func TestKeystoreAdapterImpl(t *testing.T) {
	t.Run("valid loop keystore impl", func(t *testing.T) {
		goodLoopSignFn := func(ctx context.Context, account string, data []byte) (signed []byte, err error) {
			sig, err := adapters.SignatureFromBigInts(big.NewInt(7), big.NewInt(11))
			require.NoError(t, err)
			return sig.Bytes()
		}
		ksa := txm.NewKeystoreAdapter(&testLoopKeystore{signFn: goodLoopSignFn})

		_, _, err := ksa.Sign(context.Background(), "anything", big.NewInt(42))
		require.NoError(t, err)
	})
	t.Run("invalid loop keystore impl", func(t *testing.T) {
		badLoopSignFn := func(ctx context.Context, account string, data []byte) (signed []byte, err error) {
			return []byte("not an adapter signature"), nil
		}
		ksa := txm.NewKeystoreAdapter(&testLoopKeystore{signFn: badLoopSignFn})

		_, _, err := ksa.Sign(context.Background(), "anything", big.NewInt(42))
		require.ErrorIs(t, err, txm.ErrBadAdapterEncoding)
	})

}

type testLoopKeystore struct {
	signFn func(ctx context.Context, account string, data []byte) (signed []byte, err error)
}

var _ loop.Keystore = &testLoopKeystore{}

func (lk *testLoopKeystore) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	return lk.signFn(ctx, account, data)
}

func (lk *testLoopKeystore) Accounts(ctx context.Context) (accounts []string, err error) {
	return nil, nil
}
