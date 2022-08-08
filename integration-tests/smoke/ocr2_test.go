package smoke_test

//revive:disable:dot-imports
import (
	"encoding/json"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	starknet "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
)

// Default config for OCR2 for starknet
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
}

type ReportingPluginConfig struct {
	AlphaReportInfinite bool `json:"alphaReportInfinite"`
	AlphaReportPpb      int  `json:"alphaReportPpb"`
	AlphaAcceptInfinite bool `json:"alphaAcceptInfinite"`
	AlphaAcceptPpb      int  `json:"alphaAcceptPpb"`
	DeltaCNanoseconds   int  `json:"deltaCNanoseconds"`
}

// Loads and returns the default starknet gauntlet config
func LoadOCR2Config(nKeys []ctfClient.NodeKeysBundle) (*OCR2Config, error) {
	var offChainKeys []string
	var onChainKeys []string
	var peerIds []string
	var txKeys []string
	for _, key := range nKeys {
		offChainKeys = append(offChainKeys, strings.Replace(key.OCR2Key.Data.Attributes.OffChainPublicKey, "ocr2off_starknet_", "", 1))
		peerIds = append(peerIds, key.PeerID)
		txKeys = append(txKeys, key.TXKey.Data.ID)
		onChainKeys = append(onChainKeys, "0x"+strings.Replace(key.OCR2Key.Data.Attributes.OnChainPublicKey, "ocr2on_starknet_", "", 1))
	}

	var payload = &OCR2Config{
		F:             1,
		Signers:       onChainKeys,
		Transmitters:  txKeys,
		OnchainConfig: "",
		OffchainConfig: &OffchainConfig{
			DeltaProgressNanoseconds: 8000000000,
			DeltaResendNanoseconds:   30000000000,
			DeltaRoundNanoseconds:    3000000000,
			DeltaGraceNanoseconds:    500000000,
			DeltaStageNanoseconds:    20000000000,
			RMax:                     5,
			S:                        []int{1, 2},
			OffchainPublicKeys:       offChainKeys,
			PeerIds:                  peerIds,
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
		},
		OffchainConfigVersion: 2,
		Secret:                "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
	}

	return payload, nil
}

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err error
		// These are one of the the default addresses based on the seed we pass to starknet which is 123
		walletAddress           string = "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502"
		walletPrivKey           string = "0xc4da537c1651ddae44867db30d67b366"
		linkTokenAddress        string
		accessControllerAddress string
		ocrAddress              string
		sg                      *starknet.StarknetGauntlet
		t                       *starknet.Test
		nKeys                   []ctfClient.NodeKeysBundle
		nAccounts               []string
		gauntletPath            string = "../../packages-ts/starknet-gauntlet-cli/networks/"
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			os.Setenv("PRIVATE_KEY", walletPrivKey)
			os.Setenv("ACCOUNT", walletAddress)
			sg, err = starknet.NewStarknetGauntlet("../../")
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
			// Setting this to the root of the repo for cmd exec func for Gauntlet
		})

		By("Deploying the environment", func() {
			t = t.DeployCluster(5)
			Expect(err).ShouldNot(HaveOccurred(), "Deploying cluster should not fail")
			nKeys = t.GetNodeKeys()
			sg.SetupNetwork(t.GetStarkNetAddress(), gauntletPath)
		})

		By("Funding nodes", func() {
			for _, key := range nKeys {
				nAccount, err := sg.DeployAccountContract(100, key.TXKey.Data.ID)
				Expect(err).ShouldNot(HaveOccurred(), "Funding node should not fail")
				nAccounts = append(nAccounts, nAccount)
			}
			t.FundAccounts(nAccounts)
		})

		By("Deploying LINK token contract", func() {
			linkTokenAddress, err := sg.DeployLinkTokenContract()
			Expect(err).ShouldNot(HaveOccurred(), "LINK Contract deployment should not fail")
			os.Setenv("LINK", linkTokenAddress)
		})

		By("Deploying access controller contract", func() {
			accessControllerAddress, err = sg.DeployAccessControllerContract()
			Expect(err).ShouldNot(HaveOccurred(), "Access controller contract deployment should not fail")
			os.Setenv("BILLING_ACCESS_CONTROLLER", accessControllerAddress)
		})

		By("Deploying OCR2 contract", func() {
			ocrAddress, err = sg.DeployOCR2ControllerContract(0, 9999999999, 18, "auto", linkTokenAddress)
			Expect(err).ShouldNot(HaveOccurred(), "OCR contract deployment should not fail")
		})

		By("Setting OCR2 billing", func() {
			_, err = sg.SetOCRBilling(1, 1, ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR billing should not fail")
		})

		By("Setting the Config Details on OCR2 Contract", func() {
			cfg, err := LoadOCR2Config(nKeys)
			Expect(err).ShouldNot(HaveOccurred(), "Loading OCR config should not fail")
			parsedConfig, err := json.Marshal(cfg)
			Expect(err).ShouldNot(HaveOccurred(), "Parsing OCR config should not fail")
			_, err = sg.SetConfigDetails(nKeys[1:], string(parsedConfig), ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR config should not fail")
		})

		By("Setting up bootstrap and oracle nodes", func() {
			t.CreateJobsForContract(ocrAddress)
		})

	})

	Describe("with OCRv2 job", func() {
		It("works", func() {
		})
	})

	AfterEach(func() {
		By("Tearing down the environment", func() {
			//	err = actions.TeardownSuite(t.Env, "./", t.GetChainlinkNodes(), nil, nil)
			//	Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
