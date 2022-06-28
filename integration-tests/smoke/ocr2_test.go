package smoke_test

//revive:disable:dot-imports
import (
	"encoding/json"
	"math/big"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	it "github.com/smartcontractkit/chainlink-starknet/integration-tests"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
	gauntlet "github.com/smartcontractkit/chainlink-testing-framework/gauntlet"
	"github.com/smartcontractkit/chainlink-testing-framework/utils"
	"github.com/smartcontractkit/helmenv/environment"
)

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err           error
		nets          *blockchain.Networks
		cls           []client.Chainlink
		networkL1     blockchain.EVMClient
		networkL2     blockchain.EVMClient
		ocrDeployer   *it.OCRDeployer
		starkDeployer *it.StarkNetContractDeployer
		e             *environment.Environment
		// These are one of the the default addresses based on the seed we pass to starknet which is 123
		walletAddress           string = "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502"
		walletPrivKey           string = "0xc4da537c1651ddae44867db30d67b366"
		linkTokenAddress        string
		accessControllerAddress string
		ocrControllerAddress    string
		g                       *gauntlet.Gauntlet
		options                 gauntlet.ExecCommandOptions
		gr                      *it.GauntletResponse
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			os.Setenv("PRIVATE_KEY", walletPrivKey)
			os.Setenv("ACCOUNT", walletAddress)
			g, err = it.NewStarknetGauntlet()
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
			// Remove this when relay is finished with development
			g.Network = "local"
			options = gauntlet.ExecCommandOptions{
				ErrHandling:       []string{},
				CheckErrorsInRead: true,
			}
		})

		By("Deploying the environment", func() {
			e, err = environment.DeployOrLoadEnvironment(ops.DefaultStarkNETEnv())
			Expect(err).ShouldNot(HaveOccurred())
			err = e.ConnectAll()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Connecting to launched resources", func() {
			networkRegistry := blockchain.NewDefaultNetworkRegistry()
			networkRegistry.RegisterNetwork(
				"l2_starknet_dev",
				it.GetStarkNetClient,
				it.GetStarkNetURLs,
			)
			nets, err = networkRegistry.GetNetworks(e)
			Expect(err).ShouldNot(HaveOccurred())
			networkL1, err = nets.Get(0)
			Expect(err).ShouldNot(HaveOccurred())
			networkL2, err = nets.Get(1)
			Expect(err).ShouldNot(HaveOccurred())
			ocrDeployer, err = it.NewOCRDeployer(networkL1)
			Expect(err).ShouldNot(HaveOccurred())
			starkDeployer, err = it.NewStarkNetContractDeployer(networkL2)
			Expect(err).ShouldNot(HaveOccurred())
			nets.Default.ParallelTransactions(true)
		})
		By("Funding Chainlink nodes", func() {
			cls, err = client.ConnectChainlinkNodes(e)
			Expect(err).ShouldNot(HaveOccurred())
			err = actions.FundChainlinkNodes(cls, networkL1, big.NewFloat(3))
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Deploying L1 contracts", func() {
			err = ocrDeployer.Deploy()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Deploying L2 contracts", func() {
			err = starkDeployer.Deploy()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Deploying LINK token contract", func() {
			_, err := g.ExecCommand([]string{"ERC20:deploy", "--link"}, options)
			Expect(err).ShouldNot(HaveOccurred(), "LINK Contract deployment should not fail")
			gr, err = it.FetchGauntletJsonOutput()
			Expect(err).ShouldNot(HaveOccurred(), "Fetching LINK token gauntlet output should not fail")
			linkTokenAddress = gr.Responses[0].Contract
			os.Setenv("LINK", linkTokenAddress)
		})

		By("Deploying access controller contract", func() {
			_, err := g.ExecCommand([]string{"access_controller:deploy"}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Access controller contract deployment should not fail")
			gr, err = it.FetchGauntletJsonOutput()
			Expect(err).ShouldNot(HaveOccurred(), "Fetching access controller gauntlet output should not fail")
			accessControllerAddress = gr.Responses[0].Contract
			os.Setenv("BILLING_ACCESS_CONTROLLER", accessControllerAddress)
		})

		By("Deploying OCR2 controller contract", func() {
			_, err := g.ExecCommand([]string{"ocr2:deploy", "--minSubmissionValue=0", "--maxSubmissionValue=9999999999", "--decimals=18", "--name=auto"}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Access controller contract deployment should not fail")
			gr, err = it.FetchGauntletJsonOutput()
			Expect(err).ShouldNot(HaveOccurred(), "Fetching OCR controller gauntlet output should not fail")
			ocrControllerAddress = gr.Responses[0].Contract

		})

		By("Setting OCR2 billing", func() {
			_, err := g.ExecCommand([]string{"ocr2:set_billing", "--observationPaymentGjuels=0", "--transmissionPaymentGjuels=1", ocrControllerAddress}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR billing should not fail")
		})

		By("Setting the Config Details on OCR2 Contract", func() {
			config, err := it.LoadOCR2Config()
			Expect(err).ShouldNot(HaveOccurred(), "Loading OCR2 config should not fail")
			parsedConfig, err := json.Marshal(config)
			Expect(err).ShouldNot(HaveOccurred(), "Parsing OCR2 config should not fail ")
			_, err = g.ExecCommand([]string{"ocr2:set_config", "--input=" + string(parsedConfig), ocrControllerAddress}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR config details should not fail")
		})

	})

	Describe("with OCRv2 job", func() {
		It("works", func() {
		})
	})

	AfterEach(func() {
		By("Tearing down the environment", func() {
			err = actions.TeardownSuite(e, nets, utils.ProjectRoot, nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
