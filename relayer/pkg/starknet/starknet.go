package starknet

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	ops "github.com/smartcontractkit/chainlink-starknet/relayer/ops"
	blockchain "github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

const (
	serviceKeyL1        = "Hardhat"
	serviceKeyL2        = "starknet-dev"
	serviceKeyChainlink = "chainlink"
	ChainName           = "starknet"
	ChainId             = "devnet"
	// These are one of the the default addresses based on the seed we pass to devnet which is 123
	defaultWalletAddress = "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502"
	defaultWalletPrivKey = "0xc4da537c1651ddae44867db30d67b366"
)

var (
	err error
	t   *Test
)

type Test struct {
	sc         *StarkNetDevnetClient
	cc         *ChainlinkClient
	mockServer *ctfClient.MockserverClient
	Env        *environment.Environment
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

func (t *Test) NewStarkNetDevnetClient(cfg *blockchain.EVMNetwork) (*StarkNetDevnetClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	t.sc = &StarkNetDevnetClient{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		client: resty.New().SetBaseURL(t.GetStarkNetAddress()),
	}
	return t.sc, nil
}

// Deploys and sets up config of the environment and nodes
func (t *Test) DeployCluster(nodes int) *Test {
	t = &Test{
		sc: &StarkNetDevnetClient{},
		cc: &ChainlinkClient{},
	}
	t.DeployEnv(nodes)
	t.SetupClients()
	t.CreateKeys()
	return t
}

// Deploys the environment
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
					"version": "custom.73db21cf91b5608d20ef0aa15d49f825cd49f0b6",
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

// Sets up the starknet client
func (t *Test) SetupClients() {
	t.sc, err = t.NewStarkNetDevnetClient(&blockchain.EVMNetwork{
		Name: serviceKeyL2,
		URL:  t.Env.URLs[serviceKeyL2][1],
		PrivateKeys: []string{
			defaultWalletPrivKey,
		},
	})

	Expect(err).ShouldNot(HaveOccurred())
}

// Creates node keys and defines chain and nodes for each node
func (t *Test) CreateKeys() {
	t.cc.chainlinkNodes, err = client.ConnectChainlinkNodes(t.Env)
	Expect(err).ShouldNot(HaveOccurred(), "Connecting to chainlink nodes shouldn't fail")
	t.cc.nKeys, err = ctfClient.CreateNodeKeysBundle(t.cc.chainlinkNodes, ChainName)
	Expect(err).ShouldNot(HaveOccurred(), "Creating key bundles should not fail")
	for _, n := range t.cc.chainlinkNodes {
		_, _, err = n.CreateStarknetChain(&client.StarknetChainAttributes{
			Type:    ChainName,
			ChainID: ChainId,
			Config:  client.StarknetChainConfig{},
		})
		Expect(err).ShouldNot(HaveOccurred(), "Creating starknet chain should not fail")
		_, _, err = n.CreateStarknetNode(&client.StarknetNodeAttributes{
			Name:    ChainName,
			ChainID: ChainId,
			Url:     t.Env.URLs[serviceKeyL2][1],
		})
		Expect(err).ShouldNot(HaveOccurred(), "Creating starknet node should not fail")
	}

}

func (s *StarkNetDevnetClient) loadL1MesagingContract() error {
	resp, err := s.client.R().SetBody(map[string]interface{}{
		"networkUrl": t.Env.URLs[serviceKeyL1][1],
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	Expect(err).ShouldNot(HaveOccurred())
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}

func (s *StarkNetDevnetClient) autoSyncL1() {
	t := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				log.Debug().Msg("Shutting down L1 sync")
				return
			case <-t.C:
				log.Debug().Msg("Syncing L1")
				_, err := s.client.R().Post("/postman/flush")
				if err != nil {
					log.Error().Err(err).Msg("failed to sync L1")
				}
			}
		}
	}()
}

