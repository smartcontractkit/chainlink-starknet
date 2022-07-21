//go:build integration

package txm

import (
	"context"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet/db"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTxm(t *testing.T) {
	url := SetupLocalStarkNetNode(t)
	localKeys := TestKeys(t, 2) // generate 2 keys

	// mock keystore
	ks := new(mocks.Keystore)
	ks.On("Get", mock.AnythingOfType("string")).Return(
		func(id string) keys.Key {
			return localKeys[id]
		},
		func(id string) error {

			_, ok := localKeys[id]
			if !ok {
				return errors.New("key does not exist")
			}
			return nil
		},
	)

	lggr, err := logger.New()
	require.NoError(t, err)
	cfg := starknet.NewConfig(db.ChainCfg{}, lggr)
	timeout := cfg.RequestTimeout()
	client, err := starknet.NewClient("devnet", url, lggr, &timeout)
	require.NoError(t, err)
	getClient := func() types.Provider {
		return client
	}

	txm, err := New(lggr, ks, cfg, getClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Healthy())
	require.NoError(t, txm.Ready())

	for k := range localKeys {
		for i := 0; i < 5; i++ {
			require.NoError(t, txm.Enqueue(types.Transaction{
				SenderAddress:      k,
				ContractAddress:    k, // send to self
				EntryPointSelector: "get_nonce",
			}))
		}
	}
	time.Sleep(30 * time.Second)

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
}
