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
	DeltaProgressNanoseconds                           string                 `json:"deltaProgress"`
	DeltaResendNanoseconds                             string                 `json:"deltaResend"`
	DeltaRoundNanoseconds                              string                 `json:"deltaRound"`
	DeltaGraceNanoseconds                              string                 `json:"deltaGrace"`
	DeltaStageNanoseconds                              string                 `json:"deltaStage"`
	RMax                                               int                    `json:"rMax"`
	S                                                  []int                  `json:"s"`
	OffchainPublicKeys                                 []string               `json:"offchainPublicKeys"`
	PeerIDs                                            []string               `json:"peerIds"`
	ReportingPluginConfig                              *ReportingPluginConfig `json:"reportingPluginConfig"`
	MaxDurationQueryNanoseconds                        string                 `json:"maxDurationQuery"`
	MaxDurationObservationNanoseconds                  string                 `json:"maxDurationObservation"`
	MaxDurationReportNanoseconds                       string                 `json:"maxDurationReport"`
	MaxDurationShouldAcceptFinalizedReportNanoseconds  string                 `json:"maxDurationShouldAcceptFinalizedReport"`
	MaxDurationShouldTransmitAcceptedReportNanoseconds string                 `json:"maxDurationShouldTransmitAcceptedReport"`
	ConfigPublicKeys                                   []string               `json:"configPublicKeys"`
	ConfigEncodingSecret                               string                 `json:"configEncodingSecret"`
	SignerSecret                                       string                 `json:"signerSecret"`
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
		DeltaProgressNanoseconds: "150000000000ns", // 120s
		DeltaResendNanoseconds:   "150000000000ns", // 150s
		DeltaRoundNanoseconds:    "90000000000ns",  // 90s
		DeltaGraceNanoseconds:    "5000000000ns",   // 5s
		DeltaStageNanoseconds:    "30000000000ns",  // 20s
		RMax:                     5,
		S:                        []int{1, 1}, // Needs to array with length of transmitting nodes
		// OffchainPublicKeys:       offChainKeys, // user defined
		// PeerIDs:                  peerIds, // user defined
		ReportingPluginConfig: &ReportingPluginConfig{
			AlphaReportInfinite: false,
			AlphaReportPpb:      0,
			AlphaAcceptInfinite: false,
			AlphaAcceptPpb:      0,
			DeltaCNanoseconds:   1000000000,
		},
		MaxDurationQueryNanoseconds:                        "2000000000ns",
		MaxDurationObservationNanoseconds:                  "1000000000ns",
		MaxDurationReportNanoseconds:                       "2000000000ns",
		MaxDurationShouldAcceptFinalizedReportNanoseconds:  "2000000000ns",
		MaxDurationShouldTransmitAcceptedReportNanoseconds: "2000000000ns",
		// ConfigPublicKeys:                                   cfgKeys, // user defined
		ConfigEncodingSecret: "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
		SignerSecret:         "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
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
