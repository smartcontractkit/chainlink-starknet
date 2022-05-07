package smoke_test

//revive:disable:dot-imports
import (
	"math/big"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	it "github.com/smartcontractkit/chainlink-starknet/integration-tests"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
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
	)

	BeforeEach(func() {
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
