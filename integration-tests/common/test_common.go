package common

import (
	"context"
	"encoding/hex"
	"fmt"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

var (
	err error

	// These are one of the default addresses based on the seed we pass to devnet which is 0
	defaultWalletPrivKey = ops.PrivateKeys0Seed[0]
	defaultWalletAddress string // derived in init()
	rpcRequestTimeout    = time.Second * 300
	dumpPath             = "/dumps/dump.pkl"
	mockServerValue      = 900000
)

func init() {
	// wallet contract derivation
	var keyBytes []byte
	keyBytes, err = hex.DecodeString(strings.TrimPrefix(defaultWalletPrivKey, "0x"))
	if err != nil {
		panic(err)
	}
	defaultWalletAddress = "0x" + hex.EncodeToString(keys.PubKeyToAccount(keys.Raw(keyBytes).Key().PublicKey(), ops.DevnetClassHash, ops.DevnetSalt))
}

type Test struct {
	Devnet                *devnet.StarkNetDevnetClient
	Cc                    *ChainlinkClient
	Starknet              *starknet.Client
	OCR2Client            *ocr2.Client
	Sg                    *gauntlet.StarknetGauntlet
	mockServer            *ctfClient.MockserverClient
	L1RPCUrl              string
	Common                *Common
	LinkTokenAddr         string
	OCRAddr               string
	AccessControllerAddr  string
	ProxyAddr             string
	ObservationSource     string
	JuelsPerFeeCoinSource string
	T                     *testing.T
}

type ChainlinkClient struct {
	NKeys          []client.NodeKeysBundle
	ChainlinkNodes []*client.Chainlink
	bTypeAttr      *client.BridgeTypeAttributes
	bootstrapPeers []client.P2PData
}

// DeployCluster Deploys and sets up config of the environment and nodes
func (testState *Test) DeployCluster() {
	lggr := logger.Nop()
	testState.Cc = &ChainlinkClient{}
	testState.ObservationSource = testState.GetDefaultObservationSource()
	testState.JuelsPerFeeCoinSource = testState.GetDefaultJuelsPerFeeCoinSource()
	testState.DeployEnv()
	testState.SetupClients()
	if testState.Common.Testnet {
		testState.Common.Env.URLs[testState.Common.ServiceKeyL2][1] = testState.Common.L2RPCUrl
	}
	testState.Cc.NKeys, testState.Cc.ChainlinkNodes, err = testState.Common.CreateKeys(testState.Common.Env)
	require.NoError(testState.T, err, "Creating chains and keys should not fail")
	testState.Starknet, err = starknet.NewClient(testState.Common.ChainId, testState.Common.L2RPCUrl, lggr, &rpcRequestTimeout)
	require.NoError(testState.T, err, "Creating starknet client should not fail")
	testState.OCR2Client, err = ocr2.NewClient(testState.Starknet, lggr)
	require.NoError(testState.T, err, "Creating ocr2 client should not fail")
	if !testState.Common.Testnet {
		err = os.Setenv("PRIVATE_KEY", testState.GetDefaultPrivateKey())
		require.NoError(testState.T, err, "Setting private key should not fail")
		err = os.Setenv("ACCOUNT", testState.GetDefaultWalletAddress())
		require.NoError(testState.T, err, "Setting account address should not fail")
		testState.Devnet.AutoDumpState() // Auto dumping devnet state to avoid losing contracts on crash
	}
}

// DeployEnv Deploys the environment
func (testState *Test) DeployEnv() {
	err = testState.Common.Env.Run()
	require.NoError(testState.T, err)
	testState.mockServer, err = ctfClient.ConnectMockServer(testState.Common.Env)
	require.NoError(testState.T, err, "Creating mockserver clients shouldn't fail")
}

// SetupClients Sets up the starknet client
func (testState *Test) SetupClients() {
	if testState.Common.Testnet {
		log.Debug().Msg(fmt.Sprintf("Overriding L2 RPC: %s", testState.Common.L2RPCUrl))
	} else {
		testState.Common.L2RPCUrl = testState.Common.Env.URLs[testState.Common.ServiceKeyL2][0] // For local runs setting local ip
		if testState.Common.InsideK8 {
			testState.Common.L2RPCUrl = testState.Common.Env.URLs[testState.Common.ServiceKeyL2][1] // For remote runner setting remote IP
		}
		testState.Devnet = testState.Devnet.NewStarkNetDevnetClient(testState.Common.L2RPCUrl, dumpPath)
		require.NoError(testState.T, err)
	}
}

// LoadOCR2Config Loads and returns the default starknet gauntlet config
func (testState *Test) LoadOCR2Config() (*ops.OCR2Config, error) {
	var offChaiNKeys []string
	var onChaiNKeys []string
	var peerIds []string
	var txKeys []string
	var cfgKeys []string
	for _, key := range testState.Cc.NKeys {
		offChaiNKeys = append(offChaiNKeys, key.OCR2Key.Data.Attributes.OffChainPublicKey)
		peerIds = append(peerIds, key.PeerID)
		txKeys = append(txKeys, key.TXKey.Data.ID)
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

func (testState *Test) SetUpNodes(mockServerVal int) {
	testState.SetBridgeTypeAttrs(&client.BridgeTypeAttributes{
		Name: "bridge-mockserver",
		URL:  testState.GetMockServerURL(),
	})
	err = testState.SetMockServerValue("", mockServerVal)
	require.NoError(testState.T, err, "Setting mock server value should not fail")
	err = testState.Common.CreateJobsForContract(testState.GetChainlinkClient(), testState.ObservationSource, testState.JuelsPerFeeCoinSource, testState.OCRAddr)
	require.NoError(testState.T, err, "Creating jobs should not fail")
}

// GetStarkNetAddress Returns the local StarkNET address
func (testState *Test) GetStarkNetAddress() string {
	return testState.Common.Env.URLs[testState.Common.ServiceKeyL2][0]
}

// GetStarkNetAddressRemote Returns the remote StarkNET address
func (testState *Test) GetStarkNetAddressRemote() string {
	return testState.Common.Env.URLs[testState.Common.ServiceKeyL2][1]
}

// GetNodeKeys Returns the node key bundles
func (testState *Test) GetNodeKeys() []client.NodeKeysBundle {
	return testState.Cc.NKeys
}

func (testState *Test) GetChainlinkNodes() []*client.Chainlink {
	return testState.Cc.ChainlinkNodes
}

func (testState *Test) GetDefaultPrivateKey() string {
	return defaultWalletPrivKey
}

func (testState *Test) GetDefaultWalletAddress() string {
	return defaultWalletAddress
}

func (testState *Test) GetChainlinkClient() *ChainlinkClient {
	return testState.Cc
}

func (testState *Test) GetStarknetDevnetClient() *devnet.StarkNetDevnetClient {
	return testState.Devnet
}

func (testState *Test) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	testState.Cc.bTypeAttr = attr
}

func (testState *Test) GetMockServerURL() string {
	return testState.mockServer.Config.ClusterURL
}

func (testState *Test) SetMockServerValue(path string, val int) error {
	return testState.mockServer.SetValuePath(path, val)
}

// ConfigureL1Messaging Sets the L1 messaging contract location and RPC url on L2
func (testState *Test) ConfigureL1Messaging() {
	err = testState.Devnet.LoadL1MessagingContract(testState.L1RPCUrl)
	require.NoError(testState.T, err, "Setting up L1 messaging should not fail")
}

func (testState *Test) GetDefaultObservationSource() string {
	return `
			val [type = "bridge" name="bridge-mockserver"]
			parse [type="jsonparse" path="data,result"]
			val -> parse
			`
}

func (testState *Test) GetDefaultJuelsPerFeeCoinSource() string {
	return `"""
			sum  [type="sum" values=<[451000]> ]
			sum
			"""
			`
}

func (testState *Test) ValidateRounds(rounds int, isSoak bool) error {
	ctx := context.Background() // context background used because timeout handled by requestTimeout param
	// assert new rounds are occurring
	details := ocr2.TransmissionDetails{}
	increasing := 0 // track number of increasing rounds
	var stuck bool
	stuckCount := 0
	var positive bool
	var negative bool
	var sign = -1

	// validate balance in aggregator
	resLINK, errLINK := testState.Starknet.CallContract(ctx, starknet.CallOps{
		ContractAddress: caigotypes.HexToHash(testState.LinkTokenAddr),
		Selector:        "balanceOf",
		Calldata:        []string{caigotypes.HexToBN(testState.OCRAddr).String()},
	})
	require.NoError(testState.T, errLINK, "Reader balance from LINK contract should not fail")
	resAgg, errAgg := testState.Starknet.CallContract(ctx, starknet.CallOps{
		ContractAddress: caigotypes.HexToHash(testState.OCRAddr),
		Selector:        "link_available_for_payment",
	})
	require.NoError(testState.T, errAgg, "Reader balance from LINK contract should not fail")
	balLINK, _ := new(big.Int).SetString(resLINK[0], 0)
	balAgg, _ := new(big.Int).SetString(resAgg[0], 0)
	assert.Equal(testState.T, balLINK.Cmp(big.NewInt(0)), 1, "Aggregator should have non-zero balance")
	assert.GreaterOrEqual(testState.T, balLINK.Cmp(balAgg), 0, "Aggregator payment balance should be <= actual LINK balance")

	for start := time.Now(); time.Since(start) < testState.Common.TTL; {
		log.Info().Msg(fmt.Sprintf("Elapsed time: %s, Round wait: %s ", time.Since(start), testState.Common.TTL))
		var res ocr2.TransmissionDetails
		res, err = testState.OCR2Client.LatestTransmissionDetails(ctx, caigotypes.HexToHash(testState.OCRAddr))
		// end condition: enough rounds have occurred, and positive and negative answers have been seen
		if !isSoak && increasing >= rounds && positive && negative {
			break
		}

		// end condition: rounds have been stuck
		if stuck && stuckCount > 50 {
			log.Debug().Msg("failing to fetch transmissions means blockchain may have stopped")
			break
		}

		rand.Seed(time.Now().UnixNano())
		sign *= -1
		mockServerValue = (rand.Intn(900000000-0+1) + 0) * sign
		log.Info().Msg(fmt.Sprintf("Setting adapter value to %d", mockServerValue))
		err = testState.SetMockServerValue("", mockServerValue)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("Setting mock server value error: %+v", err))
		}
		// try to fetch rounds
		time.Sleep(5 * time.Second)

		if err != nil {
			log.Error().Msg(fmt.Sprintf("Transmission Error: %+v", err))
			continue
		}
		log.Info().Msg(fmt.Sprintf("Transmission Details: %+v", res))

		// continue if no changes
		if res.Epoch == 0 && res.Round == 0 {
			continue
		}

		// answer comparison (atleast see a positive and negative value once)
		ansCmp := res.LatestAnswer.Cmp(big.NewInt(0))
		positive = ansCmp == 1 || positive
		negative = ansCmp == -1 || negative

		// if changes from zero values set (should only initially)
		if res.Epoch > 0 && details.Epoch == 0 {
			if !isSoak {
				assert.Greater(testState.T, res.Epoch, details.Epoch)
				assert.GreaterOrEqual(testState.T, res.Round, details.Round)
				assert.NotEqual(testState.T, ansCmp, 0) // assert changed from 0
				assert.NotEqual(testState.T, res.Digest, details.Digest)
				assert.Equal(testState.T, details.LatestTimestamp.Before(res.LatestTimestamp), true)
			}
			details = res
			continue
		}
		// check increasing rounds
		if !isSoak {
			assert.Equal(testState.T, res.Digest, details.Digest, "Config digest should not change")
		} else {
			if res.Digest != details.Digest {
				log.Error().Msg(fmt.Sprintf("Config digest should not change, expected %s got %s", details.Digest, res.Digest))
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
		assert.GreaterOrEqual(testState.T, increasing, rounds, "Round + epochs should be increasing")
		assert.Equal(testState.T, positive, true, "Positive value should have been submitted")
		assert.Equal(testState.T, negative, true, "Negative value should have been submitted")
		assert.Equal(testState.T, stuck, false, "Round + epochs should not be stuck")
	}

	// Test proxy reading
	// TODO: would be good to test proxy switching underlying feeds
	var roundDataRaw []string
	roundDataRaw, err = testState.Starknet.CallContract(ctx, starknet.CallOps{
		ContractAddress: caigotypes.HexToHash(testState.ProxyAddr),
		Selector:        "latest_round_data",
	})
	if !isSoak {
		require.NoError(testState.T, err, "Reading round data from proxy should not fail")
		assert.Equal(testState.T, len(roundDataRaw), 5, "Round data from proxy should match expected size")
	}
	value := starknet.HexToSignedBig(roundDataRaw[1]).Int64()
	if value < 0 {
		assert.Equal(testState.T, value, int64(mockServerValue), "Reading from proxy should return correct value")
	}

	return nil
}
