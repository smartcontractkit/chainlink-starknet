package common

import (
	"context"
	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	ops "github.com/smartcontractkit/chainlink-starknet/relayer/ops"
	starknet "github.com/smartcontractkit/chainlink-starknet/relayer/ops/devnet"
	blockchain "github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"strings"
	"time"
)

const (
	ChainName = "starknet"
	ChainId   = "devnet"
	// These are one of the default addresses based on the seed we pass to devnet which is 123
	defaultWalletAddress = "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502"
	defaultWalletPrivKey = "0xc4da537c1651ddae44867db30d67b366"
)

var (
	err error
)

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

// OCR2Config Default config for OCR2 for starknet
type OCR2Config struct {
	F                     int             `json:"f"`
	Signers               []string        `json:"signers"`
	Transmitters          []string        `json:"transmitters"`
	OnchainConfig         string          `json:"onchainConfig"`
	OffchainConfig        *OffchainConfig `json:"offchainConfig"`
	OffchainConfigVersion int             `json:"offchainConfigVersion"`
	Secret                string          `json:"secret"`
}

type OffchainConfig struct {
	DeltaProgressNanoseconds                           int64                  `json:"deltaProgressNanoseconds"`
	DeltaResendNanoseconds                             int64                  `json:"deltaResendNanoseconds"`
	DeltaRoundNanoseconds                              int64                  `json:"deltaRoundNanoseconds"`
	DeltaGraceNanoseconds                              int                    `json:"deltaGraceNanoseconds"`
	DeltaStageNanoseconds                              int64                  `json:"deltaStageNanoseconds"`
	RMax                                               int                    `json:"rMax"`
	S                                                  []int                  `json:"s"`
	OffchainPublicKeys                                 []string               `json:"offchainPublicKeys"`
	PeerIds                                            []string               `json:"peerIds"`
	ReportingPluginConfig                              *ReportingPluginConfig `json:"reportingPluginConfig"`
	MaxDurationQueryNanoseconds                        int                    `json:"maxDurationQueryNanoseconds"`
	MaxDurationObservationNanoseconds                  int                    `json:"maxDurationObservationNanoseconds"`
	MaxDurationReportNanoseconds                       int                    `json:"maxDurationReportNanoseconds"`
	MaxDurationShouldAcceptFinalizedReportNanoseconds  int                    `json:"maxDurationShouldAcceptFinalizedReportNanoseconds"`
	MaxDurationShouldTransmitAcceptedReportNanoseconds int                    `json:"maxDurationShouldTransmitAcceptedReportNanoseconds"`
}

type ReportingPluginConfig struct {
	AlphaReportInfinite bool `json:"alphaReportInfinite"`
	AlphaReportPpb      int  `json:"alphaReportPpb"`
	AlphaAcceptInfinite bool `json:"alphaAcceptInfinite"`
	AlphaAcceptPpb      int  `json:"alphaAcceptPpb"`
	DeltaCNanoseconds   int  `json:"deltaCNanoseconds"`
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
	t.Env = environment.New(&environment.Config{
		NamespacePrefix: "smoke-ocr-starknet",
		TTL:             3 * time.Hour,
		InsideK8s:       false,
	}).
		// AddHelm(hardhat.New(nil)).
		AddHelm(ops.New(nil)).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
			"replicas": nodes,
			"chainlink": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   "795953128386.dkr.ecr.us-west-2.amazonaws.com/chainlink",
					"version": "custom.2205e48ec7979b34fbc7a15ec2234bd16ca35122",
				},
			},
			"env": map[string]interface{}{
				"STARKNET_ENABLED":            "true",
				"EVM_ENABLED":                 "false",
				"EVM_RPC_ENABLED":             "false",
				"CHAINLINK_DEV":               "false",
				"FEATURE_OFFCHAIN_REPORTING2": "true",
				"feature_offchain_reporting":  "false",
				"P2P_NETWORKING_STACK":        "V2",
				"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:8090",
				"P2PV2_DELTA_DIAL":            "5s",
				"P2PV2_DELTA_RECONCILE":       "5s",
				"p2p_listen_port":             "0",
			},
		}))
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
func (t *Test) LoadOCR2Config() (*OCR2Config, error) {
	var offChainKeys []string
	var onChainKeys []string
	var peerIds []string
	var txKeys []string
	for _, key := range t.cc.nKeys {
		offChainKeys = append(offChainKeys, strings.Replace(key.OCR2Key.Data.Attributes.OffChainPublicKey, "ocr2off_starknet_", "", 1))
		peerIds = append(peerIds, key.PeerID)
		txKeys = append(txKeys, key.TXKey.Data.ID)
		onChainKeys = append(onChainKeys, "0x"+strings.Replace(key.OCR2Key.Data.Attributes.OnChainPublicKey, "ocr2on_starknet_", "", 1))
	}

	var payload = &OCR2Config{
		F:             1,
		Signers:       onChainKeys,
		Transmitters:  txKeys,
		OnchainConfig: "",
		OffchainConfig: &OffchainConfig{
			DeltaProgressNanoseconds: 8000000000,
			DeltaResendNanoseconds:   30000000000,
			DeltaRoundNanoseconds:    3000000000,
			DeltaGraceNanoseconds:    500000000,
			DeltaStageNanoseconds:    20000000000,
			RMax:                     5,
			S:                        []int{1, 2},
			OffchainPublicKeys:       offChainKeys,
			PeerIds:                  peerIds,
			ReportingPluginConfig: &ReportingPluginConfig{
				AlphaReportInfinite: false,
				AlphaReportPpb:      0,
				AlphaAcceptInfinite: false,
				AlphaAcceptPpb:      0,
				DeltaCNanoseconds:   0,
			},
			MaxDurationQueryNanoseconds:                        0,
			MaxDurationObservationNanoseconds:                  1000000000,
			MaxDurationReportNanoseconds:                       200000000,
			MaxDurationShouldAcceptFinalizedReportNanoseconds:  200000000,
			MaxDurationShouldTransmitAcceptedReportNanoseconds: 200000000,
		},
		OffchainConfigVersion: 2,
		Secret:                "awe accuse polygon tonic depart acuity onyx inform bound gilbert expire",
	}

	return payload, nil
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

func (t *Test) SetBridgeTypeAttrs(attr *client.BridgeTypeAttributes) {
	t.cc.bTypeAttr = attr
}

// ConfigureL1Messaging Sets the L1 messaging contract location and RPC url on L2
func (t *Test) ConfigureL1Messaging() {
	err := starknet.LoadL1MessagingContract()
	Expect(err).ShouldNot(HaveOccurred(), "Setting up L1 messaging should not fail")
}
