package common

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/testconfig"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/config"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/chainlink"
	mock_adapter "github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/mock-adapter"
	"github.com/smartcontractkit/chainlink-testing-framework/utils/ptr"
	"github.com/smartcontractkit/chainlink/integration-tests/docker/test_env"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	common_cfg "github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/environment"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/integration-tests/types/config/node"
	cl "github.com/smartcontractkit/chainlink/v2/core/services/chainlink"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay"
)

var (
	chainName            = "starknet"
	chainId              = "SN_GOERLI"
	DefaultL2RPCInternal = "http://starknet-dev:5000"
)

type Common struct {
	ChainDetails    *ChainDetails
	TestEnvDetails  *TestEnvDetails
	Env             *environment.Environment
	RPCDetails      *RPCDetails
	ChainlinkConfig string
	TestConfig      *testconfig.TestConfig
}

type ChainDetails struct {
	ChainName string
	ChainId   string
}

type TestEnvDetails struct {
	TestDuration time.Duration
	K8Config     *environment.Config
	NodeOpts     []test_env.ClNodeOption
}

type RPCDetails struct {
	RPCL1Internal      string
	RPCL2Internal      string
	RPCL1External      string
	RPCL2External      string
	MockServerUrl      string
	MockServerEndpoint string
	P2PPort            string
}

func New(testConfig *testconfig.TestConfig) *Common {
	var c *Common

	duration, err := time.ParseDuration(*testConfig.OCR2.TestDuration)
	if err != nil {
		panic("Invalid test duration")
	}

	c = &Common{
		TestConfig: testConfig,
		ChainDetails: &ChainDetails{
			ChainName: chainName,
			ChainId:   chainId,
		},
		TestEnvDetails: &TestEnvDetails{
			TestDuration: duration,
		},
		RPCDetails: &RPCDetails{
			P2PPort:       "6690",
			RPCL2Internal: DefaultL2RPCInternal,
		},
	}

	return c
}

func (c *Common) Default(t *testing.T, namespacePrefix string) (*Common, error) {
	c.TestEnvDetails.K8Config = &environment.Config{
		NamespacePrefix: fmt.Sprintf("starknet-%s", namespacePrefix),
		TTL:             c.TestEnvDetails.TestDuration,
		Test:            t,
	}

	if *c.TestConfig.Common.InsideK8s {
		toml := c.DefaultNodeConfig()
		tomlString, err := toml.TOMLString()
		if err != nil {
			return nil, err
		}
		c.Env = environment.New(c.TestEnvDetails.K8Config).
			AddHelm(devnet.New(nil)).
			AddHelm(mock_adapter.New(nil)).
			AddHelm(chainlink.New(0, map[string]interface{}{
				"toml":     tomlString,
				"replicas": c.TestConfig.OCR2.NodeCount,
			}))
	}

	return c, nil
}

func (c *Common) DefaultNodeConfig() *cl.Config {
	starkConfig := config.TOMLConfig{
		Enabled: ptr.Ptr(true),
		ChainID: ptr.Ptr(c.ChainDetails.ChainId),
		Nodes: []*config.Node{
			{
				Name: ptr.Ptr("primary"),
				URL:  common_cfg.MustParseURL(c.RPCDetails.RPCL2Internal),
			},
		},
	}
	baseConfig := node.NewBaseConfig()
	baseConfig.Starknet = config.TOMLConfigs{
		&starkConfig,
	}
	baseConfig.OCR2.Enabled = ptr.Ptr(true)
	baseConfig.P2P.V2.Enabled = ptr.Ptr(true)
	fiveSecondDuration := common_cfg.MustNewDuration(5 * time.Second)

	baseConfig.P2P.V2.DeltaDial = fiveSecondDuration
	baseConfig.P2P.V2.DeltaReconcile = fiveSecondDuration
	baseConfig.P2P.V2.ListenAddresses = &[]string{"0.0.0.0:6690"}

	return baseConfig
}

