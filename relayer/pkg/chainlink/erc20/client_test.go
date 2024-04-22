package erc20

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
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestERC20Client(t *testing.T) {
	chainID := "SN_SEPOLIA"
	lggr := logger.Test(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := io.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		switch {
		case r.RequestURI == "/":
			type Request struct {
				Selector string `json:"entry_point_selector"`
			}
			type Call struct {
				Method string            `json:"method"`
				Params []json.RawMessage `json:"params"`
			}

			call := Call{}
			require.NoError(t, json.Unmarshal(req, &call))

			switch call.Method {
			case "starknet_call":
				raw := call.Params[0]
				reqdata := Request{}
				err := json.Unmarshal([]byte(raw), &reqdata)
				require.NoError(t, err)

				fmt.Printf("%v %v\n", reqdata.Selector, starknetutils.GetSelectorFromNameFelt("latest_transmission_details").String())
				switch reqdata.Selector {
				case starknetutils.GetSelectorFromNameFelt("decimals").String():
					// latest transmission details response
					out = []byte(`{"result":["0x1"]}`)
				case starknetutils.GetSelectorFromNameFelt("balance_of").String():
					// latest transmission details response
					out = []byte(`{"result":["0x0"]}`)
				default:
					require.False(t, true, "unsupported contract method %s", reqdata.Selector)
				}
			default:
				require.False(t, true, "unsupported RPC method")
			}
		default:
			require.False(t, true, "unsupported endpoint")
		}

		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	url := mockServer.URL
	duration := 10 * time.Second
	reader, err := starknet.NewClient(chainID, url, "", lggr, &duration)
	require.NoError(t, err)
	client, err := NewClient(reader, lggr, &felt.Zero)
	assert.NoError(t, err)

	// contractAddress, err := starknetutils.HexToFelt(ocr2ContractAddress)
	// require.NoError(t, err)

	t.Run("get balance", func(t *testing.T) {
		balance, err := client.BalanceOf(context.Background(), &felt.Zero)
		require.NoError(t, err)
		require.Equal(t, uint64(0), balance.Uint64())
		// require.Equal(t, new(big.Int), balance)
	})

	t.Run("get decimals", func(t *testing.T) {
		decimals, err := client.Decimals(context.Background())
		require.NoError(t, err)
		require.Equal(t, uint64(1), decimals.Uint64())
	})

}
