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
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var (
	myTimeout     = 100 * time.Second
	blockNumber   = 48719
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
	blockHashAndNumberResponse = fmt.Sprintf(`{"block_hash": "%s", "block_number": %d}`,
		"0x725407fcc3bd43e50884f50f1e0ef32aa9f814af3da475411934a7dbd4b41a",
		48719,
	)
	// hex-encoded value for "SN_SEPOLIA"
	chainIDHex = "0x534e5f5345504f4c4941"
)

func TestChainClient(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		type Call struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
			ID     uint              `json:"id,omitempty"`
		}

		call := Call{}
		errMarshal := json.Unmarshal(req, &call)
		if errMarshal == nil {
			switch call.Method {
			case "starknet_getBlockWithTxs":
				out = []byte(fmt.Sprintf(`{ %s }`, blockResponse))
			case "starknet_blockNumber":
				out = []byte(`{"result": 1}`)
			case "starknet_blockHashAndNumber":
				out = []byte(fmt.Sprintf(`{"result": %s}`, blockHashAndNumberResponse))
			case "starknet_chainId":
				out = []byte(fmt.Sprintf(`{"result": "%s"}`, chainIDHex))
			default:
				require.False(t, true, "unsupported RPC method %s", call.Method)
			}
		} else {
			// batch method
			var batchCall []Call
			errBatchMarshal := json.Unmarshal(req, &batchCall)
			assert.NoError(t, errBatchMarshal)

			// special case where we test chainID call
			if len(batchCall) == 1 {
				response := fmt.Sprintf(`
				[
					{ "jsonrpc": "2.0",
						"id": %d,
						"result": "%s"
					}
				]`, batchCall[0].ID, chainIDHex)
				out = []byte(response)
			} else {
				response := fmt.Sprintf(`
			[
				{ "jsonrpc": "2.0",
					"id": %d,
					"result": "%s"
				},
				{
					"jsonrpc": "2.0",
					"id": %d,
					%s
				},
				{
					"jsonrpc": "2.0",
					"id": %d,
					%s
				},
				{
					"jsonrpc": "2.0",
					"id": %d,
					"result": %s
				}
			]`, batchCall[0].ID, chainIDHex,
					batchCall[1].ID, blockResponse,
					batchCall[2].ID, blockResponse,
					batchCall[3].ID, blockHashAndNumberResponse,
				)

				out = []byte(response)
			}
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

	t.Run("get BlockByNumber", func(t *testing.T) {
		block, err := client.BlockByNumber(context.TODO(), uint64(blockNumber))
		require.NoError(t, err)
		assert.Equal(t, uint64(blockNumber), block.BlockNumber)
	})

	t.Run("get LatestBlockHashAndNumber", func(t *testing.T) {
		output, err := client.LatestBlockHashAndNumber(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, blockHash, output.BlockHash)
		assert.Equal(t, uint64(blockNumber), output.BlockNumber)
	})

	t.Run("get ChainID", func(t *testing.T) {
		output, err := client.ChainID(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, chainIDHex, output)
	})

	t.Run("get Batch", func(t *testing.T) {
		builder := NewBatchBuilder()
		builder.
			RequestChainID().
			RequestBlockByHash(blockHash).
			RequestBlockByNumber(uint64(blockNumber)).
			RequestLatestBlockHashAndNumber()

		results, err := client.Batch(context.TODO(), builder)
		require.NoError(t, err)

		assert.Equal(t, 4, len(results))

		t.Run("gets ChainID in Batch", func(t *testing.T) {
			assert.Equal(t, "starknet_chainId", results[0].Method)
			assert.Nil(t, results[0].Error)
			id, ok := results[0].Result.(*string)
			assert.True(t, ok)
			fmt.Println(id)
			assert.Equal(t, chainIDHex, *id)
		})

		t.Run("gets BlockByHash in Batch", func(t *testing.T) {
			assert.Equal(t, "starknet_getBlockWithTxs", results[1].Method)
			assert.Nil(t, results[1].Error)
			block, ok := results[1].Result.(*FinalizedBlock)
			assert.True(t, ok)
			assert.Equal(t, blockHash, block.BlockHash)
		})

		t.Run("gets BlockByNumber in Batch", func(t *testing.T) {
			assert.Equal(t, "starknet_getBlockWithTxs", results[2].Method)
			assert.Nil(t, results[2].Error)
			block, ok := results[2].Result.(*FinalizedBlock)
			assert.True(t, ok)
			assert.Equal(t, uint64(blockNumber), block.BlockNumber)
		})

		t.Run("gets LatestBlockHashAndNumber in Batch", func(t *testing.T) {
			assert.Equal(t, "starknet_blockHashAndNumber", results[3].Method)
			assert.Nil(t, results[3].Error)
			info, ok := results[3].Result.(*starknetrpc.BlockHashAndNumberOutput)
			assert.True(t, ok)
			assert.Equal(t, blockHash, info.BlockHash)
			assert.Equal(t, uint64(blockNumber), info.BlockNumber)
		})
	})
}
