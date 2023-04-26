package medianreport

import (
	"github.com/smartcontractkit/libocr/offchainreporting2plus/chains/evmutil"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// NOTE: this should sit in the ocr2 package but that causes import cycles
func RawReportContext(repctx types.ReportContext) [3][32]byte {
	rawReportContext := evmutil.RawReportContext(repctx)
	// NOTE: Ensure extra_hash is 31 bytes with first byte blanked out
	// libocr generates a 32 byte extraHash but we need to fit it into a felt
	rawReportContext[2][0] = 0
	return rawReportContext
}