func (c *Common) SetLocalEnvironment(t *testing.T) {
	// Run scripts to set up local test environment
	log.Info().Msg("Starting starknet-devnet container...")
	err := exec.Command("../../scripts/devnet.sh").Run()
	require.NoError(t, err, "Could not start devnet container")
	// TODO: add hardhat too
	log.Info().Msg("Starting postgres container...")
	err = exec.Command("../../scripts/postgres.sh").Run()
	require.NoError(t, err, "Could not start postgres container")
	log.Info().Msg("Starting mock adapter...")
	err = exec.Command("../../scripts/mock-adapter.sh").Run()
	require.NoError(t, err, "Could not start mock adapter")
	log.Info().Msg("Starting core nodes...")
	cmd := exec.Command("../../scripts/core.sh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CL_CONFIG=%s", c.ChainlinkConfig))
	err = cmd.Run()
	require.NoError(t, err, "Could not start core nodes")
	log.Info().Msg("Set up local stack complete.")

	// Set ChainlinkNodeDetails
	var nodeDetails []*environment.ChainlinkNodeDetail
	var basePort = 50100
	for i := 0; i < *c.TestConfig.OCR2.NodeCount; i++ {
		dbLocalIP := fmt.Sprintf("postgresql://postgres:postgres@chainlink.postgres:5432/starknet_test_%d?sslmode=disable", i+1)
		nodeDetails = append(nodeDetails, &environment.ChainlinkNodeDetail{
			ChartName: "unused",
			PodName:   "unused",
			LocalIP:   "http://127.0.0.1:" + strconv.Itoa(basePort+i),
			// InternalIP: "http://host.container.internal:" + strconv.Itoa(basePort+i), // TODO: chainlink.core.${i}:6688
			InternalIP: fmt.Sprintf("http://chainlink.core.%d:6688", i+1), // TODO: chainlink.core.1:6688
			DBLocalIP:  dbLocalIP,
		})
	}
	c.Env.ChainlinkNodeDetails = nodeDetails
}

func (c *Common) TearDownLocalEnvironment(t *testing.T) {
	log.Info().Msg("Tearing down core nodes...")
	err := exec.Command("../../scripts/core.down.sh").Run()
	require.NoError(t, err, "Could not tear down core nodes")
	log.Info().Msg("Tearing down mock adapter...")
	err = exec.Command("../../scripts/mock-adapter.down.sh").Run()
	require.NoError(t, err, "Could not tear down mock adapter")
	log.Info().Msg("Tearing down postgres container...")
	err = exec.Command("../../scripts/postgres.down.sh").Run()
	require.NoError(t, err, "Could not tear down postgres container")
	log.Info().Msg("Tearing down devnet container...")
	err = exec.Command("../../scripts/devnet.down.sh").Run()
	require.NoError(t, err, "Could not tear down devnet container")
	log.Info().Msg("Tear down local stack complete.")
}

// connectChainlinkNodes creates a chainlink client for each node in the environment
// This is a non k8s version of the function in chainlink_k8s.go
// https://github.com/smartcontractkit/chainlink/blob/cosmos-test-keys/integration-tests/client/chainlink_k8s.go#L77
func connectChainlinkNodes(e *environment.Environment) ([]*client.ChainlinkClient, error) {
	var clients []*client.ChainlinkClient
	for _, nodeDetails := range e.ChainlinkNodeDetails {
		c, err := client.NewChainlinkClient(&client.ChainlinkConfig{
			URL:        nodeDetails.LocalIP,
			Email:      "notreal@fakeemail.ch",
			Password:   "fj293fbBnlQ!f9vNs",
			InternalIP: parseHostname(nodeDetails.InternalIP),
		}, log.Logger)
		if err != nil {
			return nil, err
		}
		log.Debug().
			Str("URL", c.Config.URL).
			Str("Internal IP", c.Config.InternalIP).
			Str("Chart Name", nodeDetails.ChartName).
			Str("Pod Name", nodeDetails.PodName).
			Msg("Connected to Chainlink node")
		clients = append(clients, c)
	}
	return clients, nil
}

func parseHostname(s string) string {
	r := regexp.MustCompile(`://(?P<Host>.*):`)
	return r.FindStringSubmatch(s)[1]
}

