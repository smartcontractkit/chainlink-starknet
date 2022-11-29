package common

import (
	"context"
	"encoding/hex"
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
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
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
	observationSource    = `
			val [type = "bridge" name="bridge-mockserver"]
			parse [type="jsonparse" path="data,result"]
			val -> parse
			`
	juelsPerFeeCoinSource = `"""
			sum  [type="sum" values=<[451000]> ]
			sum
			"""
			`
)

func init() {
	// wallet contract derivation
	keyBytes, err := hex.DecodeString(strings.TrimPrefix(defaultWalletPrivKey, "0x"))
	if err != nil {
		panic(err)
	}
	defaultWalletAddress = "0x" + hex.EncodeToString(keys.PubKeyToAccount(keys.Raw(keyBytes).Key().PublicKey(), ops.DevnetClassHash, ops.DevnetSalt))

}

type Test struct {
	Devnet               *devnet.StarkNetDevnetClient
	Cc                   *ChainlinkClient
	Starknet             *starknet.Client
	OCR2Client           *ocr2.Client
	Sg                   *gauntlet.StarknetGauntlet
	mockServer           *ctfClient.MockserverClient
	L1RPCUrl             string
	Common               *Common
	LinkTokenAddr        string
	OCRAddr              string
	AccessControllerAddr string
	ProxyAddr            string
}

type StarkNetDevnetClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *blockchain.EVMNetwork
	client *resty.Client
}
type ChainlinkClient struct {
	NKeys          []client.NodeKeysBundle
	ChainlinkNodes []*client.Chainlink
	bTypeAttr      *client.BridgeTypeAttributes
	bootstrapPeers []client.P2PData
}

// DeployCluster Deploys and sets up config of the environment and nodes
func (t *Test) DeployCluster() {
	lggr := logger.Nop()
	t.Cc = &ChainlinkClient{}
	t.DeployEnv()
	t.SetupClients()
	if t.Common.Testnet {
		t.Common.Env.URLs[t.Common.ServiceKeyL2][1] = t.Common.L2RPCUrl
	}
	t.Cc.NKeys, t.Cc.ChainlinkNodes, err = t.Common.CreateKeys(t.Common.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating chains and keys should not fail")
	t.Starknet, err = starknet.NewClient(t.Common.ChainId, t.Common.L2RPCUrl, lggr, &rpcRequestTimeout)
	Expect(err).ShouldNot(HaveOccurred(), "Creating starknet client should not fail")
	t.OCR2Client, err = ocr2.NewClient(t.Starknet, lggr)
	Expect(err).ShouldNot(HaveOccurred(), "Creating ocr2 client should not fail")
	if !t.Common.Testnet {
		err = os.Setenv("PRIVATE_KEY", t.GetDefaultPrivateKey())
		Expect(err).ShouldNot(HaveOccurred(), "Setting private key should not fail")
		err = os.Setenv("ACCOUNT", t.GetDefaultWalletAddress())
		Expect(err).ShouldNot(HaveOccurred(), "Setting account address should not fail")
		t.Devnet.AutoDumpState() // Auto dumping devnet state to avoid losing contracts on crash
	}
}

// DeployEnv Deploys the environment
func (t *Test) DeployEnv() {
	err = t.Common.Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
	t.mockServer, err = ctfClient.ConnectMockServer(t.Common.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")
}

// SetupClients Sets up the starknet client
func (t *Test) SetupClients() {
	if t.Common.Testnet {
		log.Debug().Msg(fmt.Sprintf("Overriding L2 RPC: %s", t.Common.L2RPCUrl))
	} else {
		t.Common.L2RPCUrl = t.Common.Env.URLs[t.Common.ServiceKeyL2][0] // For local runs setting local ip
		if t.Common.InsideK8 {
			t.Common.L2RPCUrl = t.Common.Env.URLs[t.Common.ServiceKeyL2][1] // For remote runner setting remote IP
		}
		t.Devnet = t.Devnet.NewStarkNetDevnetClient(t.Common.L2RPCUrl, dumpPath)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

// LoadOCR2Config Loads and returns the default starknet gauntlet config
func (t *Test) LoadOCR2Config() (*ops.OCR2Config, error) {
	var offChaiNKeys []string
	var onChaiNKeys []string
	var peerIds []string
	var txKeys []string
	var cfgKeys []string
	for _, key := range t.Cc.NKeys {
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

func (t *Test) SetUpNodes(mockServerVal int) {
	t.SetBridgeTypeAttrs(&client.BridgeTypeAttributes{
		Name: "bridge-mockserver",
		URL:  t.GetMockServerURL(),
	})
	err = t.SetMockServerValue("", mockServerVal)
	Expect(err).ShouldNot(HaveOccurred(), "Setting mock server value should not fail")
	err = t.Common.CreateJobsForContract(t.GetChainlinkClient(), observationSource, juelsPerFeeCoinSource, t.OCRAddr)
	Expect(err).ShouldNot(HaveOccurred(), "Creating jobs should not fail")
}

// GetStarkNetAddress Returns the local StarkNET address
func (t *Test) GetStarkNetAddress() string {
	return t.Common.Env.URLs[t.Common.ServiceKeyL2][0]
}

// GetStarkNetAddressRemote Returns the remote StarkNET address
func (t *Test) GetStarkNetAddressRemote() string {
	return t.Common.Env.URLs[t.Common.ServiceKeyL2][1]
}

// GetNodeKeys Returns the node key bundles
func (t *Test) GetNodeKeys() []client.NodeKeysBundle {
	return t.Cc.NKeys
}

func (t *Test) GetChainlinkNodes() []*client.Chainlink {
	return t.Cc.ChainlinkNodes
}

func (t *Test) GetDefaultPrivateKey() string {
	return defaultWalletPrivKey
}

func (t *Test) GetDefaultWalletAddress() string {
	return defaultWalletAddress
}

func (t *Test) GetChainlinkClient() *ChainlinkClient {
	return t.Cc
}

func (t *Test) GetStarknetDevnetClient() *devnet.StarkNetDevnetClient {
	return t.Devnet
}

func (t *Test) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	t.Cc.bTypeAttr = attr
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
