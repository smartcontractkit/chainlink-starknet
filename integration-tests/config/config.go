package config

var (
	starkTokenAddress = "0x04718f5a0fc34cc1af16a1cdee98ffb20c31f5cd61d6ab07201858f4287c938d"
)

type Config struct {
	ChainName         string
	ChainID           string
	StarkTokenAddress string
	L2RPCInternal     string
	TokenName         string
}

func SepoliaConfig() *Config {
	return &Config{
		ChainName:         "starknet",
		ChainID:           "SN_SEPOLIA",
		StarkTokenAddress: starkTokenAddress,
		// Will be overridden if set in toml
		L2RPCInternal: "https://starknet-sepolia.public.blastapi.io/rpc/v0_6",
	}
}

func DevnetConfig() *Config {
	return &Config{
		ChainName:         "starknet",
		ChainID:           "SN_GOERLI",
		StarkTokenAddress: starkTokenAddress,
		// Will be overridden if set in toml
		L2RPCInternal: "http://starknet-dev:5000",
		TokenName:     "FRI",
	}
}
