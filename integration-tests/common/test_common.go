package common

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

var (
	err       error
	clImage   string
	clVersion string

	// These are one of the default addresses based on the seed we pass to devnet which is 0
	defaultWalletPrivKey = ops.PrivateKeys0Seed[0]
	defaultWalletAddress string // derived in init()
	rpcRequestTimeout    = time.Second * 300
	dumpPath             = "/dumps/dump.pkl"
)

func init() {
	// pass in flags to override default chainlink image & version
	flag.StringVar(&clImage, "chainlink-image", "", "specify chainlink image to be used")
	flag.StringVar(&clVersion, "chainlink-version", "", "specify chainlink version to be used")

	// wallet contract derivation
	keyBytes, err := hex.DecodeString(strings.TrimPrefix(defaultWalletPrivKey, "0x"))
	if err != nil {
		panic(err)
	}
	defaultWalletAddress = "0x" + hex.EncodeToString(keys.PubKeyToAccount(keys.Raw(keyBytes).Key().PublicKey(), ops.DevnetClassHash, ops.DevnetSalt))
}

type Test struct {
	Devnet               *devnet.StarkNetDevnetClient
	cc                   *ChainlinkClient
	Starknet             *starknet.Client
	OCR2Client           *ocr2.Client
	Sg                   *gauntlet.StarknetGauntlet
	mockServer           *ctfClient.MockserverClient
	Env                  *environment.Environment
	L1RPCUrl             string
	L2RPCUrl             string
	Common               *Common
	InsideK8s            bool
	LinkTokenAddr        string
	OCRAddr              string
	AccessControllerAddr string
	ProxyAddr            string
	Testnet              bool
}

type StarkNetDevnetClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *blockchain.EVMNetwork
	client *resty.Client
}
type ChainlinkClient struct {
	nKeys          []ctfClient.NodeKeysBundle
	chainlinkNodes []*client.Chainlink
	bTypeAttr      *client.BridgeTypeAttributes
	bootstrapPeers []client.P2PData
}

