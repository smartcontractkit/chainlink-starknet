package soak_test

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-env/pkg/helm/remotetestrunner"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
)

func init() {
	logging.Init()
}

var (
	c                   *common.Common
	remoteContainerName = "remote-test-runner"
	// Required files / folders for the remote runner
	remoteFileList = []string{
		"../../ops",
		"../../package.json",
		"../../yarn.lock",
		"../../tsconfig.json",
		"../../tsconfig.base.json",
		"../../packages-ts",
		"../../contracts",
	}
)

// Run the OCR soak test defined in ./tests/ocr_test.go
func TestOCRSoak(t *testing.T) {
	activeEVMNetwork := blockchain.LoadNetworkFromEnvironment() // Environment currently being used to soak test on
	soakTestHelper(t, "@ocr", activeEVMNetwork)
}

// builds tests, launches environment, and triggers the soak test to run
func soakTestHelper(
	t *testing.T,
	testTag string,
	activeEVMNetwork *blockchain.EVMNetwork,
) {
	var err error
	c = common.New()
	c.Default()
	remoteRunnerValues := actions.BasicRunnerValuesSetup(
		testTag,
		c.Env.Cfg.Namespace,
		"./integration-tests/soak/tests",
	)
	envValues := map[string]interface{}{
		"test_log_level": "debug",
		"INSIDE_K8":      true,
		"TTL":            c.TTL,
		"NODE_COUNT":     c.NodeCount,
		"L2_RPC_URL":     c.L2RPCUrl,
		"PRIVATE_KEY":    c.PrivateKey,
		"ACCOUNT":        c.Account,
		"CL_VERSION":     c.CLVersion,
		"CL_IMAGE":       c.CLImage,
	}

	// Set env values
	for key, value := range envValues {
		remoteRunnerValues[key] = value
	}

	// Set evm network connection for remote runner
	for key, value := range activeEVMNetwork.ToMap() {
		remoteRunnerValues[key] = value
	}
	// Need to bump resources due to yarn using memory when running in parallel
	remoteRunnerWrapper := map[string]interface{}{
		"remote_test_runner": remoteRunnerValues,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "2000m",
				"memory": "1048Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "2000m",
				"memory": "1048Mi",
			},
		},
	}

	err = c.Env.
		AddHelm(remotetestrunner.New(remoteRunnerWrapper)).
		Run()
	require.NoError(t, err, "Error launching test environment")
	// Copying required files / folders to remote runner pod
	for _, file := range remoteFileList {
		_, _, _, err = c.Env.Client.CopyToPod(
			c.Env.Cfg.Namespace,
			file,
			fmt.Sprintf("%s/%s:/root/", c.Env.Cfg.Namespace, remoteContainerName),
			remoteContainerName)

		if err != nil {
			panic(err)
		}
	}

	err = actions.TriggerRemoteTest("../../", c.Env)
	require.NoError(t, err, "Error activating remote test")
}
