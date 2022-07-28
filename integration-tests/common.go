package integration_tests

import (
	"context"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	ops "github.com/smartcontractkit/chainlink-starknet/relayer/ops"
	blockchain "github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
)

const (
	gethClient      = "Hardhat"
	devnetClient    = "starknet-dev"
	chainLinkClient = "chainlink"
	ChainName       = "starknet"
)

var (
	err             error
	Env             *environment.Environment
	sc              *StarkNetClient
	chainlinkNodes  []client.Chainlink
	mockServer      *client.MockserverClient
	nodeKeys        []NodeKeysBundle
	starknetNetwork *blockchain.EVMNetwork
)

type StarkNetClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *blockchain.EVMNetwork
	urls   []*url.URL
	client *resty.Client
	nKeys  []NodeKeysBundle
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
	CreateKeys()
	return sc
}

func DeployEnv(nodes int) {

	// nodeStruct := &Node{
	// 	ID:      0,
	// 	Name:    "starknet-devnet",
	// 	ChainID: "13337",
	// 	URL:     "0.0.0.0:5000",
	// }

	// jsonData := `
	// [
	// {
	// 		"ID": "0"
	// 		"name": "devnet",
	// 		"ChainID": "13337",
	// 		"URL": "0.0.0.0:5000"
	// 	}
	// ]`

	//b, _ := json.Marshal(jsonData)

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
					"version": "develop-root",
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
				//"STARKNET_NODES":              b,
			},
		}))
	err := Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
	mockServer, err = client.ConnectMockServer(Env)
	Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")

}

func SetupClients() {
	sc, err = NewStarkNetClient(&blockchain.EVMNetwork{
		Name:    "starknet-dev",
		URL:     "http://0.0.0.0:5000",
		ChainID: 13337,
		PrivateKeys: []string{
			"c4da537c1651ddae44867db30d67b366",
		},
		ChainlinkTransactionLimit: 500000,
		Timeout:                   2 * time.Minute,
		MinimumConfirmations:      1,
		GasEstimationBuffer:       10000,
	})
	Expect(err).ShouldNot(HaveOccurred())
}

func CreateKeys() {
	starknetNetwork := blockchain.SimulatedEVMNetworkStarknet
	starknetNetwork.URLs = Env.URLs[devnetClient]
	chainlinkNodes, err = client.ConnectChainlinkNodes(Env)
	Expect(err).ShouldNot(HaveOccurred(), "Connecting to chainlink nodes shouldn't fail")
	sc.nKeys, err = CreateNodeKeysBundle(chainlinkNodes)
	Expect(err).ShouldNot(HaveOccurred(), "Creating key bundles should not fail")
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

func CreateNodeKeysBundle(nodes []client.Chainlink) ([]NodeKeysBundle, error) {
	nkb := make([]NodeKeysBundle, 0)
	for _, n := range nodes {
		p2pkeys, err := n.ReadP2PKeys()
		if err != nil {
			return nil, err
		}

		peerID := p2pkeys.Data[0].Attributes.PeerID
		txKey, err := n.CreateTxKey(ChainName)
		if err != nil {
			return nil, err
		}

		ocrKey, err := n.CreateOCR2Key(ChainName)
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

// TODO - Adding creation of Job Specs
// func (s *StarkNetClient) CreateJobsForContract(contractNodeInfo *ContractNodeInfo) error {
// 	relayConfig := map[string]string{
// 		"nodeEndpointHTTP": "http://0.0.0.0:8899",
// 		"ocr2ProgramID":    s.nKeys[0].OCR2Key.Data.ID,
// 		"transmissionsID":  s.nKeys[0].TXKey.Data.ID,
// 		//	"storeProgramID":   contractNodeInfo.Store.ProgramAddress(),
// 		"chainID": "starknet",
// 	}
// 	bootstrapPeers := []client.P2PData{
// 		{
// 			RemoteIP:   Env.URLs[chainLinkClient][0],
// 			RemotePort: "6690",
// 			PeerID:     contractNodeInfo.BootstrapNodeKeysBundle.PeerID,
// 		},
// 	}
// 	jobSpec := &client.OCR2TaskJobSpec{
// 		Name:                  fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.NewV4().String()),
// 		JobType:               "bootstrap",
// 		ContractID:            s.nKeys[0].OCR2Key.Data.ID,
// 		Relay:                 ChainName,
// 		RelayConfig:           relayConfig,
// 		PluginType:            "median",
// 		P2PV2Bootstrappers:    bootstrapPeers,
// 		OCRKeyBundleID:        contractNodeInfo.BootstrapNodeKeysBundle.OCR2Key.Data.ID,
// 		TransmitterID:         contractNodeInfo.BootstrapNodeKeysBundle.TXKey.Data.ID,
// 		ObservationSource:     contractNodeInfo.BootstrapBridgeInfo.ObservationSource,
// 		JuelsPerFeeCoinSource: contractNodeInfo.BootstrapBridgeInfo.JuelsSource,
// 		TrackerPollInterval:   15 * time.Second, // faster config checking
// 	}
// 	if _, err := contractNodeInfo.BootstrapNode.CreateJob(jobSpec); err != nil {
// 		return fmt.Errorf("failed creating job for boostrap node: %w", err)
// 	}
// 	for nIdx, n := range contractNodeInfo.Nodes {
// 		jobSpec := &client.OCR2TaskJobSpec{
// 			Name:                  fmt.Sprintf("sol-OCRv2-%d-%s", nIdx, uuid.NewV4().String()),
// 			JobType:               "offchainreporting2",
// 			ContractID:            contractNodeInfo.OCR2.Address(),
// 			Relay:                 ChainName,
// 			RelayConfig:           relayConfig,
// 			PluginType:            "median",
// 			P2PV2Bootstrappers:    bootstrapPeers,
// 			OCRKeyBundleID:        contractNodeInfo.NodeKeysBundle[nIdx].OCR2Key.Data.ID,
// 			TransmitterID:         contractNodeInfo.NodeKeysBundle[nIdx].TXKey.Data.ID,
// 			ObservationSource:     contractNodeInfo.BridgeInfos[nIdx].ObservationSource,
// 			JuelsPerFeeCoinSource: contractNodeInfo.BridgeInfos[nIdx].JuelsSource,
// 			TrackerPollInterval:   15 * time.Second, // faster config checking
// 		}
// 		if _, err := n.CreateJob(jobSpec); err != nil {
// 			return fmt.Errorf("failed creating job for node %s: %w", n.URL(), err)
// 		}
// 	}
// 	return nil
// }

func GetStarkNetName() string {
	return sc.cfg.Name
}

func GetStarkNetAddress() string {
	return Env.URLs[devnetClient][0]
}

func GetNodeKeys() []NodeKeysBundle {
	return sc.nKeys
}
