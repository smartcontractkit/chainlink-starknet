package devnet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
