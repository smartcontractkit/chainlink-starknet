package starknet

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestPreConfirmedRPCCalls(t *testing.T) {
	t.Parallel()

	var lastBody string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		lastBody = string(req)

		type call struct {
			Method string `json:"method"`
		}
		var c call
		require.NoError(t, json.Unmarshal(req, &c))

		var out []byte
		switch c.Method {
		case "starknet_specVersion":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":"0.9.0"}`)
		case "starknet_getNonce":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`)
		case "starknet_call":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":["0x0"]}`)
		case "starknet_getEvents":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":{"events":[],"continuation_token":""}}`)
		case "starknet_estimateFee":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":[{"l1_gas_consumed":"0x1","l1_gas_price":"0x1","l2_gas_consumed":"0x1","l2_gas_price":"0x1","l1_data_gas_consumed":"0x0","l1_data_gas_price":"0x0","overall_fee":"0x1","unit":"FRI"}]}`)
		default:
			t.Fatalf("unsupported RPC method %s body=%s", c.Method, string(req))
		}

		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	ctx := context.Background()
	account, err := starknetutils.HexToFelt("0x123")
	require.NoError(t, err)

	t.Run("AccountNonce uses pre_confirmed", func(t *testing.T) {
		lastBody = ""
		_, err := client.AccountNonce(ctx, account)
		require.NoError(t, err)
		assert.Contains(t, lastBody, `"pre_confirmed"`)
	})

	t.Run("EstimateFeeAtPreConfirmed", func(t *testing.T) {
		lastBody = ""
		estimates, err := client.EstimateFeeAtPreConfirmed(ctx, []starknetrpc.BroadcastTxn{}, []starknetrpc.SimulationFlag{starknetrpc.SkipValidate})
		require.NoError(t, err)
		require.Len(t, estimates, 1)
		assert.Equal(t, starknetrpc.FriUnit, estimates[0].Unit)
		assert.Contains(t, lastBody, `"pre_confirmed"`)
	})

	t.Run("Call at pre_confirmed", func(t *testing.T) {
		lastBody = ""
		_, err := client.Call(ctx, starknetrpc.FunctionCall{
			ContractAddress:    account,
			EntryPointSelector: account,
		}, PreConfirmedBlockID())
		require.NoError(t, err)
		assert.Contains(t, lastBody, `"pre_confirmed"`)
	})

	t.Run("Events at pre_confirmed", func(t *testing.T) {
		lastBody = ""
		_, err := client.Events(ctx, starknetrpc.EventsInput{
			EventFilter: starknetrpc.EventFilter{
				FromBlock: PreConfirmedBlockID(),
				ToBlock:   LatestBlockID(),
				Address:   account,
			},
			ResultPageRequest: starknetrpc.ResultPageRequest{ChunkSize: 10},
		})
		require.NoError(t, err)
		assert.True(t, strings.Contains(lastBody, `"pre_confirmed"`) || strings.Contains(lastBody, `"latest"`))
	})
}

func TestCallContractUsesLatest(t *testing.T) {
	t.Parallel()

	var lastBody string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		lastBody = string(req)

		type call struct {
			Method string `json:"method"`
		}
		var c call
		require.NoError(t, json.Unmarshal(req, &c))

		var out []byte
		switch c.Method {
		case "starknet_specVersion":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":"0.9.0"}`)
		case "starknet_call":
			out = []byte(`{"jsonrpc":"2.0","id":1,"result":["0x0"]}`)
		default:
			t.Fatalf("unsupported RPC method %s", c.Method)
		}
		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	timeout := 5 * time.Second
	client, err := NewClient("SN_SEPOLIA", mockServer.URL, "", logger.Test(t), &timeout)
	require.NoError(t, err)

	account, err := starknetutils.HexToFelt("0x123")
	require.NoError(t, err)
	_, err = client.CallContract(context.Background(), CallOps{
		ContractAddress: account,
		Selector:        account,
	})
	require.NoError(t, err)
	assert.Contains(t, lastBody, `"latest"`)
	assert.NotContains(t, lastBody, `"pre_confirmed"`)
}
