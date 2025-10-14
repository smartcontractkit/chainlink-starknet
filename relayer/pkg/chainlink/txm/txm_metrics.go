package txm

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promNumSuccessfulTxs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_manager_num_successful_transactions",
		Help: "Total number of successful transactions. Note that this can err to be too high since transactions are counted on each confirmation, which can happen multiple times per transaction in the case of re-orgs",
	}, []string{"chainID"})

	promRevertedTxCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_manager_num_tx_reverted",
		Help: "Number of times a transaction reverted on-chain. Note that this can err to be too high since transactions are counted on each confirmation, which can happen multiple times per transaction in the case of re-orgs",
	}, []string{"chainID"})

	promTxAttemptCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tx_manager_tx_attempt_count",
		Help: "The number of transaction attempts that are currently being processed by the transaction manager",
	}, []string{"chainID"})

	promNumFinalizedTxs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tx_manager_num_finalized_transactions",
		Help: "Total number of finalized transactions (accepted on L1 or L2)",
	}, []string{"chainID"})
)

// prometheusMetrics implements TxMetrics using Prometheus
type prometheusMetrics struct{}

func NewPrometheusMetrics() TxMetrics {
	return &prometheusMetrics{}
}

func (m *prometheusMetrics) IncrementSuccessfulTransactions(chainID string) {
	promNumSuccessfulTxs.WithLabelValues(chainID).Inc()
}

func (m *prometheusMetrics) IncrementRevertedTransactions(chainID string) {
	promRevertedTxCount.WithLabelValues(chainID).Inc()
}

func (m *prometheusMetrics) IncrementFinalizedTransactions(chainID string) {
	promNumFinalizedTxs.WithLabelValues(chainID).Inc()
}

func (m *prometheusMetrics) SetTxAttemptCount(chainID string, count int) {
	promTxAttemptCount.WithLabelValues(chainID).Set(float64(count))
}
