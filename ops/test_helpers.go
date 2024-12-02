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
	ConfigEncodingSecret                    string                 `json:"configEncodingSecret"`
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
		// todo: increase delta round but decrease delta stage
		DeltaProgress: "120s", // 120s
		DeltaResend:   "150s", // 150s
		DeltaRound:    "90s",  // 90s
		DeltaGrace:    "5s",   // 5s
		DeltaStage:    "20s",  // 20s
		RMax:          5,
		S:             []int{1, 1}, // Needs to array with length of transmitting nodes
		// OffchainPublicKeys:       offChainKeys, // user defined
		// PeerIDs:                  peerIds, // user defined
		ReportingPluginConfig: &ReportingPluginConfig{
			AlphaReportInfinite: false,
			AlphaReportPpb:      0,
			AlphaAcceptInfinite: false,
			AlphaAcceptPpb:      0,
			DeltaCNanoseconds:   1000000000,
		},
		MaxDurationQuery:                        "20s",
		MaxDurationObservation:                  "10s",
		MaxDurationReport:                       "20s",
		MaxDurationShouldAcceptFinalizedReport:  "20s",
		MaxDurationShouldTransmitAcceptedReport: "20s",
		// ConfigPublicKeys:                                   cfgKeys, // user defined
		// https://github.com/smartcontractkit/gauntlet-plus-plus/blob/5faf35e1d372e3ae5388c295554aa4f87bc0ece0/packages-starknet/operations-data-feeds/tests/fixtures/offchain-config.fixture.ts#L23
		ConfigEncodingSecret: "abandon ability able about above absent absorb abstract absurd abuse access accident",
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
