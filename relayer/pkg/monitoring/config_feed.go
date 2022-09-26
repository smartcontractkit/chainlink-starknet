package monitoring

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
)

type StarknetFeedConfig struct {
	Name           string   `json:"name,omitempty"`
	Path           string   `json:"path,omitempty"`
	Symbol         string   `json:"symbol,omitempty"`
	HeartbeatSec   int64    `json:"heartbeat,omitempty"`
	ContractType   string   `json:"contract_type,omitempty"`
	ContractStatus string   `json:"status,omitempty"`
	MultiplyRaw    string   `json:"multiply,omitempty"`
	Multiply       *big.Int `json:"-"`

	ContractAddress string `json:"contract_address,omitempty"`
	ProxyAddress    string `json:"proxy_address,omitempty"`
}

var _ relayMonitoring.FeedConfig = StarknetFeedConfig{}

// GetID returns the state account's address as that uniquely
// identifies a feed on Starknet. In Starknet, a program is stateless and we
// use the same program for all feeds so we can't use the program
// account's address.
func (s StarknetFeedConfig) GetID() string {
	return s.ContractAddress
}

func (s StarknetFeedConfig) GetName() string {
	return s.Name
}

func (s StarknetFeedConfig) GetPath() string {
	return s.Path
}

func (s StarknetFeedConfig) GetFeedPath() string {
	return s.Path
}

func (s StarknetFeedConfig) GetSymbol() string {
	return s.Symbol
}

func (s StarknetFeedConfig) GetHeartbeatSec() int64 {
	return s.HeartbeatSec
}

func (s StarknetFeedConfig) GetContractType() string {
	return s.ContractType
}

func (s StarknetFeedConfig) GetContractStatus() string {
	return s.ContractStatus
}

func (s StarknetFeedConfig) GetMultiply() *big.Int {
	return s.Multiply
}

func (s StarknetFeedConfig) GetContractAddress() string {
	return s.ContractAddress
}

func (s StarknetFeedConfig) GetContractAddressBytes() []byte {
	return []byte(s.ContractAddress)
}

func (s StarknetFeedConfig) ToMapping() map[string]interface{} {
	return map[string]interface{}{
		"feed_name":        s.Name,
		"feed_path":        s.Path,
		"symbol":           s.Symbol,
		"heartbeat_sec":    int64(s.HeartbeatSec),
		"contract_type":    s.ContractType,
		"contract_status":  s.ContractStatus,
		"contract_address": []byte(s.ContractAddress),

		// These fields are legacy. They are required in the schema but they
		// should be set to a zero value for any other chain.
		"transmissions_account": []byte{},
		"state_account":         []byte{},
	}
}

func StarknetFeedsParser(buf io.ReadCloser) ([]relayMonitoring.FeedConfig, error) {
	rawFeeds := []StarknetFeedConfig{}
	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(&rawFeeds); err != nil {
		return nil, fmt.Errorf("unable to unmarshal feeds config data: %w", err)
	}
	feeds := make([]relayMonitoring.FeedConfig, len(rawFeeds))
	for i, rawFeed := range rawFeeds {
		feeds[i] = relayMonitoring.FeedConfig(rawFeed)
	}
	return feeds, nil
}
