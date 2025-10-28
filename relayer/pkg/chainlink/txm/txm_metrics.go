package txm

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promNumBroadcastedTxs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "txm_num_broadcasted_transactions",
		Help: "Total number of successful broadcasted transactions.",
	}, []string{"chainID"})
	promNumConfirmedTxs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "txm_num_confirmed_transactions",
		Help: "Total number of confirmed transactions. Note that this can happen multiple times per transaction in the case of re-orgs or when filling the nonce for untracked transactions.",
	}, []string{"chainID"})
	promNumNonceGaps = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "txm_num_nonce_gaps",
		Help: "Total number of nonce gaps created that the transaction manager had to fill.",
	}, []string{"chainID"})
	promReachedMaxAttempts = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "txm_reached_max_attempts",
		Help: "A gauge that is treated as boolean; 1 if the condition is true, 0 otherwise. Controls whether the TXM has reached max attempts threshold or not.",
	}, []string{"chainID"})
	promTimeUntilTxConfirmed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "txm_time_until_tx_confirmed",
		Help: "The amount of time elapsed from a transaction being broadcast to being included in a block.",
	}, []string{"chainID"})
	metricsOnce sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		// Metrics are already initialized via promauto.NewCounterVec above
		// This function exists for consistency with the pattern
	})
}

// prometheusMetrics implements TxMetrics using Prometheus
type prometheusMetrics struct {
	chainID string
}

func NewPrometheusMetrics(chainID string) TxMetrics {
	initMetrics()
	return &prometheusMetrics{chainID: chainID}
}

func (m *prometheusMetrics) IncrementNumBroadcastedTxs(ctx context.Context) {
	initMetrics()
	promNumBroadcastedTxs.WithLabelValues(m.chainID).Inc()
}

func (m *prometheusMetrics) IncrementNumConfirmedTxs(ctx context.Context, confirmedTransactions int) {
	initMetrics()
	promNumConfirmedTxs.WithLabelValues(m.chainID).Add(float64(confirmedTransactions))
}

func (m *prometheusMetrics) IncrementNumNonceGaps(ctx context.Context) {
	initMetrics()
	promNumNonceGaps.WithLabelValues(m.chainID).Inc()
}

func (m *prometheusMetrics) ReachedMaxAttempts(ctx context.Context, reached bool) {
	initMetrics()
	var value float64
	if reached {
		value = 1
	}
	promReachedMaxAttempts.WithLabelValues(m.chainID).Set(value)
}

func (m *prometheusMetrics) RecordTimeUntilTxConfirmed(ctx context.Context, duration float64) {
	initMetrics()
	promTimeUntilTxConfirmed.WithLabelValues(m.chainID).Observe(duration)
}
