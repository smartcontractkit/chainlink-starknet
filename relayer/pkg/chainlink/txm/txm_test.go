//go:build integration

package txm

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/curve"
	"github.com/NethermindEth/starknet.go/devnet"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	adapters "github.com/smartcontractkit/chainlink-common/pkg/loop/adapters/starknet"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestIntegration_Txm(t *testing.T) {
	ctx := t.Context()
	var nTransactions uint64 = 2 // Number of txs per key. If you increase that you might have to increase the confirmation timeout
	// url := SetupLocalStarknetNode(t)
	url := "http://127.0.0.1:5050"
	devnet := devnet.NewDevNet(url)
	accounts, err := devnet.Accounts()
	require.NoError(t, err)

	// parse keys into expected format
	type Key struct {
		PrivateKey *big.Int
		Account    string
	}
	localKeys := map[string]Key{}
	for i := range accounts {
		publicKey := accounts[i].PublicKey
		fmt.Printf("account %v pubkey %v\n", accounts[i].Address, publicKey)
		localKeys[publicKey] = Key{
			PrivateKey: starknetutils.HexToBN(accounts[i].PrivateKey),
			Account:    accounts[i].Address,
		}
	}

	// mock keystore
	looppKs := NewLooppKeystore(func(publicKey string) (*big.Int, error) {
		key, ok := localKeys[publicKey]
		if !ok {
			return nil, fmt.Errorf("key does not exist id=%s", publicKey)
		}
		return key.PrivateKey, nil
	})
	ksAdapter := NewKeystoreAdapter(looppKs)

	lggr, _ := logger.TestObserved(t, zapcore.DebugLevel)
	timeout := 10 * time.Second
	client, err := starknet.NewClient("SN_SEPOLIA", url+"/rpc", "", lggr, &timeout)
	require.NoError(t, err)

	getFeederClient := func() (*starknet.FeederClient, error) {
		return starknet.NewTestFeederClient(t), nil
	}

	getClient := func() (*starknet.Client, error) {
		return client, err
	}

	// mock config to prevent import cycle
	cfg := mocks.NewConfig(t)
	cfg.On("TxTimeout").Return(20 * time.Second)
	cfg.On("ConfirmationPoll").Return(1 * time.Second)

	txm, err := New(lggr, ksAdapter.Loopp(), cfg, getClient, getFeederClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Ready())

	accountAddresses := make(map[*felt.Felt]*felt.Felt) // address -> latestNonce
	for publicKeyStr := range localKeys {
		publicKey, err := starknetutils.HexToFelt(publicKeyStr)
		require.NoError(t, err)

		accountAddress, err := starknetutils.HexToFelt(localKeys[publicKeyStr].Account)
		require.NoError(t, err)

		c, err := getClient()
		require.NoError(t, err)
		latestNonce, err := c.AccountNonceLatest(ctx, accountAddress)
		require.NoError(t, err)
		accountAddresses[accountAddress] = latestNonce

		contractAddress, err := starknetutils.HexToFelt("0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7")
		require.NoError(t, err)

		selector := starknetutils.GetSelectorFromNameFelt("totalSupply")

		for range nTransactions {
			require.NoError(t, txm.Enqueue(ctx, accountAddress, publicKey, starknetrpc.FunctionCall{
				ContractAddress:    contractAddress, // send to ETH token contract
				EntryPointSelector: selector,
			}))
		}
	}

	assert.Eventually(t, func() bool {
		queued, unconfirmed := txm.InflightCount()
		return queued == 0 && unconfirmed == 0
	}, 15*time.Second, 500*time.Millisecond)
	require.NoError(t, txm.Close())
	// Ensure all transactions are confirmed via nonce
	for accountAddress, initialNonce := range accountAddresses {
		c, err := getClient()
		require.NoError(t, err)
		latestNonce, err := c.AccountNonceLatest(ctx, accountAddress)
		require.NoError(t, err)
		require.Equal(t, int(0), new(felt.Felt).Add(initialNonce, new(felt.Felt).SetUint64(nTransactions)).Cmp(latestNonce))
	}
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
	x, y, err := curve.Curve.Sign(starkHash, k)
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
