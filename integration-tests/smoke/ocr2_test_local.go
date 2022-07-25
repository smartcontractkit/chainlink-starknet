package smoke

//revive:disable:dot-imports
import (
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"

	it "github.com/smartcontractkit/chainlink-starknet/integration-tests"
)

var _ = Describe("StarkNET OCR suite @ocr1", func() {
	var (
	// err error

	// e *environment.Environment
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			it.DeployCluster(5)
		})

		By("Deploying the environment", func() {

		})

		By("Connecting to launched resources", func() {

		})
		By("Funding Chainlink nodes", func() {

		})

	})

	Describe("with OCRv2 job", func() {
		It("works", func() {
		})
	})

	AfterEach(func() {
		By("Tearing down the environment", func() {
			// err = actions.TeardownSuite(e, utils.ProjectRoot, nil, nil, nil)
			// Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
