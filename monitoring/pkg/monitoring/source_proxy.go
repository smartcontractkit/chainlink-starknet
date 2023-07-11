package monitoring

import (
	"context"
	"fmt"
	"math/big"

<<<<<<< HEAD
	caigotypes "github.com/smartcontractkit/caigo/types"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
||||||| parent of 5c692ac2 (Use latest upstream sdk: starknet.go)
	caigotypes "github.com/smartcontractkit/caigo/types"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
=======
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/NethermindEth/juno/core/felt"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
>>>>>>> 5c692ac2 (Use latest upstream sdk: starknet.go)

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

type ProxyData struct {
	Answer *big.Int
}

func NewProxySourceFactory(
	ocr2Reader ocr2.OCR2Reader,
) relayMonitoring.SourceFactory {
	return &proxySourceFactory{
		ocr2Reader,
	}
}

type proxySourceFactory struct {
	ocr2Reader ocr2.OCR2Reader
}

func (s *proxySourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	contractAddress, err := starknetutils.HexToFelt(feedConfig.GetContractAddress())
	if err != nil {
		return nil, err
	}
	return &proxySource{
		contractAddress,
		s.ocr2Reader,
	}, nil
}

func (s *proxySourceFactory) GetType() string {
	return "proxy"
}

type proxySource struct {
	contractAddress *felt.Felt
	ocr2Reader      ocr2.OCR2Reader
}

func (s *proxySource) Fetch(ctx context.Context) (interface{}, error) {
	latestTransmission, err := s.ocr2Reader.LatestTransmissionDetails(ctx, s.contractAddress)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch latest_transmission_details: %w", err)
	}
	return ProxyData{
		Answer: latestTransmission.LatestAnswer,
	}, nil
}
