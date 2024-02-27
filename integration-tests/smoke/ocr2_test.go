package smoke_test

import (
	"flag"
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
	"maps"
	"testing"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	"github.com/smartcontractkit/chainlink/integration-tests/docker/test_env"
	tc "github.com/smartcontractkit/chainlink/integration-tests/testconfig"
	"github.com/stretchr/testify/require"
)

var (
	keepAlive bool
	decimals  = 9
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}

func TestOCRBasicNew(t *testing.T) {
	for _, test := range []struct {
		name string
		env  map[string]string
	}{
		{name: "embeded"},
		//{name: "plugins", env: map[string]string{
		//	"CL_MEDIAN_CMD": "chainlink-feeds",
		//	"CL_SOLANA_CMD": "chainlink-solana",
		//}},
	} {
		config, err := tc.GetConfig("Smoke", tc.OCR2)
		if err != nil {
			t.Fatal(err)
		}
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			logging.Init()
			state, err := common.NewOCRv2State(t, "localnet", false, "test", &config)
			require.NoError(t, err, "Could not setup the ocrv2 state")
			if len(test.env) > 0 {
				state.Common.NodeOpts = append(state.Common.NodeOpts, func(n *test_env.ClNode) {
					if n.ContainerEnvs == nil {
						n.ContainerEnvs = map[string]string{}
					}
					maps.Copy(n.ContainerEnvs, test.env)
				})
			}
			state.DeployCluster()
			state.Sg, err = gauntlet.NewStarknetGauntlet(fmt.Sprintf("%s/", utils.ProjectRoot))
			err = state.Sg.SetupNetwork(state.Common.L2RPCUrl, state.Account, state.PrivateKey)
			require.NoError(t, err, "Setting up gauntlet network should not fail")
			err = state.DeployGauntlet(0, 100000000000, decimals, "auto", 1, 1)
			require.NoError(t, err, "Deploying contracts should not fail")

			state.SetUpNodes()

			err = state.ValidateRounds(10, false)
			require.NoError(t, err, "Validating round should not fail")
		})
	}
}
