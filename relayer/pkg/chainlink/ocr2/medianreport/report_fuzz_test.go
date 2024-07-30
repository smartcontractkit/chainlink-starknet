//go:build go1.18
// +build go1.18

package medianreport

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

// go test -tags=go1.18 -fuzz ./...
func FuzzReportCodecMedianFromReport(f *testing.F) {
	ctx := tests.Context(f)
	cdc := ReportCodec{}
	report, err := cdc.BuildReport(ctx, []median.ParsedAttributedObservation{
		{Timestamp: uint32(time.Now().Unix()), Value: big.NewInt(10), JuelsPerFeeCoin: big.NewInt(100000), GasPriceSubunits: big.NewInt(100000)},
		{Timestamp: uint32(time.Now().Unix()), Value: big.NewInt(10), JuelsPerFeeCoin: big.NewInt(200000), GasPriceSubunits: big.NewInt(200000)},
		{Timestamp: uint32(time.Now().Unix()), Value: big.NewInt(11), JuelsPerFeeCoin: big.NewInt(300000), GasPriceSubunits: big.NewInt(300000)},
	})
	require.NoError(f, err)

	// Seed with valid report
	f.Add([]byte(report))
	f.Fuzz(func(t *testing.T, report []byte) {
		ctx := tests.Context(t)
		med, err := cdc.MedianFromReport(ctx, report)
		if err == nil {
			// Should always be able to build a report from the medians extracted
			_, err = cdc.BuildReport(ctx, []median.ParsedAttributedObservation{{Timestamp: uint32(time.Now().Unix()), Value: med, JuelsPerFeeCoin: med, GasPriceSubunits: med}})
			require.NoError(t, err)
		}
	})
}
