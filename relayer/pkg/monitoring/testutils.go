package monitoring

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

func newNullLogger() logger.Logger {
	return logger.Nop()
}

func generateChainConfig() StarknetConfig {
	return StarknetConfig{
		rpcEndpoint:      "http://starknet/6969",
		networkName:      "devnet",
		networkID:        "1",
		chainID:          "devnet",
		readTimeout:      100 * time.Millisecond,
		pollInterval:     1 * time.Second,
		linkTokenAddress: generateAddr(),
	}
}

func generateFeedConfig() StarknetFeedConfig {
	coins := []string{"btc", "eth", "matic", "link", "avax", "ftt", "srm", "usdc", "sol", "ray"}
	coin := coins[rand.Intn(len(coins))]

	return StarknetFeedConfig{
		Name:           fmt.Sprintf("%s / usd", coin),
		Path:           fmt.Sprintf("%s-usd", coin),
		Symbol:         "$",
		HeartbeatSec:   1,
		ContractType:   "ocr2",
		ContractStatus: "live",

		MultiplyRaw: "1000000",
		Multiply:    big.NewInt(1000000),

		ContractAddress: generateAddr(),
		ProxyAddress:    generateAddr(),
	}
}

func generateAddr() string {
	pool := []string{
		"0x0398eca85a333bc5de78f87d70d26f6e1f2438da6d163424b20f6190d3c38a21",
		"0x057605d472e1478b66396d8abec8f6c58348d9278d25049d9d73dafab40cde0c",
		"0x18e693006a3dc4db5adf7812c2e4ab8d7729707fcb3c439de0939f39de8d2b",
		"0x05c96456a9d58aa45e997182050f07a0649638a8f1d955935b42b6898d99e63d",
		"0x0358549759856b585a7b74ce5462e0ec0e56dbcc8fc729255150da2b62b702a6",
		"0x025d26785bc488193674b4e504f1ea0fc0bc28b0b92b7ce3e4b63ea5514bc3ab",
	}
	return pool[rand.Intn(len(pool))]
}

var (
	_ = generateChainConfig()
	_ = generateFeedConfig()
	_ = generateAddr()
)
