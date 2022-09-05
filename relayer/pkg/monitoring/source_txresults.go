package monitoring

import (
	"context"
	"fmt"
	"sync"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

func NewTxResultsSourceFactory(
	ocr2Reader ocr2.OCR2Reader,
) relayMonitoring.SourceFactory {
	return &txResultsSourceFactory{
		ocr2Reader,
	}
}

type txResultsSourceFactory struct {
	ocr2Reader ocr2.OCR2Reader
}

func (s *txResultsSourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	return &txResultsSource{
		feedConfig.GetContractAddress(),
		s.ocr2Reader,
		0,
		sync.Mutex{},
	}, nil
}

func (s *txResultsSourceFactory) GetType() string {
	return "txresults"
}

type txResultsSource struct {
	contractAddress string
	ocr2Reader      ocr2.OCR2Reader

	prevRoundID   uint32
	prevRoundIDMu sync.Mutex
}

func (s *txResultsSource) Fetch(ctx context.Context) (interface{}, error) {
	latestRoundData, err := s.ocr2Reader.LatestRoundData(ctx, s.contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest_round_data: %w", err)
	}
	s.prevRoundIDMu.Lock()
	defer s.prevRoundIDMu.Unlock()
	var numSucceeded uint32
	if s.prevRoundID != 0 {
		numSucceeded = latestRoundData.RoundID - s.prevRoundID
		s.prevRoundID = latestRoundData.RoundID
	}
	return relayMonitoring.TxResults{NumSucceeded: uint64(numSucceeded), NumFailed: 0}, nil
}
