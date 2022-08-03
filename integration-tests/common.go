package integration_tests

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

const (
	gethClient      = "Hardhat"
	devnetClient    = "starknet-dev"
	chainLinkClient = "chainlink"
	ChainName       = "starknet"
	ChainId         = "devnet"
)

var (
	err             error
	Env             *environment.Environment
	sc              *StarkNetClient
	chainlinkNodes  []*client.Chainlink
	mockServer      *ctfClient.MockserverClient
	nodeKeys        []NodeKeysBundle
	starknetNetwork *blockchain.EVMNetwork
	bTypeAttr       *client.BridgeTypeAttributes
)

type StarkNetClient struct {
	ctx            context.Context
	cancel         context.CancelFunc
	cfg            *blockchain.EVMNetwork
	urls           []*url.URL
	client         *resty.Client
	nKeys          []NodeKeysBundle
	chainlinkNodes []*client.Chainlink
}

// ContractNodeInfo contains the indexes of the nodes, bridges, NodeKeyBundles and nodes relevant to an OCR2 Contract
type ContractNodeInfo struct {
	OCR2 *OCRv2
	// Store                   *solclient.Store
	BootstrapNodeIdx        int
	BootstrapNode           client.Chainlink
	BootstrapNodeKeysBundle NodeKeysBundle
	BootstrapBridgeInfo     BridgeInfo
	NodesIdx                []int
	Nodes                   []client.Chainlink
	NodeKeysBundle          []NodeKeysBundle
	// BridgeInfos             []BridgeInfo
}

type OCRv2 struct {
	Client           *client.Chainlink
	ContractDeployer *ContractDeployer
}

type ContractDeployer struct {
	Client *client.Chainlink
	Env    *environment.Environment
}

type BridgeInfo struct {
	ObservationSource string
	JuelsSource       string
}

type Chart struct {
	Name   string
	Path   string
	Values *map[string]interface{}
}

type NodeKeysBundle struct {
	OCR2Key client.OCR2Key
	PeerID  string
	TXKey   client.TxKey
	P2PKeys client.P2PKeys
}