// Funds provided accounts with 500000 eth each
func (t *Test) FundAccounts(l2AccList []string) {
	for _, key := range l2AccList {
		_, err := t.sc.client.R().SetBody(map[string]interface{}{
			"address": key,
			"amount":  500000,
		}).Post("/mint")
		Expect(err).ShouldNot(HaveOccurred(), "Funding accounts should not fail")
	}
}

//
func (t *Test) CreateJobsForContract(ocrControllerAddress string) error {
	// Defining bootstrap peers
	for nIdx, n := range t.cc.chainlinkNodes {
		t.cc.bootstrapPeers = append(t.cc.bootstrapPeers, client.P2PData{
			RemoteIP:   n.RemoteIP(),
			RemotePort: "8090",
			PeerID:     t.cc.nKeys[nIdx].PeerID,
		})
	}
	// Defining relay config
	relayConfig := map[string]string{
		"nodeName": fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.NewV4().String()),
		"chainID":  ChainId,
	}

	// Setting up bootstrap node
	jobSpec := &client.OCR2TaskJobSpec{
		Name:               fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.NewV4().String()),
		JobType:            "bootstrap",
		ContractID:         ocrControllerAddress,
		Relay:              ChainName,
		RelayConfig:        relayConfig,
		P2PV2Bootstrappers: t.cc.bootstrapPeers,
	}
	_, _, err := t.cc.chainlinkNodes[0].CreateJob(jobSpec)
	Expect(err).ShouldNot(HaveOccurred(), "Creating bootstrap job should not fail")

	// Defining bridge
	t.cc.bTypeAttr = &client.BridgeTypeAttributes{
		Name: "bridge-cryptocompare",
		URL:  "https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=BTC,USD",
	}

	// Setting up job specs
	for nIdx, n := range t.cc.chainlinkNodes {
		if nIdx == 0 {
			continue
		}
		n.CreateBridge(t.cc.bTypeAttr)
		jobSpec := &client.OCR2TaskJobSpec{
			Name:               fmt.Sprintf("starknet-OCRv2-%d-%s", nIdx, uuid.NewV4().String()),
			JobType:            "offchainreporting2",
			ContractID:         ocrControllerAddress,
			Relay:              ChainName,
			RelayConfig:        relayConfig,
			PluginType:         "median",
			P2PV2Bootstrappers: t.cc.bootstrapPeers,
			OCRKeyBundleID:     t.cc.nKeys[nIdx].OCR2Key.Data.ID,
			TransmitterID:      t.cc.nKeys[nIdx].TXKey.Data.ID,
			ObservationSource:  client.ObservationSourceSpecBridge(*t.cc.bTypeAttr),
			JuelsPerFeeCoinSource: ` // Fetch the LINK price from a data source
			// data source 1
			ds1_link       [type="bridge" name="bridge-cryptocompare"]
			ds1_link_parse [type="jsonparse" path="BTC"]
			ds1_link -> ds1_link_parse -> divide
			// Fetch the ETH price from a data source
			// data source 1
			ds1_coin       [type="bridge" name="bridge-cryptocompare"]
			ds1_coin_parse [type="jsonparse" path="BTC"]
			ds1_coin -> ds1_coin_parse -> divide
			divide [type="divide" input="$(ds1_coin_parse)" divisor="$(ds1_link_parse)" precision="18"]
			scale  [type="multiply" times=1000000000000000000]
			divide -> scale`,
		}
		_, _, err := n.CreateJob(jobSpec)
		Expect(err).ShouldNot(HaveOccurred(), "Creating node job should not fail")

	}
	return nil
}

// Returns the config name for StarkNET
func (t *Test) GetStarkNetName() string {
	return t.sc.cfg.Name
}

// Returns the local StarkNET address
func (t *Test) GetStarkNetAddress() string {
	return t.Env.URLs[serviceKeyL2][0]
}

// Returns the node key bundles
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

// Sets the L1 messaging contract location and RPC url on L2
func (t *Test) ConfigureL1Messaging() error {
	//Currently not needed since we are testing on L2
	if err := t.sc.loadL1MesagingContract(); err != nil {
		return err
	}
	return nil
}
