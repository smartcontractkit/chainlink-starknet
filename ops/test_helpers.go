package ops

import (
	"bytes"
	"math/big"
	"net/http"
	"os/exec"
	"testing"
	"time"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// devnet key derivation
	// https://github.com/Shard-Labs/starknet-devnet/blob/master/starknet_devnet/account.py
	DevnetClassHash, _ = new(big.Int).SetString("1803505466663265559571280894381905521939782500874858933595227108099796801620", 10)
	DevnetSalt         = big.NewInt(20)
)

// SetupLocalStarkNetNode sets up a local starknet node via cli, and returns the url
func SetupLocalStarkNetNode(t *testing.T) string {
	port := utils.MustRandomPort(t)
	url := "http://127.0.0.1:" + port
	cmd := exec.Command("starknet-devnet",
		"--seed", "0", // use same seed for testing
		"--port", port,
		"--lite-mode",
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
		res, err := http.Get(url + "/is_alive")
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

		keyBytes := caigotypes.HexToHash(k).Bytes()
		rawkeys = append(rawkeys, keyBytes)
	}
	return rawkeys
}

// OCR2Config Default config for OCR2 for starknet
type OCR2Config struct {
	F                     int             `json:"f"`
	Signers               []string        `json:"signers"`
	Transmitters          []string        `json:"transmitters"`
	OnchainConfig         string          `json:"onchainConfig"`
	OffchainConfig        *OffchainConfig `json:"offchainConfig"`
	OffchainConfigVersion int             `json:"offchainConfigVersion"`
	Secret                string          `json:"secret"`
}

type OffchainConfig struct {
	DeltaProgressNanoseconds                           int64                  `json:"deltaProgressNanoseconds"`
	DeltaResendNanoseconds                             int64                  `json:"deltaResendNanoseconds"`
	DeltaRoundNanoseconds                              int64                  `json:"deltaRoundNanoseconds"`
	DeltaGraceNanoseconds                              int                    `json:"deltaGraceNanoseconds"`
	DeltaStageNanoseconds                              int64                  `json:"deltaStageNanoseconds"`
	RMax                                               int                    `json:"rMax"`
	S                                                  []int                  `json:"s"`
	OffchainPublicKeys                                 []string               `json:"offchainPublicKeys"`
	PeerIds                                            []string               `json:"peerIds"`
	ReportingPluginConfig                              *ReportingPluginConfig `json:"reportingPluginConfig"`
	MaxDurationQueryNanoseconds                        int                    `json:"maxDurationQueryNanoseconds"`
	MaxDurationObservationNanoseconds                  int                    `json:"maxDurationObservationNanoseconds"`
	MaxDurationReportNanoseconds                       int                    `json:"maxDurationReportNanoseconds"`
	MaxDurationShouldAcceptFinalizedReportNanoseconds  int                    `json:"maxDurationShouldAcceptFinalizedReportNanoseconds"`
	MaxDurationShouldTransmitAcceptedReportNanoseconds int                    `json:"maxDurationShouldTransmitAcceptedReportNanoseconds"`
	ConfigPublicKeys                                   []string               `json:"configPublicKeys"`
}

type ReportingPluginConfig struct {
	AlphaReportInfinite bool `json:"alphaReportInfinite"`
	AlphaReportPpb      int  `json:"alphaReportPpb"`
	AlphaAcceptInfinite bool `json:"alphaAcceptInfinite"`
	AlphaAcceptPpb      int  `json:"alphaAcceptPpb"`
	DeltaCNanoseconds   int  `json:"deltaCNanoseconds"`
}

var TestOCR2Config = OCR2Config{
	F: 1,
	// Signers:       onChainKeys, // user defined
	// Transmitters:  txKeys, // user defined
	OnchainConfig: "",
	OffchainConfig: &OffchainConfig{
		DeltaProgressNanoseconds: 8000000000,
		DeltaResendNanoseconds:   30000000000,
		DeltaRoundNanoseconds:    3000000000,
		DeltaGraceNanoseconds:    500000000,
		DeltaStageNanoseconds:    20000000000,
		RMax:                     5,
		S:                        []int{1, 2},
		// OffchainPublicKeys:       offChainKeys, // user defined
		// PeerIds:                  peerIds, // user defined
		ReportingPluginConfig: &ReportingPluginConfig{
			AlphaReportInfinite: false,
			AlphaReportPpb:      0,
			AlphaAcceptInfinite: false,
			AlphaAcceptPpb:      0,
			DeltaCNanoseconds:   0,
		},
		MaxDurationQueryNanoseconds:                        0,
		MaxDurationObservationNanoseconds:                  1000000000,
		MaxDurationReportNanoseconds:                       200000000,
		MaxDurationShouldAcceptFinalizedReportNanoseconds:  200000000,
		MaxDurationShouldTransmitAcceptedReportNanoseconds: 200000000,
		// ConfigPublicKeys:                                   cfgKeys, // user defined
	},
	OffchainConfigVersion: 2,
	Secret:                "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
}

var TestOnKeys = []string{
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733",
}

var TestTxKeys = []string{
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736",
	"0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737",
}

var TestOffKeys = []string{
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852090",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852091",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852092",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852093",
}

var TestCfgKeys = []string{
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852094",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852095",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852096",
	"af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852097",
}
