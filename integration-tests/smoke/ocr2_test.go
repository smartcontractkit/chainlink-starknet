package smoke_test

//revive:disable:dot-imports
import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	it "github.com/smartcontractkit/chainlink-starknet/integration-tests"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink-testing-framework/gauntlet"
)

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err error
		// These are one of the the default addresses based on the seed we pass to starknet which is 123
		walletAddress           string = "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502"
		walletPrivKey           string = "0xc4da537c1651ddae44867db30d67b366"
		linkTokenAddress        string
		accessControllerAddress string
		ocrAddress              string
		g                       *gauntlet.Gauntlet
		options                 gauntlet.ExecCommandOptions
		gr                      *it.GauntletResponse
		t                       *it.Test
		nKeys                   []ctfClient.NodeKeysBundle
		nAccounts               []string
		gauntletPath            string = "../../packages-ts/starknet-gauntlet-cli/networks/"
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			os.Setenv("PRIVATE_KEY", walletPrivKey)
			os.Setenv("ACCOUNT", walletAddress)
			g, err = it.NewStarknetGauntlet()
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
			options = gauntlet.ExecCommandOptions{
				ErrHandling:       []string{},
				CheckErrorsInRead: true,
			}
			// Setting this to the root of the repo for cmd exec func for Gauntlet
			g.SetWorkingDir("../../")
		})

		By("Deploying the environment", func() {
			t = t.DeployCluster(5)
			Expect(err).ShouldNot(HaveOccurred(), "Deploying cluster should not fail")
			nKeys = t.GetNodeKeys()
			g.AddNetworkConfigVar("NODE_URL", t.GetStarkNetAddress())
			g.WriteNetworkConfigMap(gauntletPath)
		})

		By("Funding nodes", func() {
			for _, key := range nKeys {
				_, err := g.ExecCommand([]string{"account:deploy", "--salt=100", "--publicKey=" + key.TXKey.Data.ID}, options)
				Expect(err).ShouldNot(HaveOccurred(), "Deploying account should not fail")
				gr, err = it.FetchGauntletJsonOutput()
				Expect(err).ShouldNot(HaveOccurred(), "Fetching account address should not fail")
				nAccounts = append(nAccounts, gr.Responses[0].Contract)
			}
			t.FundAccounts(nAccounts)
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

		By("Deploying OCR2 contract", func() {
			_, err := g.ExecCommand([]string{"ocr2:deploy", "--minSubmissionValue=0", "--billingAccessController=" + accessControllerAddress, "--maxSubmissionValue=9999999999", "--decimals=18", "--name=auto", "--link=" + linkTokenAddress}, options)
			Expect(err).ShouldNot(HaveOccurred(), "OCR contract deployment should not fail")
			gr, err = it.FetchGauntletJsonOutput()
			Expect(err).ShouldNot(HaveOccurred(), "Fetching OCR gauntlet output should not fail")
			ocrAddress = gr.Responses[0].Contract
		})

		By("Setting OCR2 billing", func() {
			_, err := g.ExecCommand([]string{"ocr2:set_billing", "--observationPaymentGjuels=0", "--transmissionPaymentGjuels=1", ocrAddress}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR billing should not fail")
		})

		By("Setting the Config Details on OCR2 Contract", func() {
			// Starting from index 1 cause of bootstrap node
			config, err := it.LoadOCR2Config(nKeys[1:])
			Expect(err).ShouldNot(HaveOccurred(), "Loading OCR2 config should not fail")
			parsedConfig, err := json.Marshal(config)
			Expect(err).ShouldNot(HaveOccurred(), "Parsing OCR2 config should not fail ")
			_, err = g.ExecCommand([]string{"ocr2:set_config", "--input=" + string(parsedConfig), "--inspect", ocrAddress}, options)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR config details should not fail")
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
			// err = actions.TeardownSuite(it.Env, nil, utils.ProjectRoot, nil, nil)
			// Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
