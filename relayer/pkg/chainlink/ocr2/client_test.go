package ocr2

import (
	"context"
	"testing"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet/db"
)

func TestOCR2Client(t *testing.T) {
	// todo: adjust for e2e tests
	chainID := gateway.GOERLI_ID
	ocr2ContractAddress := "0x756ce9ca3dff7ee1037e712fb9662be13b5dcfc0660b97d266298733e1196b"
	lggr := logger.Test(t)

	cfg := starknet.NewConfig(db.ChainCfg{}, lggr)
	duration := cfg.RequestTimeout()
	reader, err := starknet.NewClient(chainID, "", lggr, &duration)
	require.NoError(t, err)
	client, err := NewClient(reader, lggr)
	assert.NoError(t, err)

	t.Run("get billing details", func(t *testing.T) {
		_, err := client.BillingDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
	})

	t.Run("get latest config details", func(t *testing.T) {
		details, err := client.LatestConfigDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)

		_, err = client.ConfigFromEventAt(context.Background(), ocr2ContractAddress, details.Block)
		assert.NoError(t, err)
	})

	t.Run("get latest transmission details", func(t *testing.T) {
		_, err := client.LatestTransmissionDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
	})
}
