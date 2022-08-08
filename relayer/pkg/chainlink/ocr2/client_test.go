package ocr2

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	caigo "github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestOCR2Client(t *testing.T) {
	// todo: adjust for e2e tests
	chainID := gateway.GOERLI_ID
	ocr2ContractAddress := "0x756ce9ca3dff7ee1037e712fb9662be13b5dcfc0660b97d266298733e1196b"
	lggr := logger.Test(t)

	duration := 10 * time.Second
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

func TestSelector(t *testing.T) {
	bytes, err := hex.DecodeString("80c5d224cddf12d83d4ae2998d9a35b77d54490de62265c020ac35a6935e13")
	require.NoError(t, err)
	eventKey := new(big.Int)
	eventKey.SetBytes(bytes)
	assert.Equal(t, caigo.GetSelectorFromName("config_set").Cmp(eventKey), 0)
}
