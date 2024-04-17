package monitoring

import (
	"context"
	"math/big"
	"sync"

	commonMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

func NewNodeBalancesExporterFactory(log commonMonitoring.Logger, metrics Metrics) commonMonitoring.ExporterFactory {
	return &nodeBalancesExporterFactory{
		log,
		metrics,
	}
}

type nodeBalancesExporterFactory struct {
	log     commonMonitoring.Logger
	metrics Metrics
}

func (f *nodeBalancesExporterFactory) NewExporter(params commonMonitoring.ExporterParams) (commonMonitoring.Exporter, error) {
	return &nodeBalancesExporter{
		log:         f.log,
		metrics:     f.metrics,
		chainConfig: params.ChainConfig,
		addrsSet:    []ContractAddressWithBalance{},
	}, nil
}

type nodeBalancesExporter struct {
	log         commonMonitoring.Logger
	metrics     Metrics
	chainConfig commonMonitoring.ChainConfig
	addrsSet    []ContractAddressWithBalance
	addrsMu     sync.Mutex
}

func (e *nodeBalancesExporter) Export(ctx context.Context, data interface{}) {
	balanceEnvelope, isBalanceEnvelope := data.(BalanceEnvelope)
	if !isBalanceEnvelope {
		return
	}

	decimals := balanceEnvelope.Decimals
	divisor := new(big.Int).Exp(new(big.Int).SetUint64(10), decimals, nil) // 10^(decimals)

	for _, c := range balanceEnvelope.Contracts {
		balanceAns := new(big.Int).Div(c.Balance, divisor)

		e.metrics.SetBalance(
			toFloat64(balanceAns),
			c.Address.String(),
			c.Name,
			e.chainConfig.GetNetworkID(),
			e.chainConfig.GetNetworkName(),
			e.chainConfig.GetChainID())
	}

	e.addrsMu.Lock()
	defer e.addrsMu.Unlock()

	e.addrsSet = balanceEnvelope.Contracts

}

func (e *nodeBalancesExporter) Cleanup(_ context.Context) {
	e.addrsMu.Lock()
	defer e.addrsMu.Unlock()

	for _, c := range e.addrsSet {
		e.metrics.CleanupBalance(c.Address.String(), c.Name, e.chainConfig.GetNetworkID(), e.chainConfig.GetNetworkName(), e.chainConfig.GetChainID())
	}
}

func toFloat64(bignum *big.Int) float64 {
	val, _ := new(big.Float).SetInt(bignum).Float64()
	return val
}
