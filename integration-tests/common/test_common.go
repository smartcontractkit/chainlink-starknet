package common

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetdevnet "github.com/NethermindEth/starknet.go/devnet"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	test_env_ctf "github.com/smartcontractkit/chainlink-testing-framework/docker/test_env"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/integration-tests/docker/test_env"

	test_env_starknet "github.com/smartcontractkit/chainlink-starknet/integration-tests/docker/testenv"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/testconfig"

	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

var (
	rpcRequestTimeout = time.Second * 300
)

// OCRv2TestState Main testing state struct
type OCRv2TestState struct {
	Account           *AccountDetails
	Clients           *Clients
	ChainlinkNodesK8s []*client.ChainlinkK8sClient
	Common            *Common
	TestConfig        *TestConfig
	Contracts         *Contracts
}

// AccountDetails for deployment and funding
type AccountDetails struct {
	Account    string
	PrivateKey string
}

// Clients to access internal methods
type Clients struct {
	StarknetClient  *starknet.Client
	DevnetClient    *starknetdevnet.DevNet
	KillgraveClient *test_env_ctf.Killgrave
	OCR2Client      *ocr2.Client
	ChainlinkClient *ChainlinkClient
	GauntletClient  *gauntlet.StarknetGauntlet
	DockerEnv       *StarknetClusterTestEnv
}

// Contracts to store current deployed contract state
type Contracts struct {
	LinkTokenAddr         string
	OCRAddr               string
	AccessControllerAddr  string
	ProxyAddr             string
	ObservationSource     string
	JuelsPerFeeCoinSource string
}

// ChainlinkClient core node configs
type ChainlinkClient struct {
	NKeys            []client.NodeKeysBundle
	ChainlinkNodes   []*client.ChainlinkClient
	bTypeAttr        *client.BridgeTypeAttributes
	bootstrapPeers   []client.P2PData
	AccountAddresses []string
}

type StarknetClusterTestEnv struct {
	*test_env.CLClusterTestEnv
	Starknet  *test_env_starknet.Starknet
	Killgrave *test_env_ctf.Killgrave
}

type TestConfig struct {
	T          *testing.T
	L          zerolog.Logger
	TestConfig *testconfig.TestConfig
	Resty      *resty.Client
	err        error
}

func NewOCRv2State(t *testing.T, namespacePrefix string, testConfig *testconfig.TestConfig) (*OCRv2TestState, error) {
	c, err := New(testConfig).Default(t, namespacePrefix)
	if err != nil {
		return nil, err
	}
	state := &OCRv2TestState{
		Account: &AccountDetails{},
		Clients: &Clients{
			ChainlinkClient: &ChainlinkClient{},
		},
		Common: c,
		TestConfig: &TestConfig{
			T:          t,
			L:          log.Logger,
			TestConfig: testConfig,
			Resty:      nil,
			err:        nil,
		},
		Contracts: &Contracts{},
	}

	// Setting default job configs
	state.Contracts.ObservationSource = state.GetDefaultObservationSource()
	state.Contracts.JuelsPerFeeCoinSource = state.GetDefaultJuelsPerFeeCoinSource()

	if state.TestConfig.T != nil {
		state.TestConfig.L = logging.GetTestLogger(state.TestConfig.T)
	}

	return state, nil
}

