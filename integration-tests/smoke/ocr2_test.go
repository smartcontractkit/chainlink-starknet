package smoke_test

//revive:disable:dot-imports
import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	it "github.com/smartcontractkit/chainlink-starknet/integration-tests"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
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
		sg                      *it.StarknetGauntlet
		t                       *it.Test
		nKeys                   []ctfClient.NodeKeysBundle
		nAccounts               []string
		gauntletPath            string = "../../packages-ts/starknet-gauntlet-cli/networks/"
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			os.Setenv("PRIVATE_KEY", walletPrivKey)
			os.Setenv("ACCOUNT", walletAddress)
			sg, err = it.NewStarknetGauntlet("../../")
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
			_, err = sg.SetConfigDetails(nKeys[1:], ocrAddress)
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
