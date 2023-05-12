//go:build integration

package txm

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/caigo/test"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys/mocks"
	txmmock "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestIntegration_Txm(t *testing.T) {
	url := SetupLocalStarknetNode(t)
	// url := "http://127.0.0.1:5050/"
	devnet := test.NewDevNet(url)
	accounts, err := devnet.Accounts()
	require.NoError(t, err)

	// parse keys into expected format
	localKeys := map[string]keys.Key{}
	for i := 0; i < 2; i++ {
		privKey, err := caigotypes.HexToBytes(accounts[i].PrivateKey)
		require.NoError(t, err)

		key := keys.Raw(privKey).Key()
		assert.Equal(t, caigotypes.HexToHash(accounts[i].PublicKey), caigotypes.HexToHash(key.ID()))
		assert.Equal(t, caigotypes.HexToHash(accounts[i].Address), caigotypes.HexToHash(key.DevnetAccountAddrStr()))
		localKeys[key.ID()] = key
	}

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
	timeout := 10 * time.Second
	client, err := starknet.NewClient("devnet", url, lggr, &timeout)
	require.NoError(t, err)

	getClient := func() (*starknet.Client, error) {
		return client, err
	}

	// mock config to prevent import cycle
	cfg := txmmock.NewConfig(t)
	cfg.On("TxTimeout").Return(10 * time.Second) // I'm guessing this should actually just be 10?
	cfg.On("ConfirmationPoll").Return(1 * time.Second)

	txm, err := New(lggr, ks, cfg, getClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Ready())

	for k := range localKeys {
		key := caigotypes.HexToHash(k)
		for i := 0; i < 5; i++ {
			require.NoError(t, txm.Enqueue(key, caigotypes.HexToHash(localKeys[k].DevnetAccountAddrStr()), caigotypes.FunctionCall{
				ContractAddress:    caigotypes.HexToHash("0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7"), // send to ETH token contract
				EntryPointSelector: "total_supply",
			}))
		}
	}
	time.Sleep(5 * time.Second)
	var empty bool
	for i := 0; i < 60; i++ {
		count := txm.InflightCount()
		t.Logf("inflight count: %d", count)

		if count == 0 {
			empty = true
			break
		}
	}

	// stop txm
	assert.True(t, empty, "txm timed out while trying to confirm transactions")
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
}
