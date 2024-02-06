package soak_test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/ops/utils"

	"github.com/smartcontractkit/chainlink/integration-tests/actions"
)

var (
	keepAlive bool
	err       error
	testState *common.Test
	decimals  = 9
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
	require.NoError(t, err, "Deploying cluster should not fail")
	if testState.Common.Env.WillUseRemoteRunner() {
		return // short circuit here if using a remote runner
	}
	err = testState.Sg.SetupNetwork(testState.Common.L2RPCUrl)
	require.NoError(t, err, "Setting up network should not fail")
	err = testState.DeployGauntlet(0, 100000000000, decimals, "auto", 1, 1)
	require.NoError(t, err, "Deploying contracts should not fail")
	if !testState.Common.Testnet {
		testState.Devnet.AutoLoadState(testState.OCR2Client, testState.OCRAddr)
	}
	err = testState.ValidateRounds(99999999, true)
	require.NoError(t, err, "Validating round should not fail")
	t.Cleanup(func() {
		err = actions.TeardownSuite(t, testState.Common.Env, testState.Cc.ChainlinkNodes, nil, zapcore.ErrorLevel, nil, nil)
		require.NoError(t, err, "Error tearing down environment")
	})
}
