package starknet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var (
	myChainID     = "SN_SEPOLIA"
	myTimeout     = 10 * time.Second
	blockHash, _  = new(felt.Felt).SetString("0x725407fcc3bd43e50884f50f1e0ef32aa9f814af3da475411934a7dbd4b41a")
	blockResponse = []byte(`
		"result": {
				"status": "ACCEPTED_ON_L2",
				"block_hash": "0x725407fcc3bd43e50884f50f1e0ef32aa9f814af3da475411934a7dbd4b41a",
				"parent_hash": "0x5ac8b4099a26e9331a015f8437feadf56fa7fb447e8183aa1bdb3bf541a2cbb",
				"block_number": 48719,
				"new_root": "0x624f0f3cf2fbbd5951b0d90e4e1fc858f3d77cf34303781fcf3e4dc3afaf666",
				"timestamp": 1710445796,
				"sequencer_address": "0x1176a1bd84444c89232ec27754698e5d2e7e1a7f1539f12027f28b23ec9f3d8",
				"l1_gas_price": {
						"price_in_fri": "0x1d1a94a20000",
						"price_in_wei": "0x4a817c800"
				},
				"starknet_version": "0.13.1",
				"transactions": [
						{
								"transaction_hash": "0x27a9a9bc927efc37658acfb9c27b1fc56e7cfcf7a30db1ecdd9820bb2dddf0c",
								"type": "INVOKE",
								"version": "0x1",
								"nonce": "0x62b",
								"max_fee": "0x271963aac565a",
								"sender_address": "0x42db30408353b25c5a0b3dd798bfe98eba08956786374e961cc5dbb9811ec6e",
								"signature": [
										"0x66f70855f35096c6aea45be365617bc70fca65808bb85332dd3c2f6f4a86071",
										"0x43c09cab9be65cdc59b29b8018c201c1eb42d51529cb9b5dd64859652bf827f"
								],
								"calldata": [
										"0x1",
										"0x517567ac7026ce129c950e6e113e437aa3c83716cd61481c6bb8c5057e6923e",
										"0xcaffbd1bd76bd7f24a3fa1d69d1b2588a86d1f9d2359b13f6a84b7e1cbd126",
										"0xa",
										"0x53616d706c654465706f7369745374617274",
										"0x8",
										"0x4",
										"0x18343d00000001",
										"0x1",
										"0x5",
										"0x469",
										"0x2",
										"0x1",
										"0x4eb"
								]
						}]
			}`)
)

func TestChainClient(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		type Call struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
			Id     uint              `json:"id,omitempty"`
		}

		call := Call{}
		errMarshal := json.Unmarshal(req, &call)
		if errMarshal == nil {
			switch call.Method {
			case "starknet_getBlockWithTxs":
				out = []byte(fmt.Sprintf(`{ %s }`, blockResponse))
			case "starknet_blockNumber":
				out = []byte(`{"result": 1}`)
			default:
				require.False(t, true, "unsupported RPC method %s", call.Method)
			}
		} else {
			// batch method
			var batchCall []Call
			errBatchMarshal := json.Unmarshal(req, &batchCall)
			assert.NoError(t, errBatchMarshal)

			out = []byte(fmt.Sprintf(`
				[
					{
						"jsonrpc": "2.0",
						"id": %d,
						%s
					}
				]`, batchCall[0].Id, blockResponse))

		}

		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	lggr := logger.Test(t)
	client, err := NewClient(chainID, mockServer.URL, "", lggr, &myTimeout)
	require.NoError(t, err)
	assert.Equal(t, myTimeout, client.defaultTimeout)

	t.Run("get BlockByHash", func(t *testing.T) {
		block, err := client.BlockByHash(context.TODO(), blockHash)
		require.NoError(t, err)
		assert.Equal(t, blockHash, block.BlockHash)
	})

	t.Run("get Batch", func(t *testing.T) {
		builder := NewBatchBuilder()
		builder.RequestBlockByHash(blockHash)

		results, err := client.Batch(context.TODO(), builder)
		require.NoError(t, err)

		assert.Equal(t, "starknet_getBlockWithTxs", results[0].Method)
		assert.Nil(t, results[0].Error)

		block, ok := results[0].Result.(*FinalizedBlock)
		assert.True(t, ok)
		assert.Equal(t, blockHash, block.BlockHash)
	})
}