// DeployCluster Deploys and sets up config of the environment and nodes
func (m *OCRv2TestState) DeployCluster() {
	// When running soak we need to use K8S
	if *m.Common.TestConfig.Common.InsideK8s {
		m.DeployEnv()

		if m.Common.Env.WillUseRemoteRunner() {
			return
		}

		m.Common.RPCDetails.RPCL2External = m.Common.Env.URLs["starknet-dev"][0]

		// Checking whether we are running in a remote runner since the forwarding is not working there and we need the public IP
		if m.Common.RPCDetails.RPCL2External == "http://127.0.0.1:0" {
			m.Common.RPCDetails.RPCL2External = m.Common.Env.URLs["starknet-dev"][1]
		}

		// Setting RPC details
		if *m.Common.TestConfig.Common.Network == "testnet" {
			m.Common.RPCDetails.RPCL2External = *m.Common.TestConfig.Common.L2RPCUrl
			m.Common.RPCDetails.RPCL2Internal = *m.Common.TestConfig.Common.L2RPCUrl
		}
		m.Common.RPCDetails.MockServerEndpoint = m.Common.Env.URLs["qa_mock_adapter_internal"][0]
		m.Common.RPCDetails.MockServerURL = "five"

	} else { // Otherwise use docker
		env, err := test_env.NewTestEnv()
		require.NoError(m.TestConfig.T, err)
		stark := test_env_starknet.NewStarknet([]string{env.DockerNetwork.Name}, *m.Common.TestConfig.Common.DevnetImage)
		err = stark.StartContainer()
		require.NoError(m.TestConfig.T, err)

		// Setting RPC details
		m.Common.RPCDetails.RPCL2External = stark.ExternalHTTPURL
		m.Common.RPCDetails.RPCL2Internal = stark.InternalHTTPURL

		if *m.Common.TestConfig.Common.Network == "testnet" {
			m.Common.RPCDetails.RPCL2External = *m.Common.TestConfig.Common.L2RPCUrl
			m.Common.RPCDetails.RPCL2Internal = *m.Common.TestConfig.Common.L2RPCUrl
		}

		// Creating docker containers
		b, err := test_env.NewCLTestEnvBuilder().
			WithNonEVM().
			WithTestInstance(m.TestConfig.T).
			WithTestConfig(m.TestConfig.TestConfig).
			WithMockAdapter().
			WithCLNodeConfig(m.Common.DefaultNodeConfig()).
			WithCLNodes(*m.Common.TestConfig.OCR2.NodeCount).
			WithCLNodeOptions(m.Common.TestEnvDetails.NodeOpts...).
			WithStandardCleanup().
			WithTestEnv(env)
		require.NoError(m.TestConfig.T, err)
		env, err = b.Build()
		require.NoError(m.TestConfig.T, err)
		m.Clients.DockerEnv = &StarknetClusterTestEnv{
			CLClusterTestEnv: env,
			Starknet:         stark,
			Killgrave:        env.MockAdapter,
		}

		// Setting up Mock adapter
		m.Clients.KillgraveClient = env.MockAdapter
		m.Common.RPCDetails.MockServerEndpoint = m.Clients.KillgraveClient.InternalEndpoint
		m.Common.RPCDetails.MockServerURL = "mockserver-bridge"
		err = m.Clients.KillgraveClient.SetAdapterBasedIntValuePath("/mockserver-bridge", []string{http.MethodGet, http.MethodPost}, 10)
		require.NoError(m.TestConfig.T, err, "Failed to set mock adapter value")
	}

	m.TestConfig.Resty = resty.New().SetBaseURL(m.Common.RPCDetails.RPCL2External)

	if *m.Common.TestConfig.Common.InsideK8s {
		m.ChainlinkNodesK8s, m.TestConfig.err = client.ConnectChainlinkNodes(m.Common.Env)
		require.NoError(m.TestConfig.T, m.TestConfig.err)
		m.Clients.ChainlinkClient.ChainlinkNodes = m.GetChainlinkNodes()
		m.Clients.ChainlinkClient.NKeys, m.TestConfig.err = m.Common.CreateNodeKeysBundle(m.Clients.ChainlinkClient.ChainlinkNodes)
		require.NoError(m.TestConfig.T, m.TestConfig.err)
	} else {
		m.Clients.ChainlinkClient.ChainlinkNodes = m.Clients.DockerEnv.ClCluster.NodeAPIs()
		m.Clients.ChainlinkClient.NKeys, m.TestConfig.err = m.Common.CreateNodeKeysBundle(m.Clients.DockerEnv.ClCluster.NodeAPIs())
		require.NoError(m.TestConfig.T, m.TestConfig.err)
	}
	lggr := logger.Nop()
	m.Clients.StarknetClient, m.TestConfig.err = starknet.NewClient(m.Common.ChainDetails.ChainID, m.Common.RPCDetails.RPCL2External, m.Common.RPCDetails.RPCL2InternalAPIKey, lggr, &rpcRequestTimeout)
	require.NoError(m.TestConfig.T, m.TestConfig.err, "Creating starknet client should not fail")
	m.Clients.OCR2Client, m.TestConfig.err = ocr2.NewClient(m.Clients.StarknetClient, lggr)
	require.NoError(m.TestConfig.T, m.TestConfig.err, "Creating ocr2 client should not fail")

	// If we are using devnet fetch the default keys
	if *m.Common.TestConfig.Common.Network == "localnet" {
		// fetch predeployed account 0 to use as funder
		m.Clients.DevnetClient = starknetdevnet.NewDevNet(m.Common.RPCDetails.RPCL2External)
		accounts, err := m.Clients.DevnetClient.Accounts()
		require.NoError(m.TestConfig.T, err)
		account := accounts[0]
		m.Account.Account = account.Address
		m.Account.PrivateKey = account.PrivateKey
	} else {
		m.Account.Account = *m.TestConfig.TestConfig.Common.Account
		m.Account.PrivateKey = *m.TestConfig.TestConfig.Common.PrivateKey
	}
}

