package ocr2

import (
	"math/big"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.ReportCodec = (*ReportCodec)(nil)

type ReportCodec struct{}

func (c ReportCodec) BuildReport(oo []median.ParsedAttributedObservation) (types.Report, error) {
	// todo: implement
	return types.Report{}, nil
}

func (c ReportCodec) MedianFromReport(report types.Report) (*big.Int, error) {
	// todo: implement
	return nil, nil
}

func (c ReportCodec) MaxReportLength(n int) int {
	// todo: implement
	return 0
}
