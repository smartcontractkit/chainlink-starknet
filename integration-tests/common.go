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
	geth "github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	starknet "github.com/smartcontractkit/chainlink-env/pkg/helm/starknet"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink-testing-framework/contracts"
)

const (
	gethClient   = "Simulated Geth"
	devnetClient = "starknet-dev"
)

var (
	err               error
	Env               *environment.Environment
	sc                *StarkNetClient
	contractDeployer  contracts.ContractDeployer
	chainClient       blockchain.EVMClient
	linkTokenContract contracts.LinkToken
	chainlinkNodes    []client.Chainlink
	mockServer        *client.MockserverClient
	ocrInstances      []contracts.OffchainAggregator
	clientFunc        func(*environment.Environment) (blockchain.EVMClient, error)
)

// StarkNetNetworkConfig StarkNet network config
type StarkNetNetworkConfig struct {
	ContractsDeployed         bool          `mapstructure:"contracts_deployed" yaml:"contracts_deployed"`
	L1BridgeAddr              string        `mapstructure:"l1_bridge_addr" yaml:"l1_bridge_addr"`
	Name                      string        `mapstructure:"name" yaml:"name"`
	ChainID                   int64         `mapstructure:"chain_id" yaml:"chain_id"`
	URL                       string        `mapstructure:"url" yaml:"url"`
	URLs                      []string      `mapstructure:"urls" yaml:"urls"`
	Type                      string        `mapstructure:"type" yaml:"type"`
	PrivateKeys               []string      `mapstructure:"private_keys" yaml:"private_keys"`
	ChainlinkTransactionLimit uint64        `mapstructure:"chainlink_transaction_limit" yaml:"chainlink_transaction_limit"`
	Timeout                   time.Duration `mapstructure:"transaction_timeout" yaml:"transaction_timeout"`
	MinimumConfirmations      int           `mapstructure:"minimum_confirmations" yaml:"minimum_confirmations"`
	BlockGasLimit             uint64        `mapstructure:"block_gas_limit" yaml:"block_gas_limit"`
	WalletAddress             string
}

type StarkNetClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *StarkNetNetworkConfig
	urls   []*url.URL
	client *resty.Client
}

type Chart struct {
	Name   string
	Path   string
	Values *map[string]interface{}
}

func NewStarkNetClient(cfg *StarkNetNetworkConfig) (*StarkNetClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &StarkNetClient{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		client: resty.New().SetBaseURL(cfg.URL),
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func DeployCluster(nodes int) {
	DeployEnv(nodes)
	SetupClients()
	DeployContracts()
}

func DeployEnv(nodes int) {
	Env = environment.New(&environment.Config{
		NamespacePrefix: "smoke-ocr-starknet",
		TTL:             3 * time.Hour,
		InsideK8s:       false,
	}).
		AddHelm(geth.New(nil)).
		AddHelm(starknet.New(nil)).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(chainlink.New(0, map[string]interface{}{
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
		}))
	err := Env.Run()
	Expect(err).ShouldNot(HaveOccurred())
}

func SetupClients() {
	sc, err = NewStarkNetClient(&StarkNetNetworkConfig{
		ContractsDeployed: false,
		L1BridgeAddr:      "",
		Name:              "devnet",
		URL:               Env.URLs[devnetClient][0],
		URLs:              Env.URLs[gethClient],
		Type:              "starknet",
		PrivateKeys: []string{
			"0xc4da537c1651ddae44867db30d67b366",
		},
		WalletAddress: "0x6e3205f9b7c4328f00f718fdecf56ab31acfb3cd6ffeb999dcbac41236ea502",
	})
	Expect(err).ShouldNot(HaveOccurred())
	err = sc.init()
	Expect(err).ShouldNot(HaveOccurred())
	sc.autoSyncL1()
}

func DeployContracts() {
	chainClient, err = blockchain.NewEthereumMultiNodeClientSetup(blockchain.SimulatedEVMNetwork)(Env)
	Expect(err).ShouldNot(HaveOccurred(), "Connecting to blockchain nodes shouldn't fail")
	//contractDeployer, err = contracts.NewContractDeployer(chainClient)
	// Expect(err).ShouldNot(HaveOccurred(), "Deploying contracts shouldn't fail")

	// chainlinkNodes, err = client.ConnectChainlinkNodes(Env)
	// Expect(err).ShouldNot(HaveOccurred(), "Connecting to chainlink nodes shouldn't fail")
	// mockServer, err = client.ConnectMockServer(Env)
	// Expect(err).ShouldNot(HaveOccurred(), "Creating mockserver clients shouldn't fail")

	// chainClient.ParallelTransactions(true)
	// Expect(err).ShouldNot(HaveOccurred())

	// linkTokenContract, err = contractDeployer.DeployLinkTokenContract()
	// Expect(err).ShouldNot(HaveOccurred(), "Deploying Link Token Contract shouldn't fail")
	// ocrInstances = actions.DeployOCRContracts(1, linkTokenContract, contractDeployer, chainlinkNodes, chainClient)
	// err = chainClient.WaitForEvents()
	// Expect(err).ShouldNot(HaveOccurred())
}

func (s *StarkNetClient) init() error {
	resp, err := s.client.R().SetBody(map[string]interface{}{
		"networkUrl": Env.URLs[gethClient][3],
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

// func (s *StarkNetClient) Get() interface{} {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GetNetworkName() string {
// 	return "starknet-dev"
// }

// func (s *StarkNetClient) GetNetworkType() string {
// 	return "l2_starknet_dev"
// }

// func (s *StarkNetClient) GetChainID() *big.Int {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GetClients() []blockchain.EVMClient {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GetDefaultWallet() *blockchain.EthereumWallet {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GetWallets() []*blockchain.EthereumWallet {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GetNetworkConfig() *config.ETHNetwork {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) SetID(id int) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) SetDefaultWallet(num int) error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) SetWallets(wallets []*blockchain.EthereumWallet) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) LoadWallets(ns interface{}) error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) SwitchNode(node int) error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) HeaderHashByNumber(ctx context.Context, bn *big.Int) (string, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) HeaderTimestampByNumber(ctx context.Context, bn *big.Int) (uint64, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) LatestBlockNumber(ctx context.Context) (uint64, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) Fund(toAddress string, amount *big.Float) error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) DeployContract(contractName string, deployer blockchain.ContractDeployer) (*common.Address, *types.Transaction, interface{}, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) TransactionOpts(from *blockchain.EthereumWallet) (*bind.TransactOpts, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) ProcessTransaction(tx *types.Transaction) error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) IsTxConfirmed(txHash common.Hash) (bool, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) ParallelTransactions(enabled bool) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) Close() error {
// 	s.cancel()
// 	return nil
// }

// func (s *StarkNetClient) EstimateCostForChainlinkOperations(amountOfOperations int) (*big.Float, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) EstimateTransactionGasCost() (*big.Int, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) GasStats() *blockchain.GasStats {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) AddHeaderEventSubscription(key string, subscriber blockchain.HeaderEventSubscription) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) DeleteHeaderEventSubscription(key string) {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (s *StarkNetClient) WaitForEvents() error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func GetStarkNetClient(
// 	_ string,
// 	networkConfig map[string]interface{},
// 	urls []*url.URL,
// ) (blockchain.EVMClient, error) {
// 	networkSettings := &StarkNetNetworkConfig{}
// 	err := blockchain.UnmarshalNetworkConfig(networkConfig, networkSettings)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Info().
// 		Interface("URLs", networkSettings.URLs).
// 		Msg("Connecting StarkNet client")
// 	return NewStarkNetClient(networkSettings, urls)
// }
