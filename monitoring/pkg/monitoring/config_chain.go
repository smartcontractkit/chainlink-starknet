package monitoring

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

type StarknetConfig struct {
	rpcEndpoint               string
	rpcApiKey                 string
	networkName               string
	networkID                 string
	chainID                   string
	readTimeout               time.Duration
	pollInterval              time.Duration
	linkTokenAddress          string
	strkTokenAddress          *felt.Felt
	addressesToMonitorBalance []ParsedAddress
}

type ParsedAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

var _ relayMonitoring.ChainConfig = StarknetConfig{}

func (s StarknetConfig) GetRPCEndpoint() string         { return s.rpcEndpoint }
func (s StarknetConfig) GetRPCApiKey() string           { return s.rpcApiKey }
func (s StarknetConfig) GetNetworkName() string         { return s.networkName }
func (s StarknetConfig) GetNetworkID() string           { return s.networkID }
func (s StarknetConfig) GetChainID() string             { return s.chainID }
func (s StarknetConfig) GetReadTimeout() time.Duration  { return s.readTimeout }
func (s StarknetConfig) GetPollInterval() time.Duration { return s.pollInterval }
func (s StarknetConfig) GetLinkTokenAddress() string    { return s.linkTokenAddress }
func (s StarknetConfig) GetAddressesToMonitorBalance() []ParsedAddress {
	return s.addressesToMonitorBalance
}
func (s StarknetConfig) GetStrkTokenAddress() *felt.Felt { return s.strkTokenAddress }

func (s StarknetConfig) ToMapping() map[string]interface{} {
	return map[string]interface{}{
		"network_name": s.networkName,
		"network_id":   s.networkID,
		"chain_id":     s.chainID,
	}
}

func ParseStarknetConfig() (StarknetConfig, error) {
	cfg := StarknetConfig{}

	if err := parseEnvVars(&cfg); err != nil {
		return cfg, err
	}

	applyDefaults(&cfg)

	err := validateConfig(cfg)
	return cfg, err
}

func parseEnvVars(cfg *StarknetConfig) error {
	if value, isPresent := os.LookupEnv("STARKNET_RPC_ENDPOINT"); isPresent {
		cfg.rpcEndpoint = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_RPC_API_KEY"); isPresent {
		cfg.rpcApiKey = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_NETWORK_NAME"); isPresent {
		cfg.networkName = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_NETWORK_ID"); isPresent {
		cfg.networkID = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_CHAIN_ID"); isPresent {
		cfg.chainID = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_READ_TIMEOUT"); isPresent {
		readTimeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var STARKNET_READ_TIMEOUT, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.readTimeout = readTimeout
	}
	if value, isPresent := os.LookupEnv("STARKNET_POLL_INTERVAL"); isPresent {
		pollInterval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var STARKNET_POLL_INTERVAL, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.pollInterval = pollInterval
	}
	if value, isPresent := os.LookupEnv("STARKNET_LINK_TOKEN_ADDRESS"); isPresent {
		cfg.linkTokenAddress = value
	}
	if value, isPresent := os.LookupEnv("STARKNET_TRACK_BALANCE"); isPresent {
		var addrs []ParsedAddress
		err := json.Unmarshal([]byte(value), &addrs)
		if err != nil {
			return fmt.Errorf("failed to parse env var STARKNET_TRACK_BALANCE")
		}
		cfg.addressesToMonitorBalance = addrs
	}
	if value, isPresent := os.LookupEnv("STRK_TOKEN_ADDRESS"); isPresent {
		feltValue, err := starknetutils.HexToFelt(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var STRK_TOKEN_ADDRESS %w", err)
		}
		cfg.strkTokenAddress = feltValue
	}
	return nil
}

func validateConfig(cfg StarknetConfig) error {
	// Required config
	for envVarName, currentValue := range map[string]string{
		"STARKNET_RPC_ENDPOINT":       cfg.rpcEndpoint,
		"STARKNET_NETWORK_NAME":       cfg.networkName,
		"STARKNET_NETWORK_ID":         cfg.networkID,
		"STARKNET_CHAIN_ID":           cfg.chainID,
		"STARKNET_LINK_TOKEN_ADDRESS": cfg.linkTokenAddress,
	} {
		if currentValue == "" {
			return fmt.Errorf("'%s' env var is required", envVarName)
		}
	}
	// Validate URLs.
	for envVarName, currentValue := range map[string]string{
		"STARKNET_RPC_ENDPOINT": cfg.rpcEndpoint,
	} {
		if _, err := url.ParseRequestURI(currentValue); err != nil {
			return fmt.Errorf("%s='%s' is not a valid URL: %w", envVarName, currentValue, err)
		}
	}
	return nil
}

func applyDefaults(cfg *StarknetConfig) {
	if cfg.readTimeout == 0 {
		cfg.readTimeout = 2 * time.Second
	}
	if cfg.pollInterval == 0 {
		cfg.pollInterval = 5 * time.Second
	}
}
