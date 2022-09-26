package soak_test

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/remotetestrunner"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	networks "github.com/smartcontractkit/chainlink/integration-tests"
)

func init() {
	logging.Init()
}

var (
	baseEnvironmentConfig = &environment.Config{
		TTL: time.Hour * 720, // 30 days,
	}
	remoteContainerName = "remote-test-runner"
	remoteFileList      = []string{"../../ops", "../../package.json", "../../package-lock.json", "../../yarn.lock", "../../tsconfig.json", "../../tsconfig.base.json", "../../packages-ts", "../../contracts"}
)

// Run the OCR soak test defined in ./tests/ocr_test.go
func TestOCRSoak(t *testing.T) {
	activeEVMNetwork := networks.GeneralEVM() // Environment currently being used to soak test on

	baseEnvironmentConfig.NamespacePrefix = fmt.Sprintf(
		"chainlink-soak-ocr-starknet-%s",
		strings.ReplaceAll(strings.ToLower(activeEVMNetwork.Name), " ", "-"),
	)

	soakTestHelper(t, "@soak", activeEVMNetwork)
}

// builds tests, launches environment, and triggers the soak test to run
func soakTestHelper(
	t *testing.T,
	testTag string,
	activeEVMNetwork *blockchain.EVMNetwork,
) {
	exeFile, exeFileSize, err := actions.BuildGoTests("./", "../soak/tests", "../../")
	require.NoError(t, err, "Error building go tests")
	clConfig := map[string]interface{}{
		"replicas": 5,
		"env":      common.GetDefaultCoreConfig(),
	}

	testEnvironment := common.GetDefaultEnvSetup(baseEnvironmentConfig, clConfig)
	remoteRunnerValues := map[string]interface{}{
		"test_name":      testTag,
		"env_namespace":  testEnvironment.Cfg.Namespace,
		"test_file_size": fmt.Sprint(exeFileSize),
		"test_log_level": "debug",
		"INSIDE_K8":      true,
	}
	// Set evm network connection for remote runner
	for key, value := range activeEVMNetwork.ToMap() {
		remoteRunnerValues[key] = value
	}
	remoteRunnerWrapper := map[string]interface{}{"remote_test_runner": remoteRunnerValues}

	err = testEnvironment.
		AddHelm(remotetestrunner.New(remoteRunnerWrapper)).
		Run()

	// Copying required files / folders to pod
	for _, file := range remoteFileList {
		_, _, _, err = testEnvironment.Client.CopyToPod(
			testEnvironment.Cfg.Namespace,
			file,
			fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, remoteContainerName),
			remoteContainerName)

		if err != nil {
			panic(err)
		}
	}

	require.NoError(t, err, "Error launching test environment")
	err = actions.TriggerRemoteTest(exeFile, testEnvironment)
	require.NoError(t, err, "Error activating remote test")
}