func (c *Common) CreateNodeKeysBundle(nodes []*client.ChainlinkClient) ([]client.NodeKeysBundle, error) {
	nkb := make([]client.NodeKeysBundle, 0)
	for _, n := range nodes {
		p2pkeys, err := n.MustReadP2PKeys()
		if err != nil {
			return nil, err
		}

		peerID := p2pkeys.Data[0].Attributes.PeerID
		txKey, _, err := n.CreateTxKey(chainName, c.ChainDetails.ChainId)
		if err != nil {
			return nil, err
		}
		ocrKey, _, err := n.CreateOCR2Key(chainName)
		if err != nil {
			return nil, err
		}

		nkb = append(nkb, client.NodeKeysBundle{
			PeerID:  peerID,
			OCR2Key: *ocrKey,
			TXKey:   *txKey,
		})
	}
	return nkb, nil
}

// CreateJobsForContract Creates and sets up the boostrap jobs as well as OCR jobs
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, observationSource string, juelsPerFeeCoinSource string, ocrControllerAddress string, accountAddresses []string) error {
	// Define node[0] as bootstrap node
	cc.bootstrapPeers = []client.P2PData{
		{
			InternalIP:   cc.ChainlinkNodes[0].InternalIP(),
			InternalPort: c.RPCDetails.P2PPort,
			PeerID:       cc.NKeys[0].PeerID,
		},
	}

	// Defining relay config
	bootstrapRelayConfig := job.JSONConfig{
		"nodeName":       fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.New().String()),
		"accountAddress": fmt.Sprintf("%s", accountAddresses[0]),
		"chainID":        fmt.Sprintf("%s", c.ChainDetails.ChainId),
	}

	oracleSpec := job.OCR2OracleSpec{
		ContractID:                  ocrControllerAddress,
		Relay:                       relay.StarkNet,
		RelayConfig:                 bootstrapRelayConfig,
		ContractConfigConfirmations: 1, // don't wait for confirmation on devnet
	}
	// Setting up bootstrap node
	jobSpec := &client.OCR2TaskJobSpec{
		Name:           fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.New().String()),
		JobType:        "bootstrap",
		OCR2OracleSpec: oracleSpec,
	}
	fmt.Println(jobSpec.String())
	_, _, err := cc.ChainlinkNodes[0].CreateJob(jobSpec)
	if err != nil {
		return err
	}

	var p2pBootstrappers []string

	for i := range cc.bootstrapPeers {
		p2pBootstrappers = append(p2pBootstrappers, cc.bootstrapPeers[i].P2PV2Bootstrapper())
	}

	sourceValueBridge := &client.BridgeTypeAttributes{
		Name: "mockserver-bridge",
		URL:  fmt.Sprintf("%s/%s", c.RPCDetails.MockServerEndpoint, strings.TrimPrefix(c.RPCDetails.MockServerUrl, "/")),
	}

	// Setting up job specs
	for nIdx, n := range cc.ChainlinkNodes {
		if nIdx == 0 {
			continue
		}
		err := n.MustCreateBridge(sourceValueBridge)
		if err != nil {
			return err
		}
		relayConfig := job.JSONConfig{
			"nodeName":       bootstrapRelayConfig["nodeName"],
			"accountAddress": fmt.Sprintf("%s", accountAddresses[nIdx]),
			"chainID":        bootstrapRelayConfig["chainID"],
		}

		oracleSpec = job.OCR2OracleSpec{
			ContractID:                  ocrControllerAddress,
			Relay:                       relay.StarkNet,
			RelayConfig:                 relayConfig,
			PluginType:                  "median",
			OCRKeyBundleID:              null.StringFrom(cc.NKeys[nIdx].OCR2Key.Data.ID),
			TransmitterID:               null.StringFrom(cc.NKeys[nIdx].TXKey.Data.ID),
			P2PV2Bootstrappers:          pq.StringArray{strings.Join(p2pBootstrappers, ",")},
			ContractConfigConfirmations: 1, // don't wait for confirmation on devnet
			PluginConfig: job.JSONConfig{
				"juelsPerFeeCoinSource": juelsPerFeeCoinSource,
			},
		}

		jobSpec = &client.OCR2TaskJobSpec{
			Name:              fmt.Sprintf("starknet-OCRv2-%d-%s", nIdx, uuid.New().String()),
			JobType:           "offchainreporting2",
			OCR2OracleSpec:    oracleSpec,
			ObservationSource: observationSource,
		}
		fmt.Println(jobSpec.String())
		_, err = n.MustCreateJob(jobSpec)
		if err != nil {
			return err
		}
	}
	return nil
}
