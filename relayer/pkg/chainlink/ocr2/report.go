package ocr2

import (
	"math"
	"math/big"
	"sort"

	"github.com/pkg/errors"

	junotypes "github.com/NethermindEth/juno/pkg/types"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.ReportCodec = (*ReportCodec)(nil)

const (
	timestampSizeBytes       = junotypes.FeltLength
	observersSizeBytes       = junotypes.FeltLength
	observationsLenBytes     = junotypes.FeltLength
	prefixSizeBytes          = timestampSizeBytes + observersSizeBytes + observationsLenBytes
	juelsPerFeeCoinSizeBytes = junotypes.FeltLength
	observationSizeBytes     = junotypes.FeltLength
)

type ReportCodec struct{}

func (c ReportCodec) BuildReport(oo []median.ParsedAttributedObservation) (types.Report, error) {
	num := len(oo)
	if num == 0 {
		return nil, errors.New("couldn't build report from empty attributed observations")
	}

	// preserve original array
	oo = append([]median.ParsedAttributedObservation{}, oo...)
	numFelt := junotypes.BigToFelt(big.NewInt(int64(num)))

	// median timestamp
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].Timestamp < oo[j].Timestamp
	})
	timestamp := oo[num/2].Timestamp
	timestampFelt := junotypes.BigToFelt(big.NewInt(int64(timestamp)))

	// median juelsPerFeeCoin
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].JuelsPerFeeCoin.Cmp(oo[j].JuelsPerFeeCoin) < 0
	})
	juelsPerFeeCoin := oo[num/2].JuelsPerFeeCoin
	juelsPerFeeCoinFelt := junotypes.BigToFelt(juelsPerFeeCoin)

	// sort by values
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].Value.Cmp(oo[j].Value) < 0
	})

	var observers junotypes.Felt
	var observations []junotypes.Felt
	for i, o := range oo {
		observers[i] = byte(o.Observer)
		observations = append(observations, junotypes.BigToFelt(o.Value))
	}

	var report []byte
	report = append(report, timestampFelt.Bytes()...)
	report = append(report, observers.Bytes()...)
	report = append(report, numFelt.Bytes()...)
	for _, o := range observations {
		report = append(report, o.Bytes()...)
	}
	report = append(report, juelsPerFeeCoinFelt.Bytes()...)

	return report, nil
}

func (c ReportCodec) MedianFromReport(report types.Report) (*big.Int, error) {
	rLen := len(report)
	if rLen < prefixSizeBytes+juelsPerFeeCoinSizeBytes {
		return nil, errors.New("invalid report length")
	}

	numBig := junotypes.BytesToFelt(report[(timestampSizeBytes + observersSizeBytes):prefixSizeBytes]).Big()
	if !numBig.IsUint64() {
		return nil, errors.New("length of observations is invalid")
	}
	n64 := numBig.Uint64()
	if n64 == 0 {
		return nil, errors.New("unpacked report has no observations")
	}
	if n64 >= math.MaxInt8 {
		return nil, errors.New("length of observations is invalid")
	}

	n := int(n64)

	if rLen < prefixSizeBytes+(observationSizeBytes*n)+juelsPerFeeCoinSizeBytes {
		return nil, errors.New("report does not contain enough observations or is missing juels/feeCoin observation")
	}

	var observations []*big.Int
	for i := 0; i < n; i++ {
		start := prefixSizeBytes + observationSizeBytes*i
		end := start + observationSizeBytes
		o := junotypes.BytesToFelt(report[start:end]).Big()
		observations = append(observations, o)
	}

	return observations[n/2], nil
}

func (c ReportCodec) MaxReportLength(n int) int {
	return prefixSizeBytes + (n * observationSizeBytes) + juelsPerFeeCoinSizeBytes
}