// DeployEnv Deploys the environment
func (m *OCRv2TestState) DeployEnv() {
	err := m.Common.Env.Run()
	require.NoError(m.TestConfig.T, err)
}

// LoadOCR2Config Loads and returns the default starknet gauntlet config
func (m *OCRv2TestState) LoadOCR2Config() (*ops.OCR2Config, error) {
	var offChaiNKeys []string
	var onChaiNKeys []string
	var peerIDs []string
	var txKeys []string
	var cfgKeys []string
	for i, key := range m.Clients.ChainlinkClient.NKeys {
		offChaiNKeys = append(offChaiNKeys, key.OCR2Key.Data.Attributes.OffChainPublicKey)
		peerIDs = append(peerIDs, key.PeerID)
		txKeys = append(txKeys, m.Clients.ChainlinkClient.AccountAddresses[i])
		onChaiNKeys = append(onChaiNKeys, key.OCR2Key.Data.Attributes.OnChainPublicKey)
		cfgKeys = append(cfgKeys, key.OCR2Key.Data.Attributes.ConfigPublicKey)
	}

	var payload = ops.TestOCR2Config
	payload.Signers = onChaiNKeys
	payload.Transmitters = txKeys
	payload.OffchainConfig.OffchainPublicKeys = offChaiNKeys
	payload.OffchainConfig.PeerIDs = peerIDs
	payload.OffchainConfig.ConfigPublicKeys = cfgKeys

	return &payload, nil
}

func (m *OCRv2TestState) SetUpNodes() {
	err := m.Common.CreateJobsForContract(m.GetChainlinkClient(), m.Contracts.ObservationSource, m.Contracts.JuelsPerFeeCoinSource, m.Contracts.OCRAddr, m.Clients.ChainlinkClient.AccountAddresses)
	require.NoError(m.TestConfig.T, err, "Creating jobs should not fail")
}

// GetNodeKeys Returns the node key bundles
func (m *OCRv2TestState) GetNodeKeys() []client.NodeKeysBundle {
	return m.Clients.ChainlinkClient.NKeys
}

func (m *OCRv2TestState) GetChainlinkNodes() []*client.ChainlinkClient {
	// retrieve client from K8s client
	var chainlinkNodes []*client.ChainlinkClient
	for i := range m.ChainlinkNodesK8s {
		chainlinkNodes = append(chainlinkNodes, m.ChainlinkNodesK8s[i].ChainlinkClient)
	}
	return chainlinkNodes
}

func (m *OCRv2TestState) GetChainlinkClient() *ChainlinkClient {
	return m.Clients.ChainlinkClient
}

func (m *OCRv2TestState) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	m.Clients.ChainlinkClient.bTypeAttr = attr
}

func (m *OCRv2TestState) GetDefaultObservationSource() string {
	return `
			val [type = "bridge" name="mockserver-bridge"]
			parse [type="jsonparse" path="data,result"]
			val -> parse
			`
}

func (m *OCRv2TestState) GetDefaultJuelsPerFeeCoinSource() string {
	return `"""
			sum  [type="sum" values=<[451000]> ]
			sum
			"""
			`
}

