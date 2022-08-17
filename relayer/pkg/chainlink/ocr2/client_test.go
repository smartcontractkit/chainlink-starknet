//go:build integration

package ocr2

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	caigo "github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000000000000000000000000000"

func TestOCR2Client(t *testing.T) {
	// setup testing env
	url := ops.SetupLocalStarkNetNode(t)
	keys := ops.TestKeys(t, 1)
	chainID := gateway.GOERLI_ID
	lggr := logger.Test(t)
	g, err := ops.NewStarknetGauntlet("../../../../")
	g.SetupNetwork(url)

	// set env vars
	require.NoError(t, os.Setenv("PRIVATE_KEY", "0x"+hex.EncodeToString(maps.Values(keys)[0].Raw())))
	require.NoError(t, os.Setenv("ACCOUNT", maps.Keys(keys)[0]))
	require.NoError(t, os.Setenv("BILLING_ACCESS_CONTROLLER", ZERO_ADDRESS))

	// clean up env vars

	// deploy contract
	ocr2ContractAddress, err := g.DeployOCR2ControllerContract(0, 1000000, 10, "test", ZERO_ADDRESS)
	require.NoError(t, err)

	// set config
	cfg := ops.TestOCR2Config
	cfg.Signers = ops.TestOnKeys
	cfg.Transmitters = ops.TestTxKeys
	cfg.OffchainConfig.OffchainPublicKeys = ops.TestOffKeys
	cfg.OffchainConfig.PeerIds = ops.TestOnKeys // use random keys as random p2p ids
	cfg.OffchainConfig.ConfigPublicKeys = ops.TestCfgKeys
	parsedConfig, err := json.Marshal(cfg)
	_, err = g.SetConfigDetails(string(parsedConfig), ocr2ContractAddress)
	require.NoError(t, err)

	duration := 10 * time.Second
	reader, err := starknet.NewClient(chainID, url, lggr, &duration)
	require.NoError(t, err)
	client, err := NewClient(reader, lggr)
	assert.NoError(t, err)

	t.Run("get billing details", func(t *testing.T) {
		billing, err := client.BillingDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", billing)
	})

	t.Run("get latest config details", func(t *testing.T) {
		details, err := client.LatestConfigDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", details)

		config, err := client.ConfigFromEventAt(context.Background(), ocr2ContractAddress, details.Block)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", config)
	})

	t.Run("get latest transmission details", func(t *testing.T) {
		transmissions, err := client.LatestTransmissionDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", transmissions)
	})
}

func TestSelector(t *testing.T) {
	bytes, err := hex.DecodeString("80c5d224cddf12d83d4ae2998d9a35b77d54490de62265c020ac35a6935e13")
	require.NoError(t, err)
	eventKey := new(big.Int)
	eventKey.SetBytes(bytes)
	assert.Equal(t, caigo.GetSelectorFromName("config_set").Cmp(eventKey), 0)
}