// DeployCluster Deploys and sets up config of the environment and nodes
func (t *Test) DeployCluster(nodes int, commonConfig *Common) {
	lggr := logger.Nop()
	// Checking if tests need to run on remote runner
	_, t.InsideK8s = os.LookupEnv("INSIDE_K8")
	t.L2RPCUrl, t.Testnet = os.LookupEnv("L2_RPC_URL")
	t.Common = SetConfig(commonConfig)
	t.cc = &ChainlinkClient{}
	t.DeployEnv(nodes)
	t.SetupClients()
	t.cc.nKeys, t.cc.chainlinkNodes, err = t.Common.CreateKeys(t.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating chains and keys should not fail")
	t.Starknet, err = starknet.NewClient(commonConfig.ChainId, t.L2RPCUrl, lggr, &rpcRequestTimeout)
	Expect(err).ShouldNot(HaveOccurred(), "Creating starknet client should not fail")
	t.OCR2Client, err = ocr2.NewClient(t.Starknet, lggr)
	Expect(err).ShouldNot(HaveOccurred(), "Creating ocr2 client should not fail")
	if !t.Testnet {
		err = os.Setenv("PRIVATE_KEY", t.GetDefaultPrivateKey())
		Expect(err).ShouldNot(HaveOccurred(), "Setting private key should not fail")
		err = os.Setenv("ACCOUNT", t.GetDefaultWalletAddress())
		Expect(err).ShouldNot(HaveOccurred(), "Setting account address should not fail")
		t.Devnet.AutoDumpState() // Auto dumping devnet state to avoid losing contracts on crash
	}
}

// DeployEnv Deploys the environment
func (t *Test) DeployEnv(nodes int) {
	clConfig := map[string]interface{}{
		"replicas": nodes,
		"env":      GetDefaultCoreConfig(),
	}

	// if image is specified, include in config data
	// if not, do not set image data - will default to env vars
	if clImage != "" && clVersion != "" {
		clConfig["chainlink"] = map[string]interface{}{
			"image": map[string]interface{}{
				"image":   clImage,
				"version": clVersion,
			},
		}
	}
	t.Env = GetDefaultEnvSetup(&environment.Config{InsideK8s: t.InsideK8s}, clConfig)
	err := t.Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
	t.mockServer, err = ctfClient.ConnectMockServer(t.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")
}

// SetupClients Sets up the starknet client
func (t *Test) SetupClients() {
	if t.Testnet {
		log.Debug().Msg(fmt.Sprintf("Overriding L2 RPC: %s", t.L2RPCUrl))
	} else {
		t.L2RPCUrl = t.Env.URLs[t.Common.ServiceKeyL2][0] // For local runs setting local ip
		if t.InsideK8s {
			t.L2RPCUrl = t.Env.URLs[t.Common.ServiceKeyL2][1] // For remote runner setting remote IP
		}
		t.Devnet = t.Devnet.NewStarkNetDevnetClient(t.L2RPCUrl, dumpPath)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

// LoadOCR2Config Loads and returns the default starknet gauntlet config
func (t *Test) LoadOCR2Config() (*ops.OCR2Config, error) {
	var offChainKeys []string
	var onChainKeys []string
	var peerIds []string
	var txKeys []string
	var cfgKeys []string
	for _, key := range t.cc.nKeys {
		offChainKeys = append(offChainKeys, key.OCR2Key.Data.Attributes.OffChainPublicKey)
		peerIds = append(peerIds, key.PeerID)
		txKeys = append(txKeys, key.TXKey.Data.ID)
		onChainKeys = append(onChainKeys, key.OCR2Key.Data.Attributes.OnChainPublicKey)
		cfgKeys = append(cfgKeys, key.OCR2Key.Data.Attributes.ConfigPublicKey)
	}

	var payload = ops.TestOCR2Config
	payload.Signers = onChainKeys
	payload.Transmitters = txKeys
	payload.OffchainConfig.OffchainPublicKeys = offChainKeys
	payload.OffchainConfig.PeerIds = peerIds
	payload.OffchainConfig.ConfigPublicKeys = cfgKeys

	return &payload, nil
}

// GetStarkNetAddress Returns the local StarkNET address
func (t *Test) GetStarkNetAddress() string {
	return t.Env.URLs[t.Common.ServiceKeyL2][0]
}

// GetStarkNetAddressRemote Returns the remote StarkNET address
func (t *Test) GetStarkNetAddressRemote() string {
	return t.Env.URLs[t.Common.ServiceKeyL2][1]
}

// GetNodeKeys Returns the node key bundles
func (t *Test) GetNodeKeys() []ctfClient.NodeKeysBundle {
	return t.cc.nKeys
}

func (t *Test) GetChainlinkNodes() []*client.Chainlink {
	return t.cc.chainlinkNodes
}

func (t *Test) GetDefaultPrivateKey() string {
	return defaultWalletPrivKey
}

func (t *Test) GetDefaultWalletAddress() string {
	return defaultWalletAddress
}

func (t *Test) GetChainlinkClient() *ChainlinkClient {
	return t.cc
}

func (t *Test) GetStarknetDevnetClient() *devnet.StarkNetDevnetClient {
	return t.Devnet
}

func (t *Test) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	t.cc.bTypeAttr = attr
}

func (t *Test) GetMockServerURL() string {
	return t.mockServer.Config.ClusterURL
}

func (t *Test) SetMockServerValue(path string, val int) error {
	return t.mockServer.SetValuePath(path, val)
}

// ConfigureL1Messaging Sets the L1 messaging contract location and RPC url on L2
func (t *Test) ConfigureL1Messaging() {
	err = t.Devnet.LoadL1MessagingContract(t.L1RPCUrl)
	Expect(err).ShouldNot(HaveOccurred(), "Setting up L1 messaging should not fail")
}
