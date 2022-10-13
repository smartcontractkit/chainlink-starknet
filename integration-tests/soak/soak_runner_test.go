package soak_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/remotetestrunner"
	"github.com/smartcontractkit/chainlink-testing-framework/actions"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
)

func init() {
	logging.Init()
}

var (
	// Default TTL for the environment
	baseEnvironmentConfig = &environment.Config{
		TTL: time.Hour * 72, // 3 days,
	}
	nodeCount           = 5
	remoteContainerName = "remote-test-runner"
	remoteFileList      = []string{"../../ops", "../../package.json", "../../yarn.lock", "../../tsconfig.json", "../../tsconfig.base.json", "../../packages-ts", "../../contracts"}
)

// Run the OCR soak test defined in ./tests/ocr_test.go
func TestOCRSoak(t *testing.T) {
	activeEVMNetwork := blockchain.LoadNetworkFromEnvironment() // Environment currently being used to soak test on
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

	// Checking if TTL env var is set to adjust duration to custom value
	ttlValue, ttlDefined := os.LookupEnv("TTL")
	if ttlDefined == true {
		ttl, err := time.ParseDuration(ttlValue)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
		}
		baseEnvironmentConfig.TTL = ttl
	}
	// Checking if count of OCR nodes is defined in ENV
	nodeCountSet, nodeCountDefined := os.LookupEnv("NODE_COUNT")
	if nodeCountDefined == true {
		nodeCount, err = strconv.Atoi(nodeCountSet)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper node count for the test: %v", err))
		}
	}
	baseEnvironmentConfig.NamespacePrefix = fmt.Sprintf(
		"chainlink-soak-ocr-starknet-%s",
		strings.ReplaceAll(strings.ToLower(activeEVMNetwork.Name), " ", "-"),
	)
	clConfig := map[string]interface{}{
		"replicas": nodeCount,
		"env":      common.GetDefaultCoreConfig(),
	}
	testEnvironment := common.GetDefaultEnvSetup(baseEnvironmentConfig, clConfig)
	remoteRunnerValues := map[string]interface{}{
		"test_name":      testTag,
		"env_namespace":  testEnvironment.Cfg.Namespace,
		"test_file_size": fmt.Sprint(exeFileSize),
		"test_log_level": "debug",
		"INSIDE_K8":      true,
		"TTL":            ttlValue,
		"NODE_COUNT":     nodeCount,
	}
	// Set evm network connection for remote runner
	for key, value := range activeEVMNetwork.ToMap() {
		remoteRunnerValues[key] = value
	}
	remoteRunnerWrapper := map[string]interface{}{"remote_test_runner": remoteRunnerValues}

	err = testEnvironment.
		AddHelm(remotetestrunner.New(remoteRunnerWrapper)).
		Run()
	require.NoError(t, err, "Error launching test environment")
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

	err = actions.TriggerRemoteTest(exeFile, testEnvironment)
	require.NoError(t, err, "Error activating remote test")
}
