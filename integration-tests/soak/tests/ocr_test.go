package soak_test

// revive:disable:dot-imports
import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dontpanicdao/caigo/gateway"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"
	client "github.com/smartcontractkit/chainlink/integration-tests/client"
	"math/big"
	"math/rand"
	"os"
	"time"
)

var (
	keepAlive bool
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err              error
		linkTokenAddress string
		//accessControllerAddress string
		ocrAddress string
		t          *common.Test
		//nAccounts               []string
		serviceKeyL1        = "Hardhat"
		serviceKeyL2        = "starknet-dev"
		serviceKeyChainlink = "chainlink"
		chainName           = "starknet"
		chainId             = gateway.GOERLI_ID
		cfg                 *common.Common
		decimals            = 9
		mockServerVal       = 900000000
		rpcRequestTimeout   = 10 * time.Second
		roundWaitTimeout    = 10 * time.Minute
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			t = &common.Test{}
			// Setting this to the root of the repo for cmd exec func for Gauntlet
			sg, err := ops.NewStarknetGauntlet("/root/")
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
			t.Sg = sg
			err = os.Setenv("PRIVATE_KEY", t.GetDefaultPrivateKey())
			err = os.Setenv("ACCOUNT", t.GetDefaultWalletAddress())
			Expect(err).ShouldNot(HaveOccurred(), "Setting env vars should not fail")
		})

		By("Deploying the environment", func() {
			cfg = &common.Common{
				ChainName:           chainName,
				ChainId:             chainId,
				ServiceKeyChainlink: serviceKeyChainlink,
				ServiceKeyL1:        serviceKeyL1,
				ServiceKeyL2:        serviceKeyL2,
			}

			t.DeployCluster(5, cfg)
			Expect(err).ShouldNot(HaveOccurred(), "Deploying cluster should not fail")
			devnet.SetL2RpcUrl(t.Env.URLs[serviceKeyL2][1])
			t.Sg.SetupNetwork(t.GetStarkNetAddressRemote())
		})

		By("Funding nodes", func() {
			err = t.FundNodes()
			Expect(err).ShouldNot(HaveOccurred(), "Funding nodes should not fail")
		})

		By("Deploying LINK token contract", func() {
			linkTokenAddress, err = t.DeployLinkToken()
			Expect(err).ShouldNot(HaveOccurred(), "LINK token should not fail")
		})

		By("Deploying access controller contract", func() {
			accessControllerAddress, err := t.DeployAccessController()
			Expect(err).ShouldNot(HaveOccurred(), "Deploying access controller should not fail"+accessControllerAddress)
		})

		By("Deploying OCR2 contract", func() {
			ocrAddress, err = t.Sg.DeployOCR2ControllerContract(-100000000000, 100000000000, decimals, "auto", linkTokenAddress)
			Expect(err).ShouldNot(HaveOccurred(), "OCR contract deployment should not fail")
		})

		By("Setting OCR2 billing", func() {
			_, err = t.Sg.SetOCRBilling(1, 1, ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR billing should not fail")
		})

		By("Setting the Config Details on OCR2 Contract", func() {
			cfg, err := t.LoadOCR2Config()
			Expect(err).ShouldNot(HaveOccurred(), "Loading OCR config should not fail")
			parsedConfig, err := json.Marshal(cfg)
			Expect(err).ShouldNot(HaveOccurred(), "Parsing OCR config should not fail")
			_, err = t.Sg.SetConfigDetails(string(parsedConfig), ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Setting OCR config should not fail")
		})

		By("Setting up bootstrap and oracle nodes", func() {
			observationSource := `
			val [type = "bridge" name="bridge-mockserver"]
			parse [type="jsonparse" path="data,result"]
			val -> parse
			`

			// TODO: validate juels per fee coin calculation
			juelsPerFeeCoinSource := ` 
			sum  [type="sum" values=<[451000]> ]
			sum`

			t.SetBridgeTypeAttrs(&client.BridgeTypeAttributes{
				Name: "bridge-mockserver",
				URL:  t.GetMockServerURL(),
			})
			t.SetMockServerValue("", mockServerVal)
			err = t.Common.CreateJobsForContract(t.GetChainlinkClient(), observationSource, juelsPerFeeCoinSource, ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "Creating jobs should not fail")
		})

	})

	Describe("with OCRv2 job @soak", func() {
		It("Soak test OCRv2", func() {
			lggr := logger.Nop()
			url := t.Env.URLs[serviceKeyL2][1]
			roundWaitTimeout = t.Env.Cfg.TTL
			log.Info().Msg(fmt.Sprintf("Starting run for:  %+v", roundWaitTimeout))

			// build client
			reader, err := starknet.NewClient(chainId, url, lggr, &rpcRequestTimeout)
			Expect(err).ShouldNot(HaveOccurred(), "Creating starknet client should not fail")
			client, err := ocr2.NewClient(reader, lggr)
			Expect(err).ShouldNot(HaveOccurred(), "Creating ocr2 client should not fail")

			// assert new rounds are occuring
			details := ocr2.TransmissionDetails{}
			increasing := 0 // track number of increasing rounds
			var stuck bool
			stuckCount := 0
			ctx := context.Background() // context background used because timeout handeld by requestTimeout param

			// assert both positive and negative values have been seen
			var positive bool
			var negative bool
			var sign = -1

			for start := time.Now(); time.Since(start) < roundWaitTimeout; {
				log.Info().Msg(fmt.Sprintf("Elapsed time: %s, Round wait: %s ", time.Since(start), roundWaitTimeout))
				rand.Seed(time.Now().UnixNano())
				sign *= -1
				var newValue = (rand.Intn(mockServerVal-0+1) + 0) * sign

				err = t.SetMockServerValue("", newValue)
				if err != nil {
					log.Error().Err(err)
				}

				// end condition: rounds have been stuck
				if stuck && stuckCount > 50 {
					log.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
					break
				}

				// try to fetch rounds
				time.Sleep(5 * time.Second)

				res, err := client.LatestTransmissionDetails(ctx, ocrAddress)
				if err != nil {
					log.Error().Err(err)
					continue
				}
				log.Info().Msg(fmt.Sprintf("Transmission Details: %+v", res))

				// continue if no changes
				if res.Epoch == 0 && res.Round == 0 {
					continue
				}

				// answer comparison (atleast see a positive and negative value once)
				ansCmp := res.LatestAnswer.Cmp(big.NewInt(0))
				positive = ansCmp == 1 || positive
				negative = ansCmp == -1 || negative

				// if changes from zero values set (should only initially)
				if res.Epoch > 0 && details.Epoch == 0 {
					Expect(res.Epoch > details.Epoch).To(BeTrue())
					Expect(res.Round >= details.Round).To(BeTrue())
					Expect(ansCmp != 0).To(BeTrue()) // assert changed from 0
					Expect(res.Digest != details.Digest).To(BeTrue())
					Expect(details.LatestTimestamp.Before(res.LatestTimestamp)).To(BeTrue())
					details = res
					continue
				}

				// check increasing rounds
				if !(res.Digest == details.Digest) {
					stuckCount += 1
					log.Error().Msg(fmt.Sprintf("Config digest should not change, expected %s got %s", details.Digest, res.Digest))
				}
				if (res.Epoch > details.Epoch || (res.Epoch == details.Epoch && res.Round > details.Round)) && details.LatestTimestamp.Before(res.LatestTimestamp) {
					increasing += 1
					stuck = false
					stuckCount = 0 // reset counter
					continue
				}

				// reach this point, answer has not changed
				stuckCount += 1
				if stuckCount > 5 {
					stuck = true
					increasing = 0
				}
			}
			log.Info().Msg(fmt.Sprintf("Reached the end of run"))
			Expect(increasing >= 0).To(BeTrue(), "Round + epochs should be increasing")
			Expect(negative).To(BeTrue(), "Positive value should have been submitted")
			Expect(positive).To(BeTrue(), "Positive value should have been submitted")
			Expect(stuck).To(BeFalse(), "Round + epochs should not be stuck")
		})
	})

	AfterEach(func() {
		// do not clean up if keepAlive flag is used
		if keepAlive {
			return
		}

		By("Tearing down the environment", func() {
			err = actions.TeardownSuite(t.Env, "./", t.GetChainlinkNodes(), nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
