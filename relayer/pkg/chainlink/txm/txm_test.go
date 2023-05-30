//go:build integration

package txm

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	caigogw "github.com/smartcontractkit/caigo/gateway"
	"github.com/smartcontractkit/caigo/test"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestIntegration_Txm(t *testing.T) {
	n := 2 // number of txs per key
	url := SetupLocalStarknetNode(t)
	devnet := test.NewDevNet(url)
	accounts, err := devnet.Accounts()
	require.NoError(t, err)

	// parse keys into expected format
	localKeys := map[string]*big.Int{}
	localAccounts := map[string]string{}
	for i := range accounts {
		privKey, err := caigotypes.HexToBytes(accounts[i].PrivateKey)
		require.NoError(t, err)
		senderAddress := caigotypes.HexToHash(accounts[i].PublicKey).String()
		localKeys[senderAddress] = caigotypes.BytesToBig(privKey)
		localAccounts[senderAddress] = accounts[i].Address
	}

	// mock keystore
	looppKs := NewLooppKeystore(func(id string) (*big.Int, error) {
		_, ok := localKeys[id]
		if !ok {
			return nil, fmt.Errorf("key does not exist id=%s", id)
		}
		return localKeys[id], nil
	})
	ksAdapter := NewKeystoreAdapter(looppKs)
	lggr, observer := logger.TestObserved(t, zapcore.DebugLevel)
	timeout := 10 * time.Second
	client, err := starknet.NewClient(caigogw.GOERLI_ID, url, lggr, &timeout)
	require.NoError(t, err)

	getClient := func() (*starknet.Client, error) {
		return client, err
	}

	// mock config to prevent import cycle
	cfg := mocks.NewConfig(t)
	cfg.On("TxTimeout").Return(20 * time.Second)
	cfg.On("ConfirmationPoll").Return(1 * time.Second)

	txm, err := New(lggr, ksAdapter.Loopp(), cfg, getClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Ready())

	for senderAddressStr := range localKeys {
		senderAddress := caigotypes.HexToHash(senderAddressStr)
		for i := 0; i < n; i++ {
			require.NoError(t, txm.Enqueue(senderAddress, caigotypes.HexToHash(localAccounts[senderAddressStr]), caigotypes.FunctionCall{
				ContractAddress:    caigotypes.HexToHash("0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7"), // send to ETH token contract
				EntryPointSelector: "totalSupply",
			}))
		}
	}
	var empty bool
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		queued, unconfirmed := txm.InflightCount()
		accepted := len(observer.FilterMessageSnippet("ACCEPTED_ON_L2").All())
		t.Logf("inflight count: queued (%d), unconfirmed (%d), accepted (%d)", queued, unconfirmed, accepted)

		// check queue + tx store counts are 0, accepted txs == total txs broadcast
		if queued == 0 && unconfirmed == 0 && n*len(localKeys) == accepted {
			empty = true
			break
		}
	}

	// stop txm
	assert.True(t, empty, "txm timed out while trying to confirm transactions")
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
	assert.Equal(t, 0, observer.FilterLevelExact(zapcore.ErrorLevel).Len())                       // assert no error logs
	assert.Equal(t, n*len(localKeys), len(observer.FilterMessageSnippet("ACCEPTED_ON_L2").All())) // validate txs were successfully included on chain
}
