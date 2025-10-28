package config

import (
	_ "embed"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/config/configtest"
)

func ptr[T any](t T) *T {
	return &t
}

func TestDefaults_fieldsNotNil(t *testing.T) {
	configtest.AssertFieldsNotNil(t, Defaults())
}

func TestDocsTOMLComplete(t *testing.T) {
	configtest.AssertDocsTOMLComplete[TOMLConfig](t, docsTOML)
}

//go:embed testdata/config-full.toml
var fullTOML string

func TestTOMLConfig_FullMarshal(t *testing.T) {
	full := TOMLConfig{
		ChainID:   ptr("Ibiza-808"),
		Enabled:   ptr(true),
		FeederURL: config.MustParseURL("http://feeder.url"),
		Chain: Chain{
			OCR2CachePollPeriod:      config.MustNewDuration(6 * time.Hour),
			OCR2CacheTTL:             config.MustNewDuration(3 * time.Minute),
			RequestTimeout:           config.MustNewDuration(1*time.Minute + 3*time.Second),
			TxTimeout:                config.MustNewDuration(13 * time.Second),
			ConfirmationPoll:         config.MustNewDuration(42 * time.Second),
			MaxAttempts:              ptr(15),
			FeeEstimationMaxAttempts: ptr(8),
		},
		Nodes: []*Node{
			{
				Name:   ptr("primary"),
				URL:    config.MustParseURL("http://stark.node"),
				APIKey: ptr("key"),
			},
		},
	}
	configtest.AssertFullMarshal(t, full, fullTOML)
}
