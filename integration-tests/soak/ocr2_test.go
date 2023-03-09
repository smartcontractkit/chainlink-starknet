package soak_test

// revive:disable:dot-imports
import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
	"github.com/smartcontractkit/chainlink/integration-tests/actions"
	"github.com/stretchr/testify/require"
)

var (
	keepAlive     bool
	err           error
	testState     *common.Test
	decimals      = 9
	mockServerVal = 900000000
)

func init() {
	flag.BoolVar(&keepAlive, "keep-alive", false, "enable to keep the cluster alive")
}
func TestOCRSoak(t *testing.T) {
	testState = &common.Test{
		T: t,
	}
	testState.Common = common.New()
	testState.Common.Default(t)
	// Setting this to the root of the repo for cmd exec func for Gauntlet
	testState.Sg, err = gauntlet.NewStarknetGauntlet(fmt.Sprintf("%s/", utils.ProjectRoot))
	require.NoError(t, err, "Could not get a new gauntlet struct")
	testState.DeployCluster()
	if testState.Common.Env.WillUseRemoteRunner() {
		return // short circuit here if using a remote runner
	}
	require.NoError(t, err, "Deploying cluster should not fail")
	err = testState.Sg.SetupNetwork(testState.Common.L2RPCUrl)
	require.NoError(t, err, "Setting up network should not fail")
	time.Sleep(8 * time.Hour)
	err = testState.DeployGauntlet(-100000000000, 100000000000, decimals, "auto", 1, 1)
	require.NoError(t, err, "Deploying contracts should not fail")
	if !testState.Common.Testnet {
		testState.Devnet.AutoLoadState(testState.OCR2Client, testState.OCRAddr)
	}
	testState.SetUpNodes(mockServerVal)
	err = testState.ValidateRounds(99999999, true)
	require.NoError(t, err, "Validating round should not fail")
	err = actions.TeardownSuite(testState.T, testState.Common.Env, "./", testState.GetChainlinkNodes(), nil, nil)
	require.NoError(t, err)
}
