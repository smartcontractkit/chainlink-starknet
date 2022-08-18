package common

import (
	"context"
	"encoding/hex"
	"flag"
	"strings"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

const (
	ChainName = "starknet"
	ChainId   = gateway.GOERLI_ID // default chainID for local devnet and live testnet
)

var (
	err       error
	clImage   string
	clVersion string

	// These are one of the default addresses based on the seed we pass to devnet which is 0
	defaultWalletPrivKey = ops.PrivateKeys0Seed[0]
	defaultWalletAddress string // derived in init()
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
	sc         *StarkNetDevnetClient
	cc         *ChainlinkClient
	mockServer *ctfClient.MockserverClient
	Env        *environment.Environment
	Common     *Common
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
	t.Common = SetConfig(commonConfig)
	t.cc = &ChainlinkClient{}
	t.sc = &StarkNetDevnetClient{}
	t.DeployEnv(nodes)
	t.SetupClients()
	t.cc.nKeys, t.cc.chainlinkNodes, err = t.Common.CreateKeys(t.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating chains and keys should not fail")
}

// DeployEnv Deploys the environment
func (t *Test) DeployEnv(nodes int) {
	clConfig := map[string]interface{}{
		"replicas": nodes,
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

	t.Env = environment.New(&environment.Config{
		NamespacePrefix: "chainlink-smoke-ocr-starknet-ci",
		InsideK8s:       false,
	}).
		// AddHelm(hardhat.New(nil)).
		AddHelm(devnet.New(nil)).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(chainlink.New(0, clConfig))
	err := t.Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
	t.mockServer, err = ctfClient.ConnectMockServer(t.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")
}

// SetupClients Sets up the starknet client
func (t *Test) SetupClients() {
	t.sc = t.NewStarkNetDevnetClient(&blockchain.EVMNetwork{
		Name: t.Common.ServiceKeyL2,
		URL:  t.Env.URLs[t.Common.ServiceKeyL2][1],
		PrivateKeys: []string{
			defaultWalletPrivKey,
		},
	})

	Expect(err).ShouldNot(HaveOccurred())
}

func (t *Test) NewStarkNetDevnetClient(cfg *blockchain.EVMNetwork) *StarkNetDevnetClient {
	ctx, cancel := context.WithCancel(context.Background())
	t.sc = &StarkNetDevnetClient{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		client: resty.New().SetBaseURL(t.GetStarkNetAddress()),
	}
	return t.sc
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

// GetStarkNetName Returns the config name for StarkNET
func (t *Test) GetStarkNetName() string {
	return t.sc.cfg.Name
}

// GetStarkNetAddress Returns the local StarkNET address
func (t *Test) GetStarkNetAddress() string {
	return t.Env.URLs[t.Common.ServiceKeyL2][0]
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

func (t *Test) GetStarknetDevnetClient() *StarkNetDevnetClient {
	return t.sc
}

func (t *Test) GetStarknetDevnetCfg() *blockchain.EVMNetwork {
	return t.sc.cfg
}

func (t *Test) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	t.cc.bTypeAttr = attr
}

// ConfigureL1Messaging Sets the L1 messaging contract location and RPC url on L2
func (t *Test) ConfigureL1Messaging() {
	err := devnet.LoadL1MessagingContract()
	Expect(err).ShouldNot(HaveOccurred(), "Setting up L1 messaging should not fail")
}