func (m *OCRv2TestState) ValidateRounds(rounds int, isSoak bool) error {
	ctx := context.Background() // context background used because timeout handled by requestTimeout param
	// assert new rounds are occurring
	details := ocr2.TransmissionDetails{}
	increasing := 0 // track number of increasing rounds
	var stuck bool
	stuckCount := 0
	var positive bool

	// validate balance in aggregator
	linkContractAddress, err := starknetutils.HexToFelt(m.Contracts.LinkTokenAddr)
	if err != nil {
		return err
	}
	contractAddress, err := starknetutils.HexToFelt(m.Contracts.OCRAddr)
	if err != nil {
		return err
	}
	resLINK, errLINK := m.Clients.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: linkContractAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("balance_of"),
		Calldata:        []*felt.Felt{contractAddress},
	})
	require.NoError(m.TestConfig.T, errLINK, "Reader balance from LINK contract should not fail", "err", errLINK)
	resAgg, errAgg := m.Clients.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("link_available_for_payment"),
	})
	require.NoError(m.TestConfig.T, errAgg, "link_available_for_payment should not fail", "err", errAgg)
	balLINK := resLINK[0].BigInt(big.NewInt(0))
	balAgg := resAgg[1].BigInt(big.NewInt(0))
	isNegative := resAgg[0].BigInt(big.NewInt(0))
	if isNegative.Sign() > 0 {
		balAgg = new(big.Int).Neg(balAgg)
	}

	assert.Equal(m.TestConfig.T, balLINK.Cmp(big.NewInt(0)), 1, "Aggregator should have non-zero balance")
	assert.GreaterOrEqual(m.TestConfig.T, balLINK.Cmp(balAgg), 0, "Aggregator payment balance should be <= actual LINK balance")

	for start := time.Now(); time.Since(start) < m.Common.TestEnvDetails.TestDuration; {
		m.TestConfig.L.Info().Msg(fmt.Sprintf("Elapsed time: %s, Round wait: %s ", time.Since(start), m.Common.TestEnvDetails.TestDuration))
		res, err2 := m.Clients.OCR2Client.LatestTransmissionDetails(ctx, contractAddress)
		require.NoError(m.TestConfig.T, err2, "Failed to get latest transmission details")
		// end condition: enough rounds have occurred
		if !isSoak && increasing >= rounds && positive {
			break
		}

		// end condition: rounds have been stuck
		if stuck && stuckCount > 50 {
			m.TestConfig.L.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
			break
		}

		// try to fetch rounds
		time.Sleep(5 * time.Second)

		if err != nil {
			m.TestConfig.L.Error().Msg(fmt.Sprintf("Transmission Error: %+v", err))
			continue
		}
		m.TestConfig.L.Info().Msg(fmt.Sprintf("Transmission Details: %+v", res))

		// continue if no changes
		if res.Epoch == 0 && res.Round == 0 {
			continue
		}

		ansCmp := res.LatestAnswer.Cmp(big.NewInt(0))
		positive = ansCmp == 1 || positive

		// if changes from zero values set (should only initially)
		if res.Epoch > 0 && details.Epoch == 0 {
			if !isSoak {
				assert.Greater(m.TestConfig.T, res.Epoch, details.Epoch)
				assert.GreaterOrEqual(m.TestConfig.T, res.Round, details.Round)
				assert.NotEqual(m.TestConfig.T, ansCmp, 0) // assert changed from 0
				assert.NotEqual(m.TestConfig.T, res.Digest, details.Digest)
				assert.Equal(m.TestConfig.T, details.LatestTimestamp.Before(res.LatestTimestamp), true)
			}
			details = res
			continue
		}
		// check increasing rounds
		if !isSoak {
			assert.Equal(m.TestConfig.T, res.Digest, details.Digest, "Config digest should not change")
		} else {
			if res.Digest != details.Digest {
				m.TestConfig.L.Error().Msg(fmt.Sprintf("Config digest should not change, expected %s got %s", details.Digest, res.Digest))
			}
		}
		if (res.Epoch > details.Epoch || (res.Epoch == details.Epoch && res.Round > details.Round)) && details.LatestTimestamp.Before(res.LatestTimestamp) {
			increasing++
			stuck = false
			stuckCount = 0 // reset counter
			continue
		}

		// reach this point, answer has not changed
		stuckCount++
		if stuckCount > 30 {
			stuck = true
			increasing = 0
		}
	}
	if !isSoak {
		assert.GreaterOrEqual(m.TestConfig.T, increasing, rounds, "Round + epochs should be increasing")
		assert.Equal(m.TestConfig.T, positive, true, "Positive value should have been submitted")
		assert.Equal(m.TestConfig.T, stuck, false, "Round + epochs should not be stuck")
	}

	// Test proxy reading
	// TODO: would be good to test proxy switching underlying feeds

	proxyAddress, err := starknetutils.HexToFelt(m.Contracts.ProxyAddr)
	if err != nil {
		return err
	}
	roundDataRaw, err := m.Clients.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: proxyAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("latest_round_data"),
	})
	if !isSoak {
		require.NoError(m.TestConfig.T, err, "Reading round data from proxy should not fail")
		assert.Equal(m.TestConfig.T, len(roundDataRaw), 5, "Round data from proxy should match expected size")
	}
	valueBig := roundDataRaw[1].BigInt(big.NewInt(0))
	require.NoError(m.TestConfig.T, err)
	value := valueBig.Int64()
	if value < 0 {
		assert.Equal(m.TestConfig.T, value, int64(5), "Reading from proxy should return correct value")
	}

	return nil
}
