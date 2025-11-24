package txm

import (
	"context"
	"math/big"
	"sync"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	promNumBroadcastedTxs    *prometheus.CounterVec
	promNumConfirmedTxs      *prometheus.CounterVec
	promNumNonceGaps         *prometheus.CounterVec
	promTimeUntilTxConfirmed *prometheus.HistogramVec
	promEnqueueFailed        *prometheus.CounterVec
	promNonceRebroadcast     *prometheus.CounterVec
	promNextNonce            *prometheus.GaugeVec

	beholderNumBroadcastedTxs    metric.Int64Counter
	beholderNumConfirmedTxs      metric.Int64Counter
	beholderNumNonceGaps         metric.Int64Counter
	beholderTimeUntilTxConfirmed metric.Float64Histogram
	beholderEnqueueFailed        metric.Int64Counter
	beholderNonceRebroadcast     metric.Int64Counter
	beholderNextNonce            metric.Int64Gauge

	metricsOnce sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		// Initialize Prometheus metrics
		promNumBroadcastedTxs = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_num_broadcasted_transactions",
			Help: "Total number of successful broadcasted transactions.",
		}, []string{"chainID", "accountAddress"})

		promNumConfirmedTxs = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_num_confirmed_transactions",
			Help: "Total number of confirmed transactions. Note that this can happen multiple times per transaction in the case of re-orgs or when filling the nonce for untracked transactions.",
		}, []string{"chainID", "accountAddress"})

		promNumNonceGaps = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_num_nonce_gaps",
			Help: "Total number of nonce gaps created that the transaction manager had to fill.",
		}, []string{"chainID", "accountAddress"})

		promTimeUntilTxConfirmed = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "txm_time_until_tx_confirmed",
			Help: "The amount of time elapsed from a transaction being broadcast to being included in a block.",
		}, []string{"chainID", "accountAddress"})

		promEnqueueFailed = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_enqueue_failed",
			Help: "Total number of times transaction enqueue failed due to queue being full or other issues.",
		}, []string{"chainID", "accountAddress"})

		promNonceRebroadcast = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "txm_nonce_rebroadcast",
			Help: "Total number of times a nonce was rebroadcasted. This indicates resyncs or nonce gaps. Increments each time a nonce is broadcasted more than once.",
		}, []string{"chainID", "accountAddress"})

		promNextNonce = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "txm_next_nonce",
			Help: "The next nonce that will be used for the account. Updated when transactions are broadcasted or nonce is resynced.",
		}, []string{"chainID", "accountAddress"})

		// Initialize beholder metrics - these are optional and may not be available in all environments
		// Errors are ignored as beholder may not be initialized yet or may not be available
		meter := beholder.GetMeter()
		beholderNumBroadcastedTxs, _ = meter.Int64Counter("txm_num_broadcasted_transactions")
		beholderNumConfirmedTxs, _ = meter.Int64Counter("txm_num_confirmed_transactions")
		beholderNumNonceGaps, _ = meter.Int64Counter("txm_num_nonce_gaps")
		beholderTimeUntilTxConfirmed, _ = meter.Float64Histogram("txm_time_until_tx_confirmed")
		beholderEnqueueFailed, _ = meter.Int64Counter("txm_enqueue_failed")
		beholderNonceRebroadcast, _ = meter.Int64Counter("txm_nonce_rebroadcast")
		beholderNextNonce, _ = meter.Int64Gauge("txm_next_nonce")
	})
}

// prometheusMetrics implements TxMetrics using Prometheus
type prometheusMetrics struct {
	chainID              string
	numBroadcastedTxs    metric.Int64Counter
	numConfirmedTxs      metric.Int64Counter
	numNonceGaps         metric.Int64Counter
	timeUntilTxConfirmed metric.Float64Histogram
	enqueueFailed        metric.Int64Counter
	nonceRebroadcast     metric.Int64Counter
	nextNonce            metric.Int64Gauge
}

// Helper to create consistent metric attributes
func (m *prometheusMetrics) attributes(accountAddress string) metric.MeasurementOption {
	return metric.WithAttributes(
		attribute.String("chainID", m.chainID),
		attribute.String("accountAddress", accountAddress),
	)
}

func NewTxmMetrics(chainID string) TxMetrics {
	initMetrics()

	return &prometheusMetrics{
		chainID:              chainID,
		numBroadcastedTxs:    beholderNumBroadcastedTxs,
		numConfirmedTxs:      beholderNumConfirmedTxs,
		numNonceGaps:         beholderNumNonceGaps,
		timeUntilTxConfirmed: beholderTimeUntilTxConfirmed,
		enqueueFailed:        beholderEnqueueFailed,
		nonceRebroadcast:     beholderNonceRebroadcast,
		nextNonce:            beholderNextNonce,
	}
}

func (m *prometheusMetrics) IncrementNumBroadcastedTxs(ctx context.Context, accountAddress string) {
	promNumBroadcastedTxs.WithLabelValues(m.chainID, accountAddress).Inc()
	if m.numBroadcastedTxs != nil {
		m.numBroadcastedTxs.Add(ctx, 1, m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) IncrementNumConfirmedTxs(ctx context.Context, accountAddress string, confirmedTransactions int) {
	promNumConfirmedTxs.WithLabelValues(m.chainID, accountAddress).Add(float64(confirmedTransactions))
	if m.numConfirmedTxs != nil {
		m.numConfirmedTxs.Add(ctx, int64(confirmedTransactions), m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) IncrementNumNonceGaps(ctx context.Context, accountAddress string) {
	promNumNonceGaps.WithLabelValues(m.chainID, accountAddress).Inc()
	if m.numNonceGaps != nil {
		m.numNonceGaps.Add(ctx, 1, m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) RecordTimeUntilTxConfirmed(ctx context.Context, accountAddress string, duration float64) {
	promTimeUntilTxConfirmed.WithLabelValues(m.chainID, accountAddress).Observe(duration)
	if m.timeUntilTxConfirmed != nil {
		m.timeUntilTxConfirmed.Record(ctx, duration, m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) IncrementEnqueueFailed(ctx context.Context, accountAddress string) {
	promEnqueueFailed.WithLabelValues(m.chainID, accountAddress).Inc()
	if m.enqueueFailed != nil {
		m.enqueueFailed.Add(ctx, 1, m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) IncrementNonceRebroadcast(ctx context.Context, accountAddress string) {
	promNonceRebroadcast.WithLabelValues(m.chainID, accountAddress).Inc()
	if m.nonceRebroadcast != nil {
		m.nonceRebroadcast.Add(ctx, 1, m.attributes(accountAddress))
	}
}

func (m *prometheusMetrics) UpdateNextNonceMetric(ctx context.Context, accountAddress string, nonce *felt.Felt) {
	// Convert felt.Felt to float64 for Prometheus (it's a big.Int internally)
	nonceBigInt := nonce.BigInt(new(big.Int))
	nonceFloat := float64(nonceBigInt.Int64())

	promNextNonce.WithLabelValues(m.chainID, accountAddress).Set(nonceFloat)

	// Beholder uses Int64Gauge with attributes for per-account tracking
	if m.nextNonce != nil {
		m.nextNonce.Record(ctx, nonceBigInt.Int64(), m.attributes(accountAddress))
	}
}
