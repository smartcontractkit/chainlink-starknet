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
	"github.com/smartcontractkit/chainlink-starknet/ops"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
)

const (
	gethClient   = "Hardhat"
	devnetClient = "starknet-dev"
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

type Chart struct {
	Name   string
	Path   string
	Values *map[string]interface{}
}

type NodeKeysBundle struct {
	OCR2Key client.OCR2Key
	PeerID  string
	TXKey   client.TxKey
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
		client: resty.New().SetBaseURL(cfg.URL),
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
		URL:     "0.0.0.0:5000",
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

func getDevnetUrl() string {
	return sc.cfg.Name
}

func CreateNodeKeysBundle(nodes []client.Chainlink) ([]NodeKeysBundle, error) {
	nkb := make([]NodeKeysBundle, 0)
	for _, n := range nodes {
		p2pkeys, err := n.ReadP2PKeys()
		if err != nil {
			return nil, err
		}

		peerID := p2pkeys.Data[0].Attributes.PeerID
		txKey, err := n.CreateTxKey("eth")
		if err != nil {
			return nil, err
		}

		ocrKey, err := n.CreateOCR2Key("evm")
		if err != nil {
			return nil, err
		}
		nkb = append(nkb, NodeKeysBundle{
			PeerID:  peerID,
			OCR2Key: *ocrKey,
			TXKey:   *txKey,
		})

	}

	return nkb, nil
}

func GetNodeKeys() []NodeKeysBundle {
	return sc.nKeys
}
