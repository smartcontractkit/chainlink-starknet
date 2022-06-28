package integration_tests

import (
	"encoding/json"
	"io/ioutil"

	gauntlet "github.com/smartcontractkit/chainlink-testing-framework/gauntlet"
)

// Default response output for starknet gauntlet commands
type GauntletResponse struct {
	Responses []struct {
		Tx struct {
			Hash    string `json:"hash"`
			Address string `json:"address"`
			Status  string `json:"status"`
			Tx      struct {
				Address         string   `json:"address"`
				Code            string   `json:"code"`
				Result          []string `json:"result"`
				TransactionHash string   `json:"transaction_hash"`
			} `json:"tx"`
		} `json:"tx"`
		Contract string `json:"contract"`
	} `json:"responses"`
}

// Default config for OCR2 for starknet
type OCR2Config struct {
	F              int      `json:"f"`
	Signers        []string `json:"signers"`
	Transmitters   []string `json:"transmitters"`
	OnchainConfig  string   `json:"onchainConfig"`
	OffchainConfig struct {
		DeltaProgressNanoseconds int64    `json:"deltaProgressNanoseconds"`
		DeltaResendNanoseconds   int64    `json:"deltaResendNanoseconds"`
		DeltaRoundNanoseconds    int64    `json:"deltaRoundNanoseconds"`
		DeltaGraceNanoseconds    int      `json:"deltaGraceNanoseconds"`
		DeltaStageNanoseconds    int64    `json:"deltaStageNanoseconds"`
		RMax                     int      `json:"rMax"`
		S                        []int    `json:"s"`
		OffchainPublicKeys       []string `json:"offchainPublicKeys"`
		PeerIds                  []string `json:"peerIds"`
		ReportingPluginConfig    struct {
			AlphaReportInfinite bool `json:"alphaReportInfinite"`
			AlphaReportPpb      int  `json:"alphaReportPpb"`
			AlphaAcceptInfinite bool `json:"alphaAcceptInfinite"`
			AlphaAcceptPpb      int  `json:"alphaAcceptPpb"`
			DeltaCNanoseconds   int  `json:"deltaCNanoseconds"`
		} `json:"reportingPluginConfig"`
		MaxDurationQueryNanoseconds                        int `json:"maxDurationQueryNanoseconds"`
		MaxDurationObservationNanoseconds                  int `json:"maxDurationObservationNanoseconds"`
		MaxDurationReportNanoseconds                       int `json:"maxDurationReportNanoseconds"`
		MaxDurationShouldAcceptFinalizedReportNanoseconds  int `json:"maxDurationShouldAcceptFinalizedReportNanoseconds"`
		MaxDurationShouldTransmitAcceptedReportNanoseconds int `json:"maxDurationShouldTransmitAcceptedReportNanoseconds"`
	} `json:"offchainConfig"`
	OffchainConfigVersion int    `json:"offchainConfigVersion"`
	Secret                string `json:"secret"`
}

// Creates a default gauntlet config
func NewStarknetGauntlet() (*gauntlet.Gauntlet, error) {
	return gauntlet.NewGauntlet()
}

// Parse gauntlet json response that is generated after yarn gauntlet command execution
func FetchGauntletJsonOutput() (*GauntletResponse, error) {
	var payload = &GauntletResponse{}
	gauntletOutput, err := ioutil.ReadFile("../../report.json")
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(gauntletOutput, &payload)
	if err != nil {
		return payload, err
	}

	return payload, nil
}

// Loads and returns the default starknet gauntlet config
func LoadOCR2Config() (*OCR2Config, error) {
	var payload = &OCR2Config{}
	ocrConfig, err := ioutil.ReadFile("../config/ocr2Config.json")
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(ocrConfig, &payload)
	if err != nil {
		return payload, err
	}
	return payload, nil
}
