package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"

	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/environment"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/alias"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/chainlink"
	mock_adapter "github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/helm/mock-adapter"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay"
)

var (
	serviceKeyL1        = "Hardhat"
	serviceKeyL2        = "starknet-dev"
	serviceKeyChainlink = "chainlink"
	chainName           = "starknet"
	chainId             = "SN_GOERLI"
)

type Common struct {
	P2PPort             string
	ServiceKeyL1        string
	ServiceKeyL2        string
	ServiceKeyChainlink string
	ChainName           string
	ChainId             string
	NodeCount           int
	TTL                 time.Duration
	TestDuration        time.Duration
	Testnet             bool
	L2RPCUrl            string
	PrivateKey          string
	Account             string
	ClConfig            map[string]interface{}
	K8Config            *environment.Config
	Env                 *environment.Environment
}

func New() *Common {
	var err error
	c := &Common{
		ChainName:           chainName,
		ChainId:             chainId,
		ServiceKeyChainlink: serviceKeyChainlink,
		ServiceKeyL1:        serviceKeyL1,
		ServiceKeyL2:        serviceKeyL2,
	}
	// Checking if count of OCR nodes is defined in ENV
	nodeCountSet := getEnv("NODE_COUNT")
	if nodeCountSet != "" {
		c.NodeCount, err = strconv.Atoi(nodeCountSet)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper node count for the test: %v", err))
		}
	} else {
		panic("Please define NODE_COUNT")
	}

	// Checking if TTL env var is set in ENV
	ttlValue := getEnv("TTL")
	if ttlValue != "" {
		duration, err := time.ParseDuration(ttlValue)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the namespace: %v", err))
		}
		c.TTL, err = time.ParseDuration(*alias.ShortDur(duration))
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the namespace: %v", err))
		}
	} else {
		panic("Please define TTL of env")
	}

	// Setting optional parameters
	testDurationValue := getEnv("TEST_DURATION")
	if testDurationValue != "" {
		duration, err := time.ParseDuration(testDurationValue)
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
		}
		c.TestDuration, err = time.ParseDuration(*alias.ShortDur(duration))
		if err != nil {
			panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
		}
	} else {
		c.TestDuration = time.Duration(time.Minute * 15)
	}
	c.L2RPCUrl = getEnv("L2_RPC_URL") // Fetch L2 RPC url if defined
	c.Testnet = c.L2RPCUrl != ""
	c.PrivateKey = getEnv("PRIVATE_KEY")
	c.Account = getEnv("ACCOUNT")

	return c
}

// getEnv gets the environment variable if it exists and sets it for the remote runner
func getEnv(v string) string {
	val := os.Getenv(v)
	if val != "" {
		os.Setenv(fmt.Sprintf("TEST_%s", v), val)
	}
	return val
}

// CreateKeys Creates node keys and defines chain and nodes for each node
func (c *Common) CreateKeys(env *environment.Environment) ([]client.NodeKeysBundle, []*client.ChainlinkK8sClient, error) {
	chainlinkK8Nodes, err := client.ConnectChainlinkNodes(env)
	if err != nil {
		return nil, nil, err
	}

	// extract client from k8s client
	ChainlinkNodes := []*client.ChainlinkClient{}
	for i := range chainlinkK8Nodes {
		ChainlinkNodes = append(ChainlinkNodes, chainlinkK8Nodes[i].ChainlinkClient)
	}

	NKeys, _, err := client.CreateNodeKeysBundle(ChainlinkNodes, c.ChainName, c.ChainId)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range ChainlinkNodes {
		_, _, err = n.CreateStarkNetChain(&client.StarkNetChainAttributes{
			Type:    c.ChainName,
			ChainID: c.ChainId,
			Config:  client.StarkNetChainConfig{},
		})
		if err != nil {
			return nil, nil, err
		}
		_, _, err = n.CreateStarkNetNode(&client.StarkNetNodeAttributes{
			Name:    c.ChainName,
			ChainID: c.ChainId,
			Url:     env.URLs[c.ServiceKeyL2][1],
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return NKeys, chainlinkK8Nodes, nil
}

// CreateJobsForContract Creates and sets up the boostrap jobs as well as OCR jobs
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, observationSource string, juelsPerFeeCoinSource string, ocrControllerAddress string, accountAddresses []string) error {
	// Define node[0] as bootstrap node
	cc.bootstrapPeers = []client.P2PData{
		{
			InternalIP:   cc.ChainlinkNodes[0].InternalIP(),
			InternalPort: c.P2PPort,
			PeerID:       cc.NKeys[0].PeerID,
		},
	}

	// Defining relay config
	bootstrapRelayConfig := job.JSONConfig{
		"nodeName":       fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.New().String()),
		"accountAddress": fmt.Sprintf("%s", accountAddresses[0]),
		"chainID":        fmt.Sprintf("%s", c.ChainId),
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

	_, _, err := cc.ChainlinkNodes[0].CreateJob(jobSpec)
	if err != nil {
		return err
	}

	var p2pBootstrappers []string

	for i := range cc.bootstrapPeers {
		p2pBootstrappers = append(p2pBootstrappers, cc.bootstrapPeers[i].P2PV2Bootstrapper())
	}

	// Setting up job specs
	for nIdx, n := range cc.ChainlinkNodes {
		if nIdx == 0 {
			continue
		}
		_, err := n.CreateBridge(cc.bTypeAttr)
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
		_, _, err = n.CreateJob(jobSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Common) Default(t *testing.T) {
	c.K8Config = &environment.Config{NamespacePrefix: "chainlink-ocr-starknet", TTL: c.TTL, Test: t}
	starknetUrl := fmt.Sprintf("http://%s:%d/rpc", serviceKeyL2, 5000)
	if c.Testnet {
		starknetUrl = c.L2RPCUrl
	}
	baseTOML := fmt.Sprintf(`[[Starknet]]
Enabled = true
ChainID = '%s'
[[Starknet.Nodes]]
Name = 'primary'
URL = '%s'

[OCR2]
Enabled = true

[P2P]
[P2P.V2]
Enabled = true
DeltaDial = '5s'
DeltaReconcile = '5s'
ListenAddresses = ['0.0.0.0:6690']
`, c.ChainId, starknetUrl)
	log.Debug().Str("toml", baseTOML).Msg("TOML")
	c.ClConfig = map[string]interface{}{
		"replicas": c.NodeCount,
		"toml":     baseTOML,
		"db": map[string]any{
			"stateful": true,
		},
	}
	c.Env = environment.New(c.K8Config).
		AddHelm(devnet.New(nil)).
		AddHelm(mock_adapter.New(nil)).
		AddHelm(chainlink.New(0, c.ClConfig))
}
