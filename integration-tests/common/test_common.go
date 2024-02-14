package common

import (
	"context"
	"fmt"
	starknetdevnet "github.com/NethermindEth/starknet.go/devnet"
	"github.com/go-resty/resty/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	test_env_ctf "github.com/smartcontractkit/chainlink-testing-framework/docker/test_env"
	"github.com/smartcontractkit/chainlink/integration-tests/docker/test_env"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	test_env_starknet "github.com/smartcontractkit/chainlink-starknet/integration-tests/docker/test_env"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	"github.com/smartcontractkit/chainlink/integration-tests/testconfig"
	"math/big"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

var (
	rpcRequestTimeout = time.Second * 300
	dumpPath          = "/dumps/dump.pkl"
)

type OCRv2TestState struct {
	Account               string
	PrivateKey            string
	StarknetClient        *starknet.Client
	DevnetClient          *starknetdevnet.DevNet
	Killgrave             *test_env_ctf.Killgrave
	ChainlinkNodesK8s     []*client.ChainlinkK8sClient
	Cc                    *ChainlinkClient
	OCR2Client            *ocr2.Client
	Sg                    *gauntlet.StarknetGauntlet
	L1RPCUrl              string
	Common                *Common
	AccountAddresses      []string
	LinkTokenAddr         string
	OCRAddr               string
	AccessControllerAddr  string
	ProxyAddr             string
	ObservationSource     string
	JuelsPerFeeCoinSource string
	T                     *testing.T
	L                     zerolog.Logger
	TestConfig            *testconfig.TestConfig
	Resty                 *resty.Client
	err                   error
}

type ChainlinkClient struct {
	NKeys          []client.NodeKeysBundle
	ChainlinkNodes []*client.ChainlinkClient
	bTypeAttr      *client.BridgeTypeAttributes
	bootstrapPeers []client.P2PData
}

func NewOCRv2State(t *testing.T, env string, isK8s bool, namespacePrefix string, testConfig *testconfig.TestConfig) (*OCRv2TestState, error) {
	c, err := New(env, isK8s).Default(t, namespacePrefix)
	if err != nil {
		return nil, err
	}
	state := &OCRv2TestState{
		Common:         c,
		T:              t,
		L:              log.Logger,
		TestConfig:     testConfig,
		Cc:             &ChainlinkClient{},
		StarknetClient: &starknet.Client{},
	}

	if state.T != nil {
		state.L = logging.GetTestLogger(state.T)
	}

	return state, nil
}

// DeployCluster Deploys and sets up config of the environment and nodes
func (m *OCRv2TestState) DeployCluster() {
	if m.Common.IsK8s {
		m.DeployEnv()
	} else {
		env, err := test_env.NewTestEnv()
		require.NoError(m.T, err)
		stark := test_env_starknet.NewStarknet([]string{env.Network.Name})
		err = stark.StartContainer()
		require.NoError(m.T, err)
		m.Common.L2RPCUrl = stark.ExternalHttpUrl
		m.Resty = resty.New().SetBaseURL(m.Common.L2RPCUrl)
		b, err := test_env.NewCLTestEnvBuilder().
			WithNonEVM().
			WithTestInstance(m.T).
			WithTestConfig(m.TestConfig).
			WithMockAdapter().
			WithCLNodeConfig(m.Common.DefaultNodeConfig()).
			WithCLNodes(m.Common.NodeCount).
			WithCLNodeOptions(m.Common.NodeOpts...).
			WithStandardCleanup().
			WithTestEnv(env)
		require.NoError(m.T, err)
		env, err = b.Build()
		require.NoError(m.T, err)
		m.Common.DockerEnv = &StarknetClusterTestEnv{
			CLClusterTestEnv: env,
			Starknet:         stark,
			Killgrave:        env.MockAdapter,
		}
		m.Killgrave = env.MockAdapter
	}

	m.ObservationSource = m.GetDefaultObservationSource()
	m.JuelsPerFeeCoinSource = m.GetDefaultJuelsPerFeeCoinSource()
	m.SetupClients()
	if m.Common.IsK8s {
		m.Cc.NKeys, m.err = m.Common.CreateNodeKeysBundle(m.GetChainlinkNodes())
		require.NoError(m.T, m.err)
	} else {
		m.Cc.NKeys, m.err = m.Common.CreateNodeKeysBundle(m.Common.DockerEnv.ClCluster.NodeAPIs())
		require.NoError(m.T, m.err)
	}
	lggr := logger.Nop()
	m.StarknetClient, m.err = starknet.NewClient(m.Common.ChainId, m.Common.L2RPCUrl, lggr, &rpcRequestTimeout)
	require.NoError(m.T, m.err, "Creating starknet client should not fail")
	m.OCR2Client, m.err = ocr2.NewClient(m.StarknetClient, lggr)
	require.NoError(m.T, m.err, "Creating ocr2 client should not fail")
	if !m.Common.Testnet {
		// fetch predeployed account 0 to use as funder
		m.DevnetClient = starknetdevnet.NewDevNet(m.Common.L2RPCUrl)
		accounts, err := m.DevnetClient.Accounts()
		require.NoError(m.T, err)
		account := accounts[0]
		m.Account = account.Address
		m.PrivateKey = account.PrivateKey
	}
}

// DeployEnv Deploys the environment
func (m *OCRv2TestState) DeployEnv() {
	err := m.Common.Env.Run()
	require.NoError(m.T, err)
}

// SetupClients Sets up the starknet client
func (m *OCRv2TestState) SetupClients() {
	if m.Common.IsK8s {
		m.ChainlinkNodesK8s, m.err = client.ConnectChainlinkNodes(m.Common.Env)
		require.NoError(m.T, m.err)
	} else {
		m.Cc.ChainlinkNodes = m.Common.DockerEnv.ClCluster.NodeAPIs()
	}
}

// LoadOCR2Config Loads and returns the default starknet gauntlet config
func (m *OCRv2TestState) LoadOCR2Config() (*ops.OCR2Config, error) {
	var offChaiNKeys []string
	var onChaiNKeys []string
	var peerIds []string
	var txKeys []string
	var cfgKeys []string
	for i, key := range m.Cc.NKeys {
		offChaiNKeys = append(offChaiNKeys, key.OCR2Key.Data.Attributes.OffChainPublicKey)
		peerIds = append(peerIds, key.PeerID)
		txKeys = append(txKeys, m.AccountAddresses[i])
		onChaiNKeys = append(onChaiNKeys, key.OCR2Key.Data.Attributes.OnChainPublicKey)
		cfgKeys = append(cfgKeys, key.OCR2Key.Data.Attributes.ConfigPublicKey)
	}

	var payload = ops.TestOCR2Config
	payload.Signers = onChaiNKeys
	payload.Transmitters = txKeys
	payload.OffchainConfig.OffchainPublicKeys = offChaiNKeys
	payload.OffchainConfig.PeerIds = peerIds
	payload.OffchainConfig.ConfigPublicKeys = cfgKeys

	return &payload, nil
}

func (m *OCRv2TestState) SetUpNodes() {
	err := m.Common.CreateJobsForContract(m.GetChainlinkClient(), m.Killgrave, m.ObservationSource, m.JuelsPerFeeCoinSource, m.OCRAddr, m.AccountAddresses)
	require.NoError(m.T, err, "Creating jobs should not fail")
}

// GetNodeKeys Returns the node key bundles
func (m *OCRv2TestState) GetNodeKeys() []client.NodeKeysBundle {
	return m.Cc.NKeys
}

func (m *OCRv2TestState) GetChainlinkNodes() []*client.ChainlinkClient {
	return m.Cc.ChainlinkNodes
}

func (m *OCRv2TestState) GetChainlinkClient() *ChainlinkClient {
	return m.Cc
}

func (m *OCRv2TestState) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	m.Cc.bTypeAttr = attr
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
	linkContractAddress, err := starknetutils.HexToFelt(m.LinkTokenAddr)
	if err != nil {
		return err
	}
	contractAddress, err := starknetutils.HexToFelt(m.OCRAddr)
	if err != nil {
		return err
	}
	resLINK, errLINK := m.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: linkContractAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("balance_of"),
		Calldata:        []*felt.Felt{contractAddress},
	})
	require.NoError(m.T, errLINK, "Reader balance from LINK contract should not fail", "err", errLINK)
	resAgg, errAgg := m.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("link_available_for_payment"),
	})
	require.NoError(m.T, errAgg, "link_available_for_payment should not fail", "err", errAgg)
	balLINK := resLINK[0].BigInt(big.NewInt(0))
	balAgg := resAgg[1].BigInt(big.NewInt(0))
	isNegative := resAgg[0].BigInt(big.NewInt(0))
	if isNegative.Sign() > 0 {
		balAgg = new(big.Int).Neg(balAgg)
	}

	assert.Equal(m.T, balLINK.Cmp(big.NewInt(0)), 1, "Aggregator should have non-zero balance")
	assert.GreaterOrEqual(m.T, balLINK.Cmp(balAgg), 0, "Aggregator payment balance should be <= actual LINK balance")

	err = m.Killgrave.SetAdapterBasedIntValuePath("/mockserver-bridge", []string{http.MethodGet, http.MethodPost}, 10)
	require.NoError(m.T, err, "Failed to set mock adapter value")
	for start := time.Now(); time.Since(start) < m.Common.TestDuration; {
		m.L.Info().Msg(fmt.Sprintf("Elapsed time: %s, Round wait: %s ", time.Since(start), m.Common.TestDuration))
		res, err2 := m.OCR2Client.LatestTransmissionDetails(ctx, contractAddress)
		require.NoError(m.T, err2, "Failed to get latest transmission details")
		// end condition: enough rounds have occurred
		if !isSoak && increasing >= rounds && positive {
			break
		}

		// end condition: rounds have been stuck
		if stuck && stuckCount > 50 {
			m.L.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
			break
		}

		// try to fetch rounds
		time.Sleep(5 * time.Second)

		if err != nil {
			m.L.Error().Msg(fmt.Sprintf("Transmission Error: %+v", err))
			continue
		}
		m.L.Info().Msg(fmt.Sprintf("Transmission Details: %+v", res))

		// continue if no changes
		if res.Epoch == 0 && res.Round == 0 {
			continue
		}

		ansCmp := res.LatestAnswer.Cmp(big.NewInt(0))
		positive = ansCmp == 1 || positive

		// if changes from zero values set (should only initially)
		if res.Epoch > 0 && details.Epoch == 0 {
			if !isSoak {
				assert.Greater(m.T, res.Epoch, details.Epoch)
				assert.GreaterOrEqual(m.T, res.Round, details.Round)
				assert.NotEqual(m.T, ansCmp, 0) // assert changed from 0
				assert.NotEqual(m.T, res.Digest, details.Digest)
				assert.Equal(m.T, details.LatestTimestamp.Before(res.LatestTimestamp), true)
			}
			details = res
			continue
		}
		// check increasing rounds
		if !isSoak {
			assert.Equal(m.T, res.Digest, details.Digest, "Config digest should not change")
		} else {
			if res.Digest != details.Digest {
				m.L.Error().Msg(fmt.Sprintf("Config digest should not change, expected %s got %s", details.Digest, res.Digest))
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
		assert.GreaterOrEqual(m.T, increasing, rounds, "Round + epochs should be increasing")
		assert.Equal(m.T, positive, true, "Positive value should have been submitted")
		assert.Equal(m.T, stuck, false, "Round + epochs should not be stuck")
	}

	// Test proxy reading
	// TODO: would be good to test proxy switching underlying feeds

	proxyAddress, err := starknetutils.HexToFelt(m.ProxyAddr)
	if err != nil {
		return err
	}
	roundDataRaw, err := m.StarknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: proxyAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("latest_round_data"),
	})
	if !isSoak {
		require.NoError(m.T, err, "Reading round data from proxy should not fail")
		assert.Equal(m.T, len(roundDataRaw), 5, "Round data from proxy should match expected size")
	}
	valueBig := roundDataRaw[1].BigInt(big.NewInt(0))
	require.NoError(m.T, err)
	value := valueBig.Int64()
	if value < 0 {
		assert.Equal(m.T, value, int64(5), "Reading from proxy should return correct value")
	}

	return nil
}