type Node struct {
	ID        int32     `json:"ID"`
	Name      string    `json:"Name"`
	ChainID   string    `json:"ChainID"`
	URL       string    `json:"URL"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

func NewStarkNetClient(cfg *blockchain.EVMNetwork) (*StarkNetClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &StarkNetClient{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		client: resty.New().SetBaseURL(GetStarkNetAddress()),
	}
	// Currently not needed since we are testing on L2
	// if err := c.init(); err != nil {
	// 	return nil, err
	// }
	return c, nil
}

func DeployCluster(nodes int) *StarkNetClient {
	DeployEnv(nodes)
	SetupClients()
	sc.CreateKeys()
	return sc
}

func DeployEnv(nodes int) {
	Env = environment.New(&environment.Config{
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
					"version": "custom.6e31f3079442a3c50980a4929110aff64489f4cd",
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
				"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:6690",
				"P2PV2_DELTA_DIAL":            "5s",
				"P2PV2_DELTA_RECONCILE":       "5s",
				"p2p_listen_port":             "0",
			},
		}))
	err := Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
	mockServer, err = ctfClient.ConnectMockServer(Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")

}

func SetupClients() {
	sc, err = NewStarkNetClient(&blockchain.EVMNetwork{
		Name:    "starknet-dev",
		URL:     Env.URLs[devnetClient][1],
		ChainID: 13337,
		PrivateKeys: []string{
			"c4da537c1651ddae44867db30d67b366",
		},
	})
	Expect(err).ShouldNot(HaveOccurred())
}

func (sc *StarkNetClient) CreateKeys() {
	starknetNetwork := &blockchain.EVMNetwork{
		Name:    "starknet",
		ChainID: 1337,
		PrivateKeys: []string{
			"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
		},
		ChainlinkTransactionLimit: 500000,
		Timeout:                   2 * time.Minute,
		MinimumConfirmations:      1,
		GasEstimationBuffer:       10000,
	}
	starknetNetwork.URLs = Env.URLs[devnetClient]
	chainlinkNodes, err = client.ConnectChainlinkNodes(Env)
	Expect(err).ShouldNot(HaveOccurred(), "Connecting to chainlink nodes shouldn't fail")
	sc.nKeys, err = CreateNodeKeysBundle(chainlinkNodes)
	sc.chainlinkNodes = chainlinkNodes
	Expect(err).ShouldNot(HaveOccurred(), "Creating key bundles should not fail")
	for _, n := range sc.chainlinkNodes {
		_, _, err = n.CreateStarknetChain(&client.StarknetChainAttributes{
			Type:    ChainName,
			ChainID: ChainId,
			Config:  client.StarknetChainConfig{},
		})
		Expect(err).ShouldNot(HaveOccurred(), "Creating starknet chain should not fail")
		_, _, err = n.CreateStarknetNode(&client.StarknetNodeAttributes{
			Name:    ChainName,
			ChainID: ChainId,
			Url:     Env.URLs[devnetClient][1],
		})
		Expect(err).ShouldNot(HaveOccurred(), "Creating starknet node should not fail")
	}

}

func (s *StarkNetClient) init() error {
	resp, err := s.client.R().SetBody(map[string]interface{}{
		"networkUrl": Env.URLs[gethClient][1],
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	Expect(err).ShouldNot(HaveOccurred())
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}

func (s *StarkNetClient) autoSyncL1() {
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

func CreateNodeKeysBundle(nodes []*client.Chainlink) ([]NodeKeysBundle, error) {
	nkb := make([]NodeKeysBundle, 0)
	for _, n := range nodes {
		p2pkeys, err := n.MustReadP2PKeys()
		if err != nil {
			return nil, err
		}

		peerID := p2pkeys.Data[0].Attributes.PeerID
		txKey, _, err := n.CreateTxKey(ChainName)
		if err != nil {
			return nil, err
		}

		ocrKey, _, err := n.CreateOCR2Key(ChainName)
		if err != nil {
			return nil, err
		}
		nkb = append(nkb, NodeKeysBundle{
			PeerID:  peerID,
			OCR2Key: *ocrKey,
			TXKey:   *txKey,
			P2PKeys: *p2pkeys,
		})

	}

	return nkb, nil
}

func (s *StarkNetClient) FundAccounts(l2AccList []string) {
	for _, key := range l2AccList {
		_, err := s.client.R().SetBody(map[string]interface{}{
			"address": key,
			"amount":  500000,
		}).Post("/mint")
		Expect(err).ShouldNot(HaveOccurred(), "Funding accounts should not fail")
	}
}

func (s *StarkNetClient) CreateJobsForContract(ocrControllerAddress string) error {
	relayConfig := map[string]string{
		"nodeName": fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.NewV4().String()),
		"chainID":  ChainId,
	}

	// Setting up bootstrap node
	jobSpec := &OCR2TaskJobSpec{
		Name:        fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.NewV4().String()),
		JobType:     "bootstrap",
		ContractID:  ocrControllerAddress,
		Relay:       ChainName,
		RelayConfig: relayConfig,
	}
	_, _, err := sc.chainlinkNodes[0].CreateJob(jobSpec)
	Expect(err).ShouldNot(HaveOccurred(), "Creating bootstrap job should not fail")

	// Defining bridge
	bTypeAttr = &client.BridgeTypeAttributes{
		Name: "bridge-cryptocompare",
		URL:  "https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=BTC,USD",
	}
	for nIdx, n := range s.chainlinkNodes {
		if nIdx == 0 {
			continue
		}
		// Creating bridge
		n.CreateBridge(bTypeAttr)
		// Creating OCRv2 Job
		jobSpec := &OCR2TaskJobSpec{
			Name:        fmt.Sprintf("starknet-OCRv2-%d-%s", nIdx, uuid.NewV4().String()),
			JobType:     "offchainreporting2",
			ContractID:  ocrControllerAddress,
			Relay:       ChainName,
			RelayConfig: relayConfig,
			PluginType:  "median",
			P2pPeerID:   s.nKeys[nIdx].PeerID,
			P2pBootstrapPeers: []P2PData{
				P2PData{
					RemoteIP:   s.chainlinkNodes[0].RemoteIP(),
					RemotePort: "8090",
					PeerID:     s.nKeys[0].PeerID,
				},
			},
			OCRKeyBundleID:    s.nKeys[nIdx].OCR2Key.Data.ID,
			TransmitterID:     s.nKeys[nIdx].TXKey.Data.ID,
			ObservationSource: client.ObservationSourceSpecBridge(*bTypeAttr),
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

func GetStarkNetName() string {
	return sc.cfg.Name
}

func GetStarkNetAddress() string {
	return Env.URLs[devnetClient][0]
}

func GetNodeKeys() []NodeKeysBundle {
	return sc.nKeys
}
