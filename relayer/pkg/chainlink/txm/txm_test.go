//go:build integration

package txm

import (
	"context"
	"testing"
	"time"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys/mocks"
	txmmock "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestIntegration_Txm(t *testing.T) {
	url := SetupLocalStarknetNode(t)
	rawLocalKeys := TestKeys(t, 2) // generate 2 keys

	// parse keys into expected format
	localKeys := map[string]keys.Key{}
	for _, k := range rawLocalKeys {
		key := keys.Raw(k).Key()
		key.Set(DevnetClassHash, DevnetSalt)
		localKeys[key.AccountAddressStr()] = key
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
	ks.On("GetAll").Return(maps.Values(localKeys), nil)

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
	cfg.On("TxTimeout").Return(10 * time.Second)

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
			require.NoError(t, txm.Enqueue(key, caigotypes.FunctionCall{
				ContractAddress:    key, // send to self
				EntryPointSelector: "get_nonce",
			}))
		}
	}
	time.Sleep(30 * time.Second)

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
}
