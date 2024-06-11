package monitoring

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	commonMonitor "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-common/pkg/monitoring/config"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// Builds monitor instance with only the prometheus exporter.
// Does not contain kafka exporter.
func NewMonitorPrometheusOnly(
	stopCh services.StopRChan,
	log commonMonitor.Logger,
	chainConfig commonMonitor.ChainConfig,
	envelopeSourceFactory commonMonitor.SourceFactory,
	txResultsSourceFactory commonMonitor.SourceFactory,
	feedsParser commonMonitor.FeedsParser,
	nodesParser commonMonitor.NodesParser,
) (*commonMonitor.Monitor, error) {
	cfg, err := ParseWithoutKafka()
	if err != nil {
		return nil, fmt.Errorf("failed to parse generic configuration: %w", err)
	}

	metrics := commonMonitor.NewMetrics(logger.With(log, "component", "metrics"))
	chainMetrics := commonMonitor.NewChainMetrics(chainConfig)

	sourceFactories := []commonMonitor.SourceFactory{envelopeSourceFactory, txResultsSourceFactory}

	prometheusExporterFactory := commonMonitor.NewPrometheusExporterFactory(
		logger.With(log, "component", "prometheus-exporter"),
		metrics,
	)

	exporterFactories := []commonMonitor.ExporterFactory{prometheusExporterFactory}

	rddSource := commonMonitor.NewRDDSource(
		cfg.Feeds.URL, feedsParser, cfg.Feeds.IgnoreIDs,
		cfg.Nodes.URL, nodesParser,
		logger.With(log, "component", "rdd-source"),
	)

	rddPoller := commonMonitor.NewSourcePoller(
		rddSource,
		logger.With(log, "component", "rdd-poller"),
		cfg.Feeds.RDDPollInterval,
		cfg.Feeds.RDDReadTimeout,
		0, // no buffering!
	)

	manager := commonMonitor.NewManager(
		logger.With(log, "component", "manager"),
		rddPoller,
	)

	// Configure HTTP server
	httpServer := commonMonitor.NewHTTPServer(stopCh, cfg.HTTP.Address, logger.With(log, "component", "http-server"))
	httpServer.Handle("/metrics", metrics.HTTPHandler())
	httpServer.Handle("/debug", manager.HTTPHandler())
	// Required for k8s.
	httpServer.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return &commonMonitor.Monitor{
		StopCh:      stopCh,
		ChainConfig: chainConfig,
		Config:      cfg,
		Log:         log,
		// no kafka
		Producer:     nil,
		Metrics:      metrics,
		ChainMetrics: chainMetrics,
		// no kafka
		SchemaRegistry:    nil,
		SourceFactories:   sourceFactories,
		ExporterFactories: exporterFactories,

		RDDSource: rddSource,
		RDDPoller: rddPoller,

		Manager: manager,

		HTTPServer: httpServer,
	}, nil
}

type MyConfig = config.Config

func ParseWithoutKafka() (MyConfig, error) {
	cfg := MyConfig{}

	if err := parseConfigEnvVars(&cfg); err != nil {
		return cfg, err
	}

	applyConfigDefaults(&cfg)

	err := validateConfigWithoutKafka(cfg)

	return cfg, err
}

func applyConfigDefaults(cfg *MyConfig) {
	if cfg.Feeds.RDDReadTimeout == 0 {
		cfg.Feeds.RDDReadTimeout = 1 * time.Second
	}
	if cfg.Feeds.RDDPollInterval == 0 {
		cfg.Feeds.RDDPollInterval = 10 * time.Second
	}
}

func parseConfigEnvVars(cfg *MyConfig) error {
	if value, isPresent := os.LookupEnv("FEEDS_URL"); isPresent {
		cfg.Feeds.URL = value
	}
	if value, isPresent := os.LookupEnv("FEEDS_RDD_READ_TIMEOUT"); isPresent {
		readTimeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var FEEDS_RDD_READ_TIMEOUT, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.Feeds.RDDReadTimeout = readTimeout
	}
	if value, isPresent := os.LookupEnv("FEEDS_RDD_POLL_INTERVAL"); isPresent {
		pollInterval, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("failed to parse env var FEEDS_RDD_POLL_INTERVAL, see https://pkg.go.dev/time#ParseDuration: %w", err)
		}
		cfg.Feeds.RDDPollInterval = pollInterval
	}
	if value, isPresent := os.LookupEnv("FEEDS_IGNORE_IDS"); isPresent {
		ids := strings.Split(value, ",")
		for _, id := range ids {
			if id == "" {
				continue
			}
			cfg.Feeds.IgnoreIDs = append(cfg.Feeds.IgnoreIDs, strings.TrimSpace(id))
		}
	}
	if value, isPresent := os.LookupEnv("NODES_URL"); isPresent {
		cfg.Nodes.URL = value
	}

	if value, isPresent := os.LookupEnv("HTTP_ADDRESS"); isPresent {
		cfg.HTTP.Address = value
	}

	return nil
}

func validateConfigWithoutKafka(cfg MyConfig) error {
	// Required config
	for envVarName, currentValue := range map[string]string{
		"FEEDS_URL": cfg.Feeds.URL,
		"NODES_URL": cfg.Nodes.URL,

		"HTTP_ADDRESS": cfg.HTTP.Address,
	} {
		if currentValue == "" {
			return fmt.Errorf("'%s' env var is required", envVarName)
		}
	}
	// Validate URLs.
	for envVarName, currentValue := range map[string]string{
		"FEEDS_URL": cfg.Feeds.URL,
		"NODES_URL": cfg.Nodes.URL,
	} {
		if _, err := url.ParseRequestURI(currentValue); err != nil {
			return fmt.Errorf("%s='%s' is not a valid URL: %w", envVarName, currentValue, err)
		}
	}

	return nil
}
