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
	clImage             = "public.ecr.aws/chainlink/chainlink"
	clVersion           = "1.9.0"
	nodeCount           = 5              // default node count
	TTL                 = time.Hour * 72 // Default TTL for the env (3 days)
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
	baseEnvironmentConfig = &environment.Config{
		TTL: TTL,
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

	// Checking if version needs to be overridden env var is set in ENV
	envClImage, clImageDefined := os.LookupEnv("CL_IMAGE")
	envClVersion, clVersionDefined := os.LookupEnv("CL_VERSION")
	if clImageDefined && clVersionDefined {
		clImage = envClImage
		clVersion = envClVersion
	}

	// Checking if TTL env var is set in ENV
	ttlValue, ttlDefined := os.LookupEnv("TTL")
	if ttlDefined {
		TTL, err = time.ParseDuration(ttlValue)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
		}
		baseEnvironmentConfig.TTL = TTL
	}
	// Checking if count of OCR nodes is set in ENV
	nodeCountSet, nodeCountDefined := os.LookupEnv("NODE_COUNT")
	if nodeCountDefined {
		nodeCount, err = strconv.Atoi(nodeCountSet)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper node count for the test: %v", err))
		}
	}
	l2RpcUrl, _ := os.LookupEnv("L2_RPC_URL") // Fetch L2 RPC url if defined
	privateKey, _ := os.LookupEnv("PRIVATE_KEY")
	account, _ := os.LookupEnv("ACCOUNT")

	baseEnvironmentConfig.NamespacePrefix = fmt.Sprintf(
		"chainlink-soak-ocr-starknet-%s",
		strings.ReplaceAll(strings.ToLower(activeEVMNetwork.Name), " ", "-"),
	)
	clConfig := map[string]interface{}{
		"replicas": nodeCount,
		"env":      common.GetDefaultCoreConfig(),
		"chainlink": map[string]interface{}{
			"image": map[string]interface{}{
				"image":   clImage,
				"version": clVersion,
			},
		},
	}
	testEnvironment := common.GetDefaultEnvSetup(baseEnvironmentConfig, clConfig)
	remoteRunnerValues := actions.BasicRunnerValuesSetup(
		testTag,
		testEnvironment.Cfg.Namespace,
		"./integration-tests/soak/tests",
	)
	envValues := map[string]interface{}{
		"test_log_level": "debug",
		"INSIDE_K8":      true,
		"TTL":            ttlValue,
		"NODE_COUNT":     nodeCount,
		"L2_RPC_URL":     l2RpcUrl,
		"PRIVATE_KEY":    privateKey,
		"ACCOUNT":        account,
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
				"memory": "2048Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "2000m",
				"memory": "2048Mi",
			},
		},
	}

	err = testEnvironment.
		AddHelm(remotetestrunner.New(remoteRunnerWrapper)).
		Run()
	require.NoError(t, err, "Error launching test environment")
	// Copying required files / folders to remote runner pod
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

	err = actions.TriggerRemoteTest("../../", testEnvironment)
	require.NoError(t, err, "Error activating remote test")
}
