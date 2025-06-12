package smoke_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-testing-framework/lib/logging"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	tc "github.com/smartcontractkit/chainlink-starknet/integration-tests/testconfig"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
)

var (
	keepAlive bool
	decimals  = 9
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}

func TestOCRBasic(t *testing.T) {
	config, err := tc.GetConfig("Smoke", tc.OCR2)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("CHAINLINK_ENV_USER", *config.Common.User)
	require.NoError(t, err, "Could not set CHAINLINK_ENV_USER")
	err = os.Setenv("INTERNAL_DOCKER_REPO", *config.Common.InternalDockerRepo)
	require.NoError(t, err, "Could not set INTERNAL_DOCKER_REPO")

	logging.Init()
	//
	state, err := common.NewOCRv2State(t, "smoke-ocr2", &config)
	require.NoError(t, err, "Could not setup the ocrv2 state")

	// K8s specific config and cleanup
	if *config.Common.InsideK8s {
		t.Cleanup(func() {
			if err = actions.TeardownSuite(t, nil, state.Common.Env, state.ChainlinkNodesK8s, nil, zapcore.PanicLevel, nil); err != nil {
				state.TestConfig.L.Error().Err(err).Msg("Error tearing down environment")
			}
		})
	}
	state.DeployCluster()
	// Setting up G++ Client
	rpcURL := state.Common.RPCDetails.RPCL2Internal
	gppURL := state.TestConfig.TestConfig.Common.GauntletPlusPlusURL
	state.Clients.GauntletPPClient, err = gauntlet.NewStarknetGauntletPlusPlus(gppURL, rpcURL, state.Account.Account, state.Account.PrivateKey)
	require.NoError(t, err, "Setting up gauntlet++ should not fail")

	state.Clients.GauntletClient, err = gauntlet.NewStarknetGauntlet(fmt.Sprintf("%s/", utils.ProjectRoot))
	require.NoError(t, err, "Setting up gauntlet should not fail")
	err = state.Clients.GauntletClient.SetupNetwork(state.Common.RPCDetails.RPCL2External, state.Account.Account, state.Account.PrivateKey)
	require.NoError(t, err, "Setting up gauntlet network should not fail")
	err = state.DeployGauntletPP(0, 100000000000, decimals, "auto", 1, 1)
	require.NoError(t, err, "Deploying contracts should not fail")

	state.SetUpNodes()

	err = state.ValidateRounds(*config.OCR2.NumberOfRounds, false)
	require.NoError(t, err, "Validating round should not fail")
}
