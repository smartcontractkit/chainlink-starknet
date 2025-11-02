package txm

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"go.opentelemetry.io/otel/metric"
)

var (
	promNumBroadcastedTxs    *prometheus.CounterVec
	promNumConfirmedTxs      *prometheus.CounterVec
	promNumNonceGaps         *prometheus.CounterVec
	promReachedMaxAttempts   *prometheus.GaugeVec
	promTimeUntilTxConfirmed *prometheus.HistogramVec
	promEnqueueFailed        *prometheus.CounterVec

	beholderNumBroadcastedTxs    metric.Int64Counter
	beholderNumConfirmedTxs      metric.Int64Counter
	beholderNumNonceGaps         metric.Int64Counter
	beholderReachedMaxAttempts   metric.Int64Gauge
	beholderTimeUntilTxConfirmed metric.Float64Histogram
	beholderEnqueueFailed        metric.Int64Counter

	metricsOnce sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		// Initialize Prometheus metrics
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

		promEnqueueFailed = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_enqueue_failed",
			Help: "Total number of times transaction enqueue failed due to queue being full or other issues.",
		}, []string{"chainID"})

		// Initialize beholder metrics
		beholderNumBroadcastedTxs, _ = beholder.GetMeter().Int64Counter("txm_num_broadcasted_transactions")
		beholderNumConfirmedTxs, _ = beholder.GetMeter().Int64Counter("txm_num_confirmed_transactions")
		beholderNumNonceGaps, _ = beholder.GetMeter().Int64Counter("txm_num_nonce_gaps")
		beholderTimeUntilTxConfirmed, _ = beholder.GetMeter().Float64Histogram("txm_time_until_tx_confirmed")
		beholderReachedMaxAttempts, _ = beholder.GetMeter().Int64Gauge("txm_reached_max_attempts")
		beholderEnqueueFailed, _ = beholder.GetMeter().Int64Counter("txm_enqueue_failed")
	})
}

// prometheusMetrics implements TxMetrics using Prometheus
type prometheusMetrics struct {
	chainID              string
	numBroadcastedTxs    metric.Int64Counter
	numConfirmedTxs      metric.Int64Counter
	numNonceGaps         metric.Int64Counter
	reachedMaxAttempts   metric.Int64Gauge
	timeUntilTxConfirmed metric.Float64Histogram
	enqueueFailed        metric.Int64Counter
}

func NewPrometheusMetrics(chainID string) TxMetrics {
	initMetrics()

	return &prometheusMetrics{
		chainID:              chainID,
		numBroadcastedTxs:    beholderNumBroadcastedTxs,
		numConfirmedTxs:      beholderNumConfirmedTxs,
		numNonceGaps:         beholderNumNonceGaps,
		reachedMaxAttempts:   beholderReachedMaxAttempts,
		timeUntilTxConfirmed: beholderTimeUntilTxConfirmed,
		enqueueFailed:        beholderEnqueueFailed,
	}
}

func (m *prometheusMetrics) IncrementNumBroadcastedTxs(ctx context.Context) {
	promNumBroadcastedTxs.WithLabelValues(m.chainID).Inc()
	m.numBroadcastedTxs.Add(ctx, 1)
}

func (m *prometheusMetrics) IncrementNumConfirmedTxs(ctx context.Context, confirmedTransactions int) {
	promNumConfirmedTxs.WithLabelValues(m.chainID).Add(float64(confirmedTransactions))
	m.numConfirmedTxs.Add(ctx, int64(confirmedTransactions))
}

func (m *prometheusMetrics) IncrementNumNonceGaps(ctx context.Context) {
	promNumNonceGaps.WithLabelValues(m.chainID).Inc()
	m.numNonceGaps.Add(ctx, 1)
}

func (m *prometheusMetrics) ReachedMaxAttempts(ctx context.Context, reached bool) {
	var value float64
	if reached {
		value = 1
	}
	promReachedMaxAttempts.WithLabelValues(m.chainID).Set(value)
	m.reachedMaxAttempts.Record(ctx, int64(value))
}

func (m *prometheusMetrics) RecordTimeUntilTxConfirmed(ctx context.Context, duration float64) {
	promTimeUntilTxConfirmed.WithLabelValues(m.chainID).Observe(duration)
	m.timeUntilTxConfirmed.Record(ctx, duration)
}

func (m *prometheusMetrics) IncrementEnqueueFailed(ctx context.Context) {
	promEnqueueFailed.WithLabelValues(m.chainID).Inc()
	m.enqueueFailed.Add(ctx, 1)
}
