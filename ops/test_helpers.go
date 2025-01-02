package ops

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
	DeltaProgress                           string                 `json:"deltaProgress"`
	DeltaResend                             string                 `json:"deltaResend"`
	DeltaRound                              string                 `json:"deltaRound"`
	DeltaGrace                              string                 `json:"deltaGrace"`
	DeltaStage                              string                 `json:"deltaStage"`
	RMax                                    int                    `json:"rMax"`
	S                                       []int                  `json:"s"`
	OffchainPublicKeys                      []string               `json:"offchainPublicKeys"`
	PeerIDs                                 []string               `json:"peerIds"`
	ReportingPluginConfig                   *ReportingPluginConfig `json:"reportingPluginConfig"`
	MaxDurationQuery                        string                 `json:"maxDurationQuery"`
	MaxDurationObservation                  string                 `json:"maxDurationObservation"`
	MaxDurationReport                       string                 `json:"maxDurationReport"`
	MaxDurationShouldAcceptFinalizedReport  string                 `json:"maxDurationShouldAcceptFinalizedReport"`
	MaxDurationShouldTransmitAcceptedReport string                 `json:"maxDurationShouldTransmitAcceptedReport"`
	ConfigPublicKeys                        []string               `json:"configPublicKeys"`
	SignerSecret string `json:"signerSecret"`
}

type ReportingPluginConfig struct {
	SourceFinalityDepth         int    `json:"sourceFinalityDepth"`
	DestFinalityDepth           int    `json:"destFinalityDepth"`
	MaxGasPrice                 string `json:"maxGasPrice"`
	RelativeBoostPerWaitHour    string `json:"relativeBoostPerWaitHour"`
	InflightCacheExpiry         string `json:"inflightCacheExpiry"`
	RootSnoozeTime              string `json:"rootSnoozeTime"`
	DestOptimisticConfirmations int    `json:"DestOptimisticConfirmations"`
	BatchGasLimit               int    `json:"batchGasLimit"`
}

var TestOCR2Config = OCR2Config{
	F: 1,
	// Signers:       onChainKeys, // user defined
	// Transmitters:  txKeys, // user defined
	OnchainConfig: "",
	// https://github.com/smartcontractkit/gauntlet-plus-plus/blob/main/packages-starknet/operations-data-feeds/tests/fixtures/offchain-config.fixture.ts
	OffchainConfig: &OffchainConfig{
		// todo: increase delta round but decrease delta stage
		DeltaProgress: "120000000000ns", // 120s
		DeltaResend:   "5000000000ns",   // 150s
		DeltaRound:    "60000000000ns",  // 90s
		DeltaGrace:    "5000000000ns",   // 5s
		DeltaStage:    "180000000000ns", // 20s
		RMax:          3,
		S:             []int{1, 1, 1, 0}, // Needs to array with length of transmitting nodes
		// OffchainPublicKeys:       offChainKeys, // user defined
		// PeerIDs:                  peerIds, // user defined
		ReportingPluginConfig: &ReportingPluginConfig{
			SourceFinalityDepth:         5,
			DestFinalityDepth:           15,
			MaxGasPrice:                 "300000000000",
			RelativeBoostPerWaitHour:    "300000000000",
			InflightCacheExpiry:         "3m0s",
			RootSnoozeTime:              "3m0s",
			DestOptimisticConfirmations: 1,
			BatchGasLimit:               1,
		},
		MaxDurationQuery:                        "100000000ns",
		MaxDurationObservation:                  "35000000000ns",
		MaxDurationReport:                       "10000000000ns",
		MaxDurationShouldAcceptFinalizedReport:  "5000000000ns",
		MaxDurationShouldTransmitAcceptedReport: "10000000000ns",
		// ConfigPublicKeys:                                   cfgKeys, // user defined
		// https://github.com/smartcontractkit/gauntlet-plus-plus/blob/5faf35e1d372e3ae5388c295554aa4f87bc0ece0/packages-starknet/operations-data-feeds/tests/fixtures/offchain-config.fixture.ts#L23
		SignerSecret: "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
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
