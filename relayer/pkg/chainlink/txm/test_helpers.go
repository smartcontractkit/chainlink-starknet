package txm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/smartcontractkit/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DevnetAccount holds seed=0 predeployed account details from starknet-devnet-rs.
type DevnetAccount struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

// FetchDevnetAccounts returns predeployed accounts via devnet_getPredeployedAccounts.
// starknet-devnet-rs 0.8+ removed the legacy HTTP /predeployed_accounts endpoint.
func FetchDevnetAccounts(baseURL string) ([]DevnetAccount, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	payload := []byte(`{"jsonrpc":"2.0","id":1,"method":"devnet_getPredeployedAccounts","params":{}}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/rpc", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Result []DevnetAccount `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("devnet_getPredeployedAccounts: %s", rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// DevnetMint funds an account via devnet_mint JSON-RPC.
// starknet-devnet-rs 0.8+ removed the legacy HTTP POST /mint endpoint.
func DevnetMint(baseURL, address string, amount uint64, unit string) (string, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	params, err := json.Marshal(map[string]any{
		"address": address,
		"amount":  amount,
		"unit":    unit,
	})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "devnet_mint",
		"params":  json.RawMessage(params),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/rpc", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("devnet_mint (%s %d to %s): %s", unit, amount, address, rpcResp.Error.Message)
	}
	return string(rpcResp.Result), nil
}

var (
	// seed = 0 keys for starknet-devnet
	PrivateKeys0Seed = []string{
		"0xe3e70682c2094cac629f6fbed82c07cd",
		"0xf728b4fa42485e3a0a5d2f346baa9455",
		"0xeb1167b367a9c3787c65c1e582e2e662",
		"0xf7c1bd874da5e709d4713d60c8a70639",
		"0xe443df789558867f5ba91faf7a024204",
		"0x23a7711a8133287637ebdcd9e87a1613",
		"0x1846d424c17c627923c6612f48268673",
		"0xfcbd04c340212ef7cca5a5a19e4d6e3c",
		"0xb4862b21fb97d43588561712e8e5216a",
		"0x259f4329e6f4590b9a164106cf6a659e",
	}
)

// SetupLocalStarknetNode sets up a local starknet node via cli, and returns the url
func SetupLocalStarknetNode(t *testing.T) string {
	ctx := t.Context()
	port := strconv.Itoa(freeport.GetOne(t))
	url := "http://127.0.0.1:" + port
	cmd := exec.Command("starknet-devnet", // nolint:noctx
		"--seed", "0", // use same seed for testing
		"--port", port,
	)
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		assert.NoError(t, cmd.Process.Kill())
		if err2 := cmd.Wait(); assert.Error(t, err2) {
			if !assert.Contains(t, err2.Error(), "signal: killed", cmd.ProcessState.String()) {
				t.Log("starknet-devnet stderr:", stdErr.String())
			}
		}
		t.Log("starknet-devnet server closed")
	})

	// Wait for api server to boot
	var ready bool
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/is_alive", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != 200 {
			t.Logf("API server not ready yet (attempt %d)\n", i+1)
			continue
		}
		ready = true
		t.Logf("API server ready at %s\n", url)
		break
	}
	require.True(t, ready)
	return url
}

func TestKeys(t *testing.T, count int) (rawkeys [][]byte) {
	require.True(t, len(PrivateKeys0Seed) >= count, "requested more keys than available")
	for i, k := range PrivateKeys0Seed {
		// max number of keys to generate
		if i >= count {
			break
		}
		f, _ := starknetutils.HexToFelt(k)
		keyBytes := f.Bytes()
		rawkeys = append(rawkeys, keyBytes[:])
	}
	return rawkeys
}
