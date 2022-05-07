package integration_tests

import (
	"context"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-testing-framework/blockchain"
	"github.com/smartcontractkit/chainlink-testing-framework/config"
	"github.com/smartcontractkit/helmenv/environment"
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
}

// GetStarkNetURLs gets remote L1 URL and a local L2 URL
func GetStarkNetURLs(e *environment.Environment) ([]*url.URL, error) {
	var urls []*url.URL
	l2URLs, err := e.Charts.Connections("starknet").LocalURLsByPort("http", environment.HTTP)
	if err != nil {
		return nil, err
	}
	urls = append(urls, l2URLs...)
	l1URLs, err := e.Charts.Connections("geth").RemoteURLsByPort("http-rpc", environment.HTTP)
	if err != nil {
		return nil, err
	}
	urls = append(urls, l1URLs...)
	return urls, nil
}

func NewStarkNetClient(cfg *StarkNetNetworkConfig, urls []*url.URL) (*StarkNetClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &StarkNetClient{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		urls:   urls,
		client: resty.New().SetBaseURL(urls[0].String()),
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	c.autoSyncL1()
	return c, nil
}

type StarkNetClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *StarkNetNetworkConfig
	urls   []*url.URL
	client *resty.Client
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

func (s *StarkNetClient) init() error {
	resp, err := s.client.R().SetBody(map[string]interface{}{
		"networkUrl": s.urls[1].String(),
		"address":    s.cfg.L1BridgeAddr,
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}

func (s *StarkNetClient) Get() interface{} {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GetNetworkName() string {
	return "starknet-dev"
}

func (s *StarkNetClient) GetNetworkType() string {
	return "l2_starknet_dev"
}

func (s *StarkNetClient) GetChainID() *big.Int {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GetClients() []blockchain.EVMClient {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GetDefaultWallet() *blockchain.EthereumWallet {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GetWallets() []*blockchain.EthereumWallet {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GetNetworkConfig() *config.ETHNetwork {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) SetID(id int) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) SetDefaultWallet(num int) error {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) SetWallets(wallets []*blockchain.EthereumWallet) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) LoadWallets(ns interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) SwitchNode(node int) error {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) HeaderHashByNumber(ctx context.Context, bn *big.Int) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) HeaderTimestampByNumber(ctx context.Context, bn *big.Int) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) LatestBlockNumber(ctx context.Context) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) Fund(toAddress string, amount *big.Float) error {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) DeployContract(contractName string, deployer blockchain.ContractDeployer) (*common.Address, *types.Transaction, interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) TransactionOpts(from *blockchain.EthereumWallet) (*bind.TransactOpts, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) ProcessTransaction(tx *types.Transaction) error {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) IsTxConfirmed(txHash common.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) ParallelTransactions(enabled bool) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) Close() error {
	s.cancel()
	return nil
}

func (s *StarkNetClient) EstimateCostForChainlinkOperations(amountOfOperations int) (*big.Float, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) EstimateTransactionGasCost() (*big.Int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) GasStats() *blockchain.GasStats {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) AddHeaderEventSubscription(key string, subscriber blockchain.HeaderEventSubscription) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) DeleteHeaderEventSubscription(key string) {
	//TODO implement me
	panic("implement me")
}

func (s *StarkNetClient) WaitForEvents() error {
	//TODO implement me
	panic("implement me")
}

func GetStarkNetClient(
	_ string,
	networkConfig map[string]interface{},
	urls []*url.URL,
) (blockchain.EVMClient, error) {
	networkSettings := &StarkNetNetworkConfig{}
	err := blockchain.UnmarshalNetworkConfig(networkConfig, networkSettings)
	if err != nil {
		return nil, err
	}
	log.Info().
		Interface("URLs", networkSettings.URLs).
		Msg("Connecting StarkNet client")
	return NewStarkNetClient(networkSettings, urls)
}
