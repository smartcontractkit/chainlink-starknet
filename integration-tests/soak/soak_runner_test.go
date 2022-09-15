package soak_test

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
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

var baseEnvironmentConfig = &environment.Config{
	TTL: time.Hour * 720, // 30 days,
}

// Run the OCR soak test defined in ./tests/ocr_test.go
func TestOCRSoak(t *testing.T) {
	activeEVMNetwork := networks.GeneralEVM() // Environment currently being used to soak test on

	baseEnvironmentConfig.NamespacePrefix = fmt.Sprintf(
		"chainlink-soak-ocr-starknet-%s",
		strings.ReplaceAll(strings.ToLower(activeEVMNetwork.Name), " ", "-"),
	)
	testEnvironment := environment.New(baseEnvironmentConfig)
	soakTestHelper(t, "@soak", testEnvironment, activeEVMNetwork)
}

// builds tests, launches environment, and triggers the soak test to run
func soakTestHelper(
	t *testing.T,
	testTag string,
	testEnvironment *environment.Environment,
	activeEVMNetwork *blockchain.EVMNetwork,
) {
	exeFile, exeFileSize, err := actions.BuildGoTests("./", "../soak/tests", "../../")
	require.NoError(t, err, "Error building go tests")

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
	clConfig := map[string]interface{}{
		"replicas": 5,
		"env": map[string]interface{}{
			"STARKNET_ENABLED":            "true",
			"EVM_ENABLED":                 "false",
			"EVM_RPC_ENABLED":             "false",
			"CHAINLINK_DEV":               "false",
			"FEATURE_OFFCHAIN_REPORTING2": "true",
			"feature_offchain_reporting":  "false",
			"P2P_NETWORKING_STACK":        "V2",
			"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:6690",
			"P2PV2_DELTA_DIAL":            "5s",
			"P2PV2_DELTA_RECONCILE":       "5s",
			"p2p_listen_port":             "0",
		},
	}
	err = testEnvironment.
		AddHelm(remotetestrunner.New(remoteRunnerWrapper)).
		AddHelm(devnet.New(nil)).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(chainlink.New(0, clConfig)).
		Run()

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../ops",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../package.json",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../package-lock.json",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../yarn.lock",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../tsconfig.json",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")

	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../tsconfig.base.json",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")
	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../packages-ts",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")
	testEnvironment.Client.CopyToPod(
		testEnvironment.Cfg.Namespace,
		"../../contracts",
		fmt.Sprintf("%s/%s:/root/", testEnvironment.Cfg.Namespace, "remote-test-runner"),
		"remote-test-runner")
	testEnvironment.Client.ExecuteInPod(testEnvironment.Cfg.Namespace, "remote-test-runner", "remote-test-runner", []string{"yarn", "install"})
	require.NoError(t, err, "Error launching test environment")
	err = actions.TriggerRemoteTest(exeFile, testEnvironment)
	require.NoError(t, err, "Error activating remote test")
}
