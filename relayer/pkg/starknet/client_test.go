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

	"github.com/smartcontractkit/caigo/gateway"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var (
	chainID = gateway.GOERLI_ID
	timeout = 10 * time.Second
)

func TestRPCClient(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		type Call struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}

		call := Call{}
		require.NoError(t, json.Unmarshal(req, &call))

		switch call.Method {
		case "starknet_chainId":
			id := caigotypes.BigToHex(caigotypes.UTF8StrToBig(chainID))
			out = []byte(fmt.Sprintf(`{"result": "%s"}`, id))
		case "starknet_blockNumber":
			out = []byte(`{"result": 1}`)
		default:
			require.False(t, true, "unsupported RPC method")
		}
		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	lggr := logger.Test(t)
	client, err := NewClient(chainID, mockServer.URL, lggr, &timeout)
	require.NoError(t, err)
	assert.Equal(t, timeout, client.defaultTimeout)

	t.Run("get chain id", func(t *testing.T) {
		// TODO: mock the chainID query
		id, err := client.Provider.ChainID(context.Background())
		require.NoError(t, err)
		assert.Equal(t, chainID, id)
	})

	t.Run("get block height", func(t *testing.T) {
		blockNum, err := client.LatestBlockHeight(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), blockNum)
	})
}
