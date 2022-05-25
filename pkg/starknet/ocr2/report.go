package ocr2

import (
	"math/big"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.ReportCodec = (*reportCodec)(nil)

type reportCodec struct{}

func (c reportCodec) BuildReport(oo []median.ParsedAttributedObservation) (types.Report, error) {
	// todo: implement
	return types.Report{}, nil
}

func (c reportCodec) MedianFromReport(report types.Report) (*big.Int, error) {
	// todo: implement
	return nil, nil
}

func (c reportCodec) MaxReportLength(n int) int {
	// todo: implement
	return 0
}
