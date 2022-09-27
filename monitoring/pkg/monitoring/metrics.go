package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
)

// Metrics is an interface for prometheus metrics. Makes testing easier.
type Metrics interface {
	SetProxyAnswersRaw(answer float64, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string)
	SetProxyAnswers(answer float64, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string)
	Cleanup(proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string)
}

var (
	proxyAnswersRaw = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_answers_raw",
			Help: "Reports the latest raw answer from the proxy contract.",
		},
		[]string{"proxy_contract_address", "feed_id", "chain_id", "contract_status", "contract_type", "feed_name", "feed_path", "network_id", "network_name"},
	)
	proxyAnswers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_answers",
			Help: "Reports the latest answer from the proxy contract divided by the feed's multiplier parameter.",
		},
		[]string{"proxy_contract_address", "feed_id", "chain_id", "contract_status", "contract_type", "feed_name", "feed_path", "network_id", "network_name"},
	)
)

// NewMetrics does wisott
func NewMetrics(log relayMonitoring.Logger) Metrics {
	return &defaultMetrics{log}
}

type defaultMetrics struct {
	log relayMonitoring.Logger
}

func (d *defaultMetrics) SetProxyAnswersRaw(answer float64, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
	proxyAnswersRaw.With(prometheus.Labels{
		"proxy_contract_address": proxyContractAddress,
		"feed_id":                feedID,
		"chain_id":               chainID,
		"contract_status":        contractStatus,
		"contract_type":          contractType,
		"feed_name":              feedName,
		"feed_path":              feedPath,
		"network_id":             networkID,
		"network_name":           networkName,
	}).Set(answer)
}

func (d *defaultMetrics) SetProxyAnswers(answer float64, proxyContractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
	proxyAnswers.With(prometheus.Labels{
		"proxy_contract_address": proxyContractAddress,
		"feed_id":                feedID,
		"chain_id":               chainID,
		"contract_status":        contractStatus,
		"contract_type":          contractType,
		"feed_name":              feedName,
		"feed_path":              feedPath,
		"network_id":             networkID,
		"network_name":           networkName,
	}).Set(answer)
}

func (d *defaultMetrics) Cleanup(
	proxyContractAddress, feedID, chainID, contractStatus, contractType string,
	feedName, feedPath, networkID, networkName string,
) {
	labels := prometheus.Labels{
		"proxy_contract_address": proxyContractAddress,
		"feed_id":                feedID,
		"chain_id":               chainID,
		"contract_status":        contractStatus,
		"contract_type":          contractType,
		"feed_name":              feedName,
		"feed_path":              feedPath,
		"network_id":             networkID,
		"network_name":           networkName,
	}
	if !proxyAnswersRaw.Delete(labels) {
		d.log.Errorw("failed to delete metric", "name", "proxy_answers_raw", "labels", labels)
	}
	if !proxyAnswers.Delete(labels) {
		d.log.Errorw("failed to delete metric", "name", "proxy_answers", "labels", labels)
	}
}
