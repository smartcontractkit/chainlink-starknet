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
	txmmock "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTxm(t *testing.T) {
	url := SetupLocalStarkNetNode(t)
	localKeys := TestKeys(t, 2) // generate 2 keys

	// mock keystore
	ks := new(mocks.Keystore)
	ks.On("Get", mock.AnythingOfType("string")).Return(
		func(id string) keys.StarkKey {
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
	timeout := 5 * time.Second
	client, err := starknet.NewClient("devnet", url, lggr, &timeout)
	require.NoError(t, err)
	getClient := func() types.Provider {
		return client
	}

	// mock config to prevent import cycle
	cfg := new(txmmock.TxConfig)
	cfg.On("TxMaxBatchSize").Return(100)
	cfg.On("TxSendFrequency").Return(15 * time.Second)
	cfg.On("TxTimeout").Return(10 * time.Second)

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
