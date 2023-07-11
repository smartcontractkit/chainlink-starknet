//go:build integration

package txm

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/NethermindEth/starknet.go"
	starknetgw "github.com/NethermindEth/starknet.go/gateway"
	"github.com/NethermindEth/starknet.go/test"
	starknettypes "github.com/NethermindEth/starknet.go/types"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	adapters "github.com/smartcontractkit/chainlink-relay/pkg/loop/adapters/starknet"
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
		privKey, err := starknettypes.HexToBytes(accounts[i].PrivateKey)
		require.NoError(t, err)
		senderAddress, err := starknetutils.HexToFelt(accounts[i].PublicKey)
		require.NoError(t, err)
		localKeys[senderAddress.String()] = starknettypes.BytesToBig(privKey)
		localAccounts[senderAddress.String()] = accounts[i].Address
	}

	// mock keystore
	looppKs := NewLooppKeystore(func(id string) (*big.Int, error) {
		key, ok := localKeys[id]
		if !ok {
			return nil, fmt.Errorf("key does not exist id=%s", id)
		}
		return key, nil
	})
	ksAdapter := NewKeystoreAdapter(looppKs)
	lggr, observer := logger.TestObserved(t, zapcore.DebugLevel)
	timeout := 10 * time.Second
	client, err := starknet.NewClient(starknetgw.GOERLI_ID, url+"/rpc", lggr, &timeout)
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
		senderAddress, err := starknetutils.HexToFelt(senderAddressStr)
		require.NoError(t, err)

		account, err := starknetutils.HexToFelt(localAccounts[senderAddressStr])
		require.NoError(t, err)

		contractAddress, err := starknetutils.HexToFelt("0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7")
		require.NoError(t, err)

		selector := starknettypes.GetSelectorFromNameFelt("totalSupply")

		for i := 0; i < n; i++ {
			require.NoError(t, txm.Enqueue(senderAddress, account, starknettypes.FunctionCall{
				ContractAddress:    contractAddress, // send to ETH token contract
				EntryPointSelector: selector,
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

// LooppKeystore implements [loop.Keystore] interface and the requirements
// of signature d/encoding of the [KeystoreAdapter]
type LooppKeystore struct {
	Get func(id string) (*big.Int, error)
}

func NewLooppKeystore(get func(id string) (*big.Int, error)) *LooppKeystore {
	return &LooppKeystore{
		Get: get,
	}
}

var _ loop.Keystore = &LooppKeystore{}

// Sign implements [loop.Keystore]
// hash is expected to be the byte representation of big.Int
// the return []byte is encodes a starknet signature per [signature.bytes]
func (lk *LooppKeystore) Sign(ctx context.Context, id string, hash []byte) ([]byte, error) {

	k, err := lk.Get(id)
	if err != nil {
		return nil, err
	}
	// loopp spec requires passing nil hash to check existence of id
	if hash == nil {
		return nil, nil
	}

	starkHash := new(big.Int).SetBytes(hash)
	x, y, err := starknetgo.Curve.Sign(starkHash, k)
	if err != nil {
		return nil, fmt.Errorf("error signing data with curve: %w", err)
	}

	sig, err := adapters.SignatureFromBigInts(x, y)
	if err != nil {
		return nil, err
	}
	return sig.Bytes()
}

// TODO what is this supposed to return for starknet?
func (lk *LooppKeystore) Accounts(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}
