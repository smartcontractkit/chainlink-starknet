package monitoring

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-starknet/monitoring/pkg/monitoring/mocks"

	commonMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

func TestTransmissionDetailsExporter(t *testing.T) {
	chainConfig := generateChainConfig()
	feedConfig := generateFeedConfig()

	mockMetrics := mocks.NewMetrics(t)
	factory := NewTransmissionDetailsExporterFactory(mockMetrics)

	gasPrice, ok := new(big.Int).SetString("10000000000000000000", 10)
	require.True(t, ok)

	envelope := TransmissionsEnvelope{
		Transmissions: []TransmissionInfo{{
			GasPrice:          gasPrice, // 10 STRK (10^19 FRI)
			ObservationLength: 123,
		},
		},
	}

	mockMetrics.On(
		"SetTransmissionGasPrice",
		float64(10),
		feedConfig.ContractAddress,
		feedConfig.GetID(),
		chainConfig.GetChainID(),
		feedConfig.GetContractStatus(),
		feedConfig.GetContractType(),
		feedConfig.Name,
		feedConfig.Path,
		chainConfig.GetNetworkID(),
		chainConfig.GetNetworkName(),
	).Once()

	mockMetrics.On(
		"SetReportObservations",
		float64(123),
		feedConfig.ContractAddress,
		feedConfig.GetID(),
		chainConfig.GetChainID(),
		feedConfig.GetContractStatus(),
		feedConfig.GetContractType(),
		feedConfig.Name,
		feedConfig.Path,
		chainConfig.GetNetworkID(),
		chainConfig.GetNetworkName(),
	).Once()

	exporter, err := factory.NewExporter(commonMonitoring.ExporterParams{
		ChainConfig: chainConfig,
		FeedConfig:  feedConfig,
	})
	require.NoError(t, err)

	exporter.Export(context.Background(), envelope)

	// cleanup
	mockMetrics.On(
		"CleanupReportObservations",
		feedConfig.ContractAddress,
		feedConfig.GetID(),
		chainConfig.GetChainID(),
		feedConfig.GetContractStatus(),
		feedConfig.GetContractType(),
		feedConfig.Name,
		feedConfig.Path,
		chainConfig.GetNetworkID(),
		chainConfig.GetNetworkName(),
	).Once()
	mockMetrics.On(
		"CleanupTransmissionGasPrice",
		feedConfig.ContractAddress,
		feedConfig.GetID(),
		chainConfig.GetChainID(),
		feedConfig.GetContractStatus(),
		feedConfig.GetContractType(),
		feedConfig.Name,
		feedConfig.Path,
		chainConfig.GetNetworkID(),
		chainConfig.GetNetworkName(),
	).Once()

	exporter.Cleanup(context.Background())

}
