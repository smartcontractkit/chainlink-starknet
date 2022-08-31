// go:build integration

package txm

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys/mocks"
	txmmock "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestTxm(t *testing.T) {
	url := ops.SetupLocalStarkNetNode(t)
	rawLocalKeys := ops.TestKeys(t, 2) // generate 2 keys

	// parse keys into expected format
	localKeys := map[string]keys.Key{}
	for _, k := range rawLocalKeys {
		key := keys.Raw(k).Key()
		account := "0x" + hex.EncodeToString(keys.PubKeyToAccount(key.PublicKey(), ops.DevnetClassHash, ops.DevnetSalt))
		localKeys[account] = key
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

	// test fail first client
	failedFirst := false

	// should be called twice
	getClient := func() (starknet.ReaderWriter, error) {
		if !failedFirst {
			failedFirst = true

			// test return not nil
			return &starknet.Client{}, errors.New("random test error")
		}

		return client, nil
	}

	// mock config to prevent import cycle
	cfg := new(txmmock.Config)
	cfg.On("TxMaxBatchSize").Return(100)
	cfg.On("TxSendFrequency").Return(15 * time.Second)
	cfg.On("TxTimeout").Return(15 * time.Second)

	txm, err := New(lggr, ks, cfg, getClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Healthy())
	require.NoError(t, txm.Ready())

	token := "0x62230ea046a9a5fbc261ac77d03c8d41e5d442db2284587570ab46455fd2488"
	for k := range localKeys {
		for i := 0; i < 5; i++ {
			require.NoError(t, txm.Enqueue(types.Transaction{
				SenderAddress:      k,
				ContractAddress:    token,
				EntryPointSelector: "transfer",
				Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
				// Nonce:              "0",
				MaxFee: "2000000000000000", // specifying MaxFee skips estimation/simulation
			}))
		}
	}

	require.NoError(t, txm.Enqueue(types.Transaction{
		SenderAddress:      maps.Keys(localKeys)[0],
		ContractAddress:    token,
		EntryPointSelector: "transfer",
		Calldata:           []string{token, "0x1", "0x1"}, // to, uint256 (lower, higher bytes)
		// Nonce:              "0",
		MaxFee: "2000000000000000", // specifying MaxFee skips estimation/simulation
	}))
	// check > 0 in flight txs
	var seenInflight bool
	// check txs are moved out of inflight
	var doneInflight bool
	for i := 0; i < 30; i++ {
		// exit condition
		if seenInflight && doneInflight {
			break
		}

		time.Sleep(time.Second)
		count := txm.InflightCount()

		seenInflight = count > 0 || seenInflight
		doneInflight = seenInflight && count == 0
	}

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())

	// check conditions
	require.True(t, seenInflight)
	require.True(t, doneInflight)
	require.True(t, failedFirst)
}
