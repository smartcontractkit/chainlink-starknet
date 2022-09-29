package smoke_test

// revive:disable:dot-imports
import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	client "github.com/smartcontractkit/chainlink/integration-tests/client"
)

var (
	keepAlive bool
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err                     error
		linkTokenAddress        string
		accessControllerAddress string
		ocrAddress              string
		proxyAddress            string
		sg                      *gauntlet.StarknetGauntlet
		t                       *common.Test
		nAccounts               []string
		serviceKeyL1            = "Hardhat"
		serviceKeyL2            = "starknet-dev"
		serviceKeyChainlink     = "chainlink"
		chainName               = "starknet"
		chainId                 = gateway.GOERLI_ID
		cfg                     *common.Common
		decimals                = 9
		rpcRequestTimeout       = 10 * time.Second
		roundWaitTimeout        = 10 * time.Minute
		increasingCountMax      = 10
		mockServerVal           = 900000000
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			err = os.Setenv("PRIVATE_KEY", t.GetDefaultPrivateKey())
			err = os.Setenv("ACCOUNT", t.GetDefaultWalletAddress())
			Expect(err).ShouldNot(HaveOccurred(), "Setting env vars should not fail")

			// Setting this to the root of the repo for cmd exec func for Gauntlet
			sg, err = gauntlet.NewStarknetGauntlet("../../")
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
			sg.SetupNetwork(t.GetStarkNetAddress())
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
			linkTokenAddress, err = sg.DeployLinkTokenContract()
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
			ocrAddress, err = sg.DeployOCR2ControllerContract(-100000000000, 100000000000, decimals, "auto", linkTokenAddress)
			Expect(err).ShouldNot(HaveOccurred(), "OCR contract deployment should not fail")
		})

		By("Deploy proxy contract", func() {
			proxyAddress, err = sg.DeployOCR2ProxyContract(ocrAddress)
			Expect(err).ShouldNot(HaveOccurred(), "OCR2 proxy deployment should not fail")
		})

		By("Fund OCR2 contract", func() {
			_, err = sg.MintLinkToken(linkTokenAddress, ocrAddress, "100000000000000000000")
			Expect(err).ShouldNot(HaveOccurred(), "Funding OCR2 contract should not fail")
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
			_, err = sg.SetConfigDetails(string(parsedConfig), ocrAddress)
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

	Describe("with OCRv2 job", func() {
		It("works", func() {
			ctx := context.Background() // context background used because timeout handeld by requestTimeout param
			lggr := logger.Nop()
			url := t.Env.URLs[serviceKeyL2][0]

			// build client
			reader, err := starknet.NewClient(chainId, url, lggr, &rpcRequestTimeout)
			Expect(err).ShouldNot(HaveOccurred(), "Creating starknet client should not fail")
			client, err := ocr2.NewClient(reader, lggr)
			Expect(err).ShouldNot(HaveOccurred(), "Creating ocr2 client should not fail")

			// validate balance in aggregator
			resLINK, err := reader.CallContract(ctx, starknet.CallOps{
				ContractAddress: linkTokenAddress,
				Selector:        "balanceOf",
				Calldata:        []string{caigo.HexToBN(ocrAddress).String()},
			})
			Expect(err).ShouldNot(HaveOccurred(), "Reader balance from LINK contract should not fail")
			resAgg, err := reader.CallContract(ctx, starknet.CallOps{
				ContractAddress: ocrAddress,
				Selector:        "link_available_for_payment",
			})
			Expect(err).ShouldNot(HaveOccurred(), "Reader balance from LINK contract should not fail")
			balLINK, _ := new(big.Int).SetString(resLINK[0], 0)
			balAgg, _ := new(big.Int).SetString(resAgg[0], 0)
			Expect(balLINK.Cmp(big.NewInt(0)) == 1).To(BeTrue(), "Aggregator should have non-zero balance")
			Expect(balLINK.Cmp(balAgg) >= 0).To(BeTrue(), "Aggregator payment balance should be <= actual LINK balance")

			// assert new rounds are occuring
			details := ocr2.TransmissionDetails{}
			increasing := 0 // track number of increasing rounds
			var stuck bool
			stuckCount := 0

			// assert both positive and negative values have been seen
			var positive bool
			var negative bool

			for start := time.Now(); time.Since(start) < roundWaitTimeout; {
				// end condition: enough rounds have occured, and positive and negative answers have been seen
				if increasing >= increasingCountMax && positive && negative {
					break
				}

				// end condition: rounds have been stuck
				if stuck && stuckCount > 10 {
					log.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
					break
				}

				// once progression has reached halfway, change to negative values
				if increasing == increasingCountMax/2 {
					t.SetMockServerValue("", -1*mockServerVal)
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
				Expect(res.Digest == details.Digest).To(BeTrue(), "Config digest should not change")
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

			Expect(increasing >= increasingCountMax).To(BeTrue(), "Round + epochs should be increasing")
			Expect(positive).To(BeTrue(), "Positive value should have been submitted")
			Expect(negative).To(BeTrue(), "Negative value should have been submitted")
			Expect(stuck).To(BeFalse(), "Round + epochs should not be stuck")

			// Test proxy reading
			// TODO: would be good to test proxy switching underlying feeds
			roundDataRaw, err := reader.CallContract(ctx, starknet.CallOps{
				ContractAddress: proxyAddress,
				Selector:        "latest_round_data",
			})
			Expect(err).ShouldNot(HaveOccurred(), "Reading round data from proxy should not fail")
			Expect(len(roundDataRaw) == 5).Should(BeTrue(), "Round data from proxy should match expected size")
			value := starknet.HexToSignedBig(roundDataRaw[1]).Int64()
			Expect(value == int64(mockServerVal) || value == int64(-1*mockServerVal)).Should(BeTrue(), "Reading from proxy should return correct value")
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
