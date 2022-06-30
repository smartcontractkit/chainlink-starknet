package txm

import (
	"context"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/keys"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/keys/mocks"
	"github.com/stretchr/testify/require"
)

func TestTxm(t *testing.T) {
	// key
	key, err := keys.New()
	require.NoError(t, err)

	// mock keystore
	ks := new(mocks.Keystore)
	ks.On("Get", key.PublicKeyStr()).Return(key, nil)

	lggr, err := logger.New()
	require.NoError(t, err)
	txm, err := New(lggr, ks)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Healthy())
	require.NoError(t, txm.Ready())

	require.NoError(t, txm.Enqueue(types.Transaction{
		SenderAddress: key.PublicKeyStr(),
	}))
	time.Sleep(5 * time.Second)

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
}