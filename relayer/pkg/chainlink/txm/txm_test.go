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
	var confirmedCount int
	var fatalCount int

	// send many transcations to the TXM at once - all should succeed
	t.Run("confirmed-manyTxs", func(t *testing.T) {
		for k := range localKeys {
			for i := 0; i < 5; i++ {
				require.NoError(t, txm.Enqueue(types.Transaction{
					SenderAddress:      k,
					ContractAddress:    token,
					EntryPointSelector: "transfer",
					Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
				}))
			}
		}

		confirmedCount += 10
		validateTxs(t, txm, CONFIRMED, confirmedCount)
	})

	// simulate rpc failed to connect error
	// tx should be retained and then rebroadcast when endpoint is available again
	t.Run("errored-endpointError-confirmed", func(t *testing.T) {
		client.SetURL("http://broken.url")
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
		}))

		// tx should reach errored state
		validateTxs(t, txm, ERRORED, 1)

		// set URL to correct URL then TX should succeed
		client.SetURL(url)
		confirmedCount += 1
		validateTxs(t, txm, CONFIRMED, confirmedCount)
	})

	// send transaction with out of order nonce
	// tx should be sent again with a new nonce
	t.Run("errored-nonceEstimateError-confirmed", func(t *testing.T) {
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
			Nonce:              "1000",                        // nonce too high
		}))

		// tx should reach errored state
		validateTxs(t, txm, ERRORED, 1)

		// tx should then succeed with new nonce
		confirmedCount += 1
		validateTxs(t, txm, CONFIRMED, confirmedCount)
	})

	// send transaction with out of order nonce
	// tx should be sent again with a new nonce
	t.Run("errored-nonceTxError-confirmed", func(t *testing.T) {
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
			Nonce:              "1000",                        // nonce too high
			MaxFee:             "2000000000000000",            // specifying MaxFee skips estimation/simulation
		}))

		// tx should reach errored state
		validateTxs(t, txm, ERRORED, 1)

		// tx should then succeed with new nonce
		confirmedCount += 1
		validateTxs(t, txm, CONFIRMED, confirmedCount)
	})

	// send transaction with too low fee
	// tx should be sent again with proper fee
	t.Run("errored-feeError-confirmed", func(t *testing.T) {
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x0"}, // to, uint256 (lower, higher bytes)
			MaxFee:             "1",                           // maxFee too low
		}))

		// tx should reach errored state
		validateTxs(t, txm, ERRORED, 1)

		// tx should then succeed with new nonce
		confirmedCount += 1
		validateTxs(t, txm, CONFIRMED, confirmedCount)
	})

	// send transcation that reverts at sequencer
	t.Run("fatal-revertTx", func(t *testing.T) {
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x1"}, // to, uint256 (lower, higher bytes)
			MaxFee:             "2000000000000000",            // specifying MaxFee skips estimation/simulation
		}))

		fatalCount += 1
		validateTxs(t, txm, FATAL, fatalCount)
	})

	// send transcation that reverts at estimation
	t.Run("fatal-revertEstimate", func(t *testing.T) {
		require.NoError(t, txm.Enqueue(types.Transaction{
			SenderAddress:      maps.Keys(localKeys)[0],
			ContractAddress:    token,
			EntryPointSelector: "transfer",
			Calldata:           []string{token, "0x1", "0x1"}, // to, uint256 (lower, higher bytes)
		}))

		fatalCount += 1
		validateTxs(t, txm, FATAL, fatalCount)
	})

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())

	// check conditions
	require.True(t, failedFirst)
}

func validateTxs(t *testing.T, txm StarkTXM, status int, total int) {
	var count int
	for i := 0; i < 30; i++ {
		if count == total {
			break
		}

		time.Sleep(time.Second)
		count = txm.TxCount(status)
	}

	require.Equal(t, total, count)
}
