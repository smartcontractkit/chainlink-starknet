package smoke_test

//revive:disable:dot-imports
import (
	"encoding/json"
	"os"

	"github.com/dontpanicdao/caigo/gateway"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	client "github.com/smartcontractkit/chainlink/integration-tests/client"
)

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err                     error
		linkTokenAddress        string
		accessControllerAddress string
		ocrAddress              string
		sg                      *starknet.StarknetGauntlet
		t                       *common.Test
		nAccounts               []string
		gauntletPath            = "../../packages-ts/starknet-gauntlet-cli/networks/"
		serviceKeyL1            = "Hardhat"
		serviceKeyL2            = "starknet-dev"
		serviceKeyChainlink     = "chainlink"
		chainName               = "starknet"
		chainId                 = gateway.GOERLI_ID
		cfg                     *common.Common
		decimals                = 9
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			err = os.Setenv("PRIVATE_KEY", t.GetDefaultPrivateKey())
			err = os.Setenv("ACCOUNT", t.GetDefaultWalletAddress())
			Expect(err).ShouldNot(HaveOccurred(), "Setting env vars should not fail")

			// Setting this to the root of the repo for cmd exec func for Gauntlet
			sg, err = starknet.NewStarknetGauntlet("../../")
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
		})

		By("Deploying the environment", func() {
			cfg = &common.Common{
				ChainName:           chainName,
				ChainId:             chainId,
				ServiceKeyChainlink: serviceKeyChainlink,
				ServiceKeyL1:        serviceKeyL1,
				ServiceKeyL2:        serviceKeyL2,
			}
			t = &common.Test{}
			t.DeployCluster(5, cfg)
			Expect(err).ShouldNot(HaveOccurred(), "Deploying cluster should not fail")
			devnet.SetL2RpcUrl(t.Env.URLs[serviceKeyL2][0])
			sg.SetupNetwork(t.GetStarkNetAddress(), gauntletPath)
		})

		By("Funding nodes", func() {
			for _, key := range t.GetNodeKeys() {
				Expect(key.TXKey.Data.Attributes.StarkKey).NotTo(Equal(""))
				nAccount, err := sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
				Expect(err).ShouldNot(HaveOccurred(), "Funding node should not fail")
				Expect(nAccount).To(Equal(key.TXKey.Data.Attributes.AccountAddr))
				nAccounts = append(nAccounts, nAccount)
			}
			err = devnet.FundAccounts(nAccounts)
			Expect(err).ShouldNot(HaveOccurred(), "Funding accounts should not fail")
		})

		By("Deploying LINK token contract", func() {
			linkTokenAddress, err := sg.DeployLinkTokenContract()
			Expect(err).ShouldNot(HaveOccurred(), "LINK Contract deployment should not fail")
			err = os.Setenv("LINK", linkTokenAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting env vars should not fail")

		})

		By("Deploying access controller contract", func() {
			accessControllerAddress, err = sg.DeployAccessControllerContract()
			Expect(err).ShouldNot(HaveOccurred(), "Access controller contract deployment should not fail")
			err = os.Setenv("BILLING_ACCESS_CONTROLLER", accessControllerAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting env vars should not fail")

		})

		By("Deploying OCR2 contract", func() {
			ocrAddress, err = sg.DeployOCR2ControllerContract(0, 100000000000, decimals, "auto", linkTokenAddress)
			Expect(err).ShouldNot(HaveOccurred(), "OCR contract deployment should not fail")
		})

		By("Setting OCR2 billing", func() {
			_, err = sg.SetOCRBilling(1, 1, ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR billing should not fail")
		})

		By("Setting the Config Details on OCR2 Contract", func() {
			cfg, err := t.LoadOCR2Config()
			Expect(err).ShouldNot(HaveOccurred(), "Loading OCR config should not fail")
			parsedConfig, err := json.Marshal(cfg)
			Expect(err).ShouldNot(HaveOccurred(), "Parsing OCR config should not fail")
			_, err = sg.SetConfigDetails(t.GetNodeKeys()[1:], string(parsedConfig), ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR config should not fail")
		})

		By("Setting up bootstrap and oracle nodes", func() {
			// TODO: validate juels per fee coin calculation
			juelsPerFeeCoinSource := ` 
			val [type = "bridge" name="bridge-cryptocompare" requestData=<{"fsym":"LINK", "tsyms":"ETH"}>]
			parse [type="jsonparse" path="ETH"]
			scale  [type="multiply" times=1000000000]
			val -> parse -> scale`

			observationSource := `
			val [type = "bridge" name="bridge-cryptocompare" requestData=<{"fsym":"LINK", "tsyms":"USD"}>]
			parse [type="jsonparse" path="USD"]
			scale [type="multiply" times=1000000000]
			val -> parse -> scale
			`

			t.SetBridgeTypeAttrs(&client.BridgeTypeAttributes{
				Name: "bridge-cryptocompare",
				URL:  "https://min-api.cryptocompare.com/data/price",
			})
			err = t.Common.CreateJobsForContract(t.GetChainlinkClient(), observationSource, juelsPerFeeCoinSource, ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Creating jobs should not fail")
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
