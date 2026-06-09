package starknet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestBlockTagParam(t *testing.T) {
	t.Parallel()

	blockNum := uint64(42)
	hash, err := starknetutils.HexToFelt("0x123")
	require.NoError(t, err)

	tests := []struct {
		name    string
		blockID starknetrpc.BlockID
		want    interface{}
		wantErr bool
	}{
		{
			name:    "pre_confirmed tag",
			blockID: PreConfirmedBlockID(),
			want:    BlockTagPreConfirmed,
		},
		{
			name:    "latest tag",
			blockID: starknetrpc.WithBlockTag("latest"),
			want:    "latest",
		},
		{
			name:    "pending tag",
			blockID: starknetrpc.WithBlockTag(BlockTagPending),
			want:    BlockTagPending,
		},
		{
			name:    "block number",
			blockID: starknetrpc.WithBlockNumber(blockNum),
			want:    map[string]uint64{"block_number": blockNum},
		},
		{
			name:    "block hash",
			blockID: starknetrpc.WithBlockHash(hash),
			want:    map[string]string{"block_hash": hash.String()},
		},
		{
			name:    "unsupported tag",
			blockID: starknetrpc.WithBlockTag("l1_accepted"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := blockTagParam(tt.blockID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEventsInputUsesPreConfirmed(t *testing.T) {
	t.Parallel()

	input := starknetrpc.EventsInput{
		EventFilter: starknetrpc.EventFilter{
			FromBlock: PreConfirmedBlockID(),
			ToBlock:   starknetrpc.WithBlockNumber(1),
		},
	}
	assert.True(t, eventsInputUsesPreConfirmed(input))

	input.FromBlock = starknetrpc.WithBlockNumber(1)
	input.ToBlock = starknetrpc.WithBlockNumber(2)
	assert.False(t, eventsInputUsesPreConfirmed(input))
}

func TestIsUnsupportedPreConfirmedBlockTagErr(t *testing.T) {
	t.Parallel()

	assert.False(t, isUnsupportedPreConfirmedBlockTagErr(nil))
	assert.True(t, isUnsupportedPreConfirmedBlockTagErr(
		fmt.Errorf("Invalid block ID: unknown variant `pre_confirmed`, expected `latest` or `pending`"),
	))
	assert.False(t, isUnsupportedPreConfirmedBlockTagErr(fmt.Errorf("connection reset")))
}

func TestHexStringsToFelts(t *testing.T) {
	t.Parallel()

	felts, err := hexStringsToFelts([]string{"0x1", "0x2"})
	require.NoError(t, err)
	require.Len(t, felts, 2)
	assert.Equal(t, "0x1", felts[0].String())
	assert.Equal(t, "0x2", felts[1].String())

	_, err = hexStringsToFelts([]string{"not-a-felt"})
	require.Error(t, err)
}

func TestPreConfirmedRPCCalls(t *testing.T) {
	contractAddress, err := starknetutils.HexToFelt("0x517567ac7026ce129c950e6e113e437aa3c83716cd61481c6bb8c5057e6923e")
	require.NoError(t, err)
	accountAddress, err := starknetutils.HexToFelt("0x42db30408353b25c5a0b3dd798bfe98eba08956786374e961cc5dbb9811ec6e")
	require.NoError(t, err)
	eventKey := starknetutils.GetSelectorFromNameFelt("ConfigSet")

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)

		type rpcRequest struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}

		var call rpcRequest
		require.NoError(t, json.Unmarshal(req, &call))

		switch call.Method {
		case "starknet_call":
			assert.Contains(t, string(req), `"latest"`)
			assert.NotContains(t, string(req), `"pre_confirmed"`)
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1","0x2"]}`))
		case "starknet_getNonce":
			assert.Contains(t, string(req), `"pre_confirmed"`)
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x5"}`))
		case "starknet_getEvents":
			assert.Contains(t, string(req), `"pre_confirmed"`)
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"events":[],"continuation_token":""}}`))
		case "starknet_estimateFee":
			assert.Contains(t, string(req), `"pre_confirmed"`)
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"l1_gas_consumed":"0x1","l1_gas_price":"0x1","l2_gas_consumed":"0x1","l2_gas_price":"0x1","l1_data_gas_consumed":"0x1","l1_data_gas_price":"0x1","overall_fee":"0x1","unit":"FRI"}]}`))
		case "starknet_chainId":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x534e5f5345504f4c4941"}`))
		default:
			t.Fatalf("unsupported RPC method %s body=%s", call.Method, string(req))
		}
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("CallContract uses latest", func(t *testing.T) {
		results, err := client.CallContract(ctx, CallOps{
			ContractAddress: contractAddress,
			Selector:        starknetutils.GetSelectorFromNameFelt("latest_round_data"),
		})
		require.NoError(t, err)
		require.Len(t, results, 2)
	})

	t.Run("AccountNonce uses pre_confirmed", func(t *testing.T) {
		nonce, err := client.AccountNonce(ctx, accountAddress)
		require.NoError(t, err)
		assert.Equal(t, "0x5", nonce.String())
	})

	t.Run("Events uses pre_confirmed", func(t *testing.T) {
		chunk, err := client.Events(ctx, starknetrpc.EventsInput{
			EventFilter: starknetrpc.EventFilter{
				FromBlock: PreConfirmedBlockID(),
				ToBlock:   PreConfirmedBlockID(),
				Address:   contractAddress,
				Keys:      [][]*felt.Felt{{eventKey}},
			},
			ResultPageRequest: starknetrpc.ResultPageRequest{
				ChunkSize: 10,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, chunk)
		assert.Empty(t, chunk.Events)
	})

	t.Run("EstimateFeeAtPreConfirmed", func(t *testing.T) {
		estimates, err := client.EstimateFeeAtPreConfirmed(ctx, []starknetrpc.BroadcastTxn{}, []starknetrpc.SimulationFlag{starknetrpc.SKIP_VALIDATE})
		require.NoError(t, err)
		require.Len(t, estimates, 1)
		assert.Equal(t, starknetrpc.FeePaymentUnit("FRI"), estimates[0].FeeUnit)
	})
}

func TestEventsAtBlockContinuationToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		assert.True(t, strings.Contains(string(req), `"continuation_token":"next-page"`))
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"events":[],"continuation_token":""}}`))
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	contractAddress, err := starknetutils.HexToFelt("0x517567ac7026ce129c950e6e113e437aa3c83716cd61481c6bb8c5057e6923e")
	require.NoError(t, err)

	chunk, err := client.eventsAtBlockOnce(context.Background(), starknetrpc.EventsInput{
		EventFilter: starknetrpc.EventFilter{
			FromBlock: PreConfirmedBlockID(),
			ToBlock:   PreConfirmedBlockID(),
			Address:   contractAddress,
		},
		ResultPageRequest: starknetrpc.ResultPageRequest{
			ChunkSize:         1,
			ContinuationToken: "next-page",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, chunk)
}

func TestPreConfirmedFallbackToPending(t *testing.T) {
	accountAddress, err := starknetutils.HexToFelt("0x42db30408353b25c5a0b3dd798bfe98eba08956786374e961cc5dbb9811ec6e")
	require.NoError(t, err)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(req), "starknet_getNonce") && strings.Contains(string(req), `"pre_confirmed"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":24,"message":"Invalid block ID: unknown variant pre_confirmed, expected latest or pending"}}`))
		case strings.Contains(string(req), "starknet_getNonce") && strings.Contains(string(req), `"pending"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x3"}`))
		case strings.Contains(string(req), "starknet_estimateFee") && strings.Contains(string(req), `"pre_confirmed"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":24,"message":"Invalid block ID: unknown variant pre_confirmed, expected latest or pending"}}`))
		case strings.Contains(string(req), "starknet_estimateFee") && strings.Contains(string(req), `"pending"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"l1_gas_consumed":"0x1","l1_gas_price":"0x1","l2_gas_consumed":"0x1","l2_gas_price":"0x1","l1_data_gas_consumed":"0x1","l1_data_gas_price":"0x1","overall_fee":"0x1","unit":"FRI"}]}`))
		case strings.Contains(string(req), "starknet_chainId"):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x534e5f5345504f4c4941"}`))
		default:
			t.Fatalf("unexpected request: %s", string(req))
		}
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	ctx := context.Background()

	nonce, err := client.AccountNonce(ctx, accountAddress)
	require.NoError(t, err)
	assert.Equal(t, "0x3", nonce.String())

	estimates, err := client.EstimateFeeAtPreConfirmed(ctx, []starknetrpc.BroadcastTxn{}, []starknetrpc.SimulationFlag{starknetrpc.SKIP_VALIDATE})
	require.NoError(t, err)
	require.Len(t, estimates, 1)
}

func TestEventsFallbackToPending(t *testing.T) {
	contractAddress, err := starknetutils.HexToFelt("0x517567ac7026ce129c950e6e113e437aa3c83716cd61481c6bb8c5057e6923e")
	require.NoError(t, err)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(req), "starknet_getEvents") && strings.Contains(string(req), `"pre_confirmed"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":24,"message":"Invalid block ID: unknown variant pre_confirmed, expected latest or pending"}}`))
		case strings.Contains(string(req), "starknet_getEvents") && strings.Contains(string(req), `"pending"`):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"events":[],"continuation_token":""}}`))
		case strings.Contains(string(req), "starknet_chainId"):
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x534e5f5345504f4c4941"}`))
		default:
			t.Fatalf("unexpected request: %s", string(req))
		}
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	chunk, err := client.Events(context.Background(), starknetrpc.EventsInput{
		EventFilter: starknetrpc.EventFilter{
			FromBlock: PreConfirmedBlockID(),
			ToBlock:   PreConfirmedBlockID(),
			Address:   contractAddress,
		},
		ResultPageRequest: starknetrpc.ResultPageRequest{
			ChunkSize: 10,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, chunk)
}
