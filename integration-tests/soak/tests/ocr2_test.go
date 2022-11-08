package soak_test

// revive:disable:dot-imports
import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	caigotypes "github.com/dontpanicdao/caigo/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"
)

var (
	keepAlive bool
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}

var _ = Describe("StarkNET OCR suite @ocr", func() {
	var (
		err           error
		t             *common.Test
		decimals      = 9
		mockServerVal = 900000000
	)

	BeforeEach(func() {
		By("Gauntlet preparation", func() {
			t = &common.Test{}
			t.Common = common.New()
			t.Common.Default()
			// Setting this to the root of the repo for cmd exec func for Gauntlet
			t.Sg, err = gauntlet.NewStarknetGauntlet("/root/")
			Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
		})

		By("Deploying the environment", func() {
			t.DeployCluster()
			Expect(err).ShouldNot(HaveOccurred(), "Deploying cluster should not fail")
			err = t.Sg.SetupNetwork(t.Common.L2RPCUrl)
			Expect(err).ShouldNot(HaveOccurred(), "Setting up network should not fail")
			err = t.DeployGauntlet(-100000000000, 100000000000, decimals, "auto", 1, 1)
			Expect(err).ShouldNot(HaveOccurred(), "Deploying contracts should not fail")
			if !t.Common.Testnet {
				t.Devnet.AutoLoadState(t.OCR2Client, t.OCRAddr)
			}
		})

		By("Setting up bootstrap and oracle nodes", func() {
			t.SetUpNodes(mockServerVal)
		})

	})

	Describe("with OCRv2 job @soak", func() {
		It("Soak test OCRv2", func() {
			log.Info().Msg(fmt.Sprintf("Starting run for:  %+v", t.Common.TTL))
			// assert new rounds are occurring
			details := ocr2.TransmissionDetails{}
			increasing := 0 // track number of increasing rounds
			var stuck bool
			stuckCount := 0
			ctx := context.Background() // context background used because timeout handled by requestTimeout param

			// assert both positive and negative values have been seen
			var positive bool
			var negative bool
			var sign = -1

			for start := time.Now(); time.Since(start) < t.Common.TTL; {
				log.Info().Msg(fmt.Sprintf("Elapsed time: %s, Round wait: %s ", time.Since(start), t.Common.TTL))
				rand.Seed(time.Now().UnixNano())
				sign *= -1
				var newValue = (rand.Intn(mockServerVal-0+1) + 0) * sign

				err = t.SetMockServerValue("", newValue)
				if err != nil {
					log.Error().Msg(fmt.Sprintf("Setting mock server value error: %+v", err))
				}

				// end condition: rounds have been stuck
				if stuck && stuckCount > 50 {
					log.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
					break
				}

				// try to fetch rounds
				time.Sleep(5 * time.Second)

				res, err := t.OCR2Client.LatestTransmissionDetails(ctx, caigotypes.HexToHash(t.OCRAddr))
				if err != nil {
					log.Error().Msg(fmt.Sprintf("Transmission Error: %+v", err))
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
			err = actions.TeardownSuite(t.Common.Env, "./", t.GetChainlinkNodes(), nil, nil)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
