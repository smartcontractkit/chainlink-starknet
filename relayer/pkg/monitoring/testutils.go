package monitoring

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

// Generators

func generateChainConfig() StarknetConfig {
	return StarknetConfig{
		RPCEndpoint:  "http://starknet:9999",
		NetworkName:  "starknet-mainnet-beta",
		NetworkID:    "1",
		ChainID:      "starknet-mainnet-beta",
		ReadTimeout:  100 * time.Millisecond,
		PollInterval: time.Duration(1+rand.Intn(5)) * time.Second,
	}
}

func generateFeedConfig() StarknetFeedConfig {
	coins := []string{"btc", "eth", "matic", "link", "avax", "ftt", "srm", "usdc", "sol", "ray"}
	coin := coins[rand.Intn(len(coins))]
	contract := generatePublicKey()
	return StarknetFeedConfig{
		Name:            fmt.Sprintf("%s / usd", coin),
		Path:            fmt.Sprintf("%s-usd", coin),
		Symbol:          "$",
		HeartbeatSec:    1,
		ContractType:    "ocr2",
		ContractStatus:  "status",
		ContractAddress: contract,
	}
}

func generatePublicKey() string {
	arr := generate32ByteArr()
	return string(arr[:])
}

func generate32ByteArr() [32]byte {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		panic("unable to generate [32]byte from rand")
	}
	var out [32]byte
	copy(out[:], buf[:32])
	return out
}

func newNullLogger() logger.Logger {
	return logger.Nop()
}

// This utilities are used primarely in tests but are present in the monitoring package because they are not inside a file ending in _test.go.
// This is done in order to expose NewRandomDataReader for use in cmd/monitoring.
// The following code is added to comply with the "unused" linter:
var (
	_ = generateChainConfig()
	_ = generateFeedConfig()
	_ = generatePublicKey()
	_ = generate32ByteArr()
	_ = newNullLogger()
)
