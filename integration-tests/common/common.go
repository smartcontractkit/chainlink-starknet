package common

import (
	"bytes"
	"fmt"
	"io"
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

	"github.com/smartcontractkit/chainlink-testing-framework/k8s/environment"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/pkg/alias"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay"
)

var (
	serviceKeyL1        = "Hardhat"
	serviceKeyL2        = "chainlink-starknet.starknet-devnet"
	serviceKeyChainlink = "chainlink"
	chainName           = "starknet"
	chainId             = "SN_GOERLI"
	defaultNodeUrl      = "http://127.0.0.1:5050"
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
	MockUrl             string
	PrivateKey          string
	Account             string
	ChainlinkConfig     string
	Env                 *environment.Environment
}

// getEnv gets the environment variable if it exists and sets it for the remote runner
func getEnv(v string) string {
	val := os.Getenv(v)
	if val != "" {
		os.Setenv(fmt.Sprintf("TEST_%s", v), val)
	}
	return val
}

func getNodeCount() int {
	// Checking if count of OCR nodes is defined in ENV
	nodeCountSet := getEnv("NODE_COUNT")
	if nodeCountSet == "" {
		nodeCountSet = "4"
	}
	nodeCount, err := strconv.Atoi(nodeCountSet)
	if err != nil {
		panic(fmt.Sprintf("Please define a proper node count for the test: %v", err))
	}
	return nodeCount
}

func getTTL() time.Duration {
	ttlValue := getEnv("TTL")
	if ttlValue == "" {
		ttlValue = "72h"
	}
	duration, err := time.ParseDuration(ttlValue)
	if err != nil {
		panic(fmt.Sprintf("Please define a proper TTL for the test: %v", err))
	}
	t, err := time.ParseDuration(*alias.ShortDur(duration))
	if err != nil {
		panic(fmt.Sprintf("Please define a proper TTL for the test: %v", err))
	}
	return t
}

func getTestDuration() time.Duration {
	testDurationValue := getEnv("TEST_DURATION")
	if testDurationValue == "" {
		return time.Duration(time.Minute * 15)
	}
	duration, err := time.ParseDuration(testDurationValue)
	if err != nil {
		panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
	}
	t, err := time.ParseDuration(*alias.ShortDur(duration))
	if err != nil {
		panic(fmt.Sprintf("Please define a proper duration for the test: %v", err))
	}
	return t
}

func New(t *testing.T) *Common {
	c := &Common{
		ChainName:           chainName,
		ChainId:             chainId,
		NodeCount:           getNodeCount(),
		TTL:                 getTTL(),
		TestDuration:        getTestDuration(),
		ServiceKeyChainlink: serviceKeyChainlink,
		ServiceKeyL1:        serviceKeyL1,
		ServiceKeyL2:        serviceKeyL2,
		L2RPCUrl:            getEnv("L2_RPC_URL"), // Fetch L2 RPC url if defined
		MockUrl:             "http://host.containers.internal:6060",
		PrivateKey:          getEnv("PRIVATE_KEY"),
		Account:             getEnv("ACCOUNT"),
		// P2PPort:             "6690",
	}
	c.Testnet = c.L2RPCUrl != ""

	// TODO: HAXX: we force the URL to a local docker container
	c.L2RPCUrl = defaultNodeUrl

	starknetUrl := fmt.Sprintf("http://%s:%d/rpc", serviceKeyL2, 5050)
	if c.Testnet {
		starknetUrl = c.L2RPCUrl
	}

	chainlinkConfig := fmt.Sprintf(`[[Starknet]]
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

[WebServer]
HTTPPort = 6688
[WebServer.TLS]
HTTPSPort = 0
`, c.ChainId, starknetUrl)

	c.ChainlinkConfig = chainlinkConfig
	log.Debug().Str("toml", chainlinkConfig).Msg("Created chainlink config")

	c.Env = &environment.Environment{}
	return c
}

// CapturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type CapturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// NewCapturingPassThroughWriter creates new CapturingPassThroughWriter
func NewCapturingPassThroughWriter(w io.Writer) *CapturingPassThroughWriter {
	return &CapturingPassThroughWriter{
		w: w,
	}
}

func (w *CapturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *CapturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func debug(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	if startErr := cmd.Start(); startErr != nil {
		panic(startErr)
	}

	doneStdOut := make(chan any)
	doneStdErr := make(chan any)
	osstdout := NewCapturingPassThroughWriter(os.Stdout)
	osstderr := NewCapturingPassThroughWriter(os.Stderr)
	go handleOutput(osstdout, stdout, doneStdOut)
	go handleOutput(osstderr, stderr, doneStdErr)

	err = cmd.Wait()

	errStdOut := <-doneStdOut
	if errStdOut != nil {
		fmt.Println("error writing to standard out")
	}

	errStdErr := <-doneStdErr
	if errStdErr != nil {
		fmt.Println("error writing to standard in")
	}

	if err != nil {
		fmt.Printf("Command finished with error: %v\n", err)
	}

	return err
}

func handleOutput(dst io.Writer, src io.Reader, done chan<- any) {
	_, err := io.Copy(dst, src)
	done <- err
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
	// easy debug
	err = debug(cmd)
	require.NoError(t, err, "Could not start core nodes")
	log.Info().Msg("Set up local stack complete.")

	// Set ChainlinkNodeDetails
	var nodeDetails []*environment.ChainlinkNodeDetail
	var basePort = 50100
	for i := 0; i < c.NodeCount; i++ {
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

// CreateKeys Creates node keys and defines chain and nodes for each node
func (c *Common) CreateKeys(env *environment.Environment) ([]client.NodeKeysBundle, []*client.ChainlinkClient, error) {
	nodes, err := connectChainlinkNodes(env)
	if err != nil {
		return nil, nil, err
	}

	NKeys, _, err := client.CreateNodeKeysBundle(nodes, c.ChainName, c.ChainId)
	if err != nil {
		return nil, nil, err
	}
	// for _, n := range nodes {
	// 	_, _, err = n.CreateStarkNetChain(&client.StarkNetChainAttributes{
	// 		Type:    c.ChainName,
	// 		ChainID: c.ChainId,
	// 		Config:  client.StarkNetChainConfig{},
	// 	})
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// 	_, _, err = n.CreateStarkNetNode(&client.StarkNetNodeAttributes{
	// 		Name:    c.ChainName,
	// 		ChainID: c.ChainId,
	// 		Url:     "http://", // TODO:
	// 	})
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// }
	return NKeys, nodes, nil
}

// CreateJobsForContract Creates and sets up the boostrap jobs as well as OCR jobs
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, mockUrl string, observationSource string, juelsPerFeeCoinSource string, ocrControllerAddress string, accountAddresses []string) error {
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
		"accountAddress": accountAddresses[0],
		"chainID":        c.ChainId,
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

	sourceValueBridge := &client.BridgeTypeAttributes{
		Name:        "mockserver-bridge",
		URL:         fmt.Sprintf("%s/%s", mockUrl, "five"),
		RequestData: "{}",
	}

	// Setting up job specs
	for nIdx, n := range cc.ChainlinkNodes {
		if nIdx == 0 {
			continue
		}
		_, err := n.CreateBridge(sourceValueBridge)
		if err != nil {
			return err
		}
		relayConfig := job.JSONConfig{
			"nodeName":       bootstrapRelayConfig["nodeName"],
			"accountAddress": accountAddresses[nIdx],
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
