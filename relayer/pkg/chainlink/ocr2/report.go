package ocr2

import (
	"math/big"
	"sort"

	"github.com/pkg/errors"

	junotypes "github.com/NethermindEth/juno/pkg/types"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.ReportCodec = (*reportCodec)(nil)

const (
	timestampSizeBytes       = junotypes.FeltLength
	observersSizeBytes       = junotypes.FeltLength
	observationsLenBytes     = junotypes.FeltLength
	prefixSizeBytes          = timestampSizeBytes + observersSizeBytes + observationsLenBytes
	juelsPerFeeCoinSizeBytes = junotypes.FeltLength
	observationSizeBytes     = junotypes.FeltLength
)

type reportCodec struct{}

func (c reportCodec) BuildReport(oo []median.ParsedAttributedObservation) (types.Report, error) {
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

func (c reportCodec) MedianFromReport(report types.Report) (*big.Int, error) {
	rLen := len(report)
	if rLen < prefixSizeBytes+juelsPerFeeCoinSizeBytes {
		return nil, errors.New("invalid report length")
	}

	numFelt := junotypes.BytesToFelt(report[(timestampSizeBytes + observersSizeBytes):prefixSizeBytes])
	num := int(numFelt.Big().Int64())
	if num == 0 {
		return nil, errors.New("unpacked report has no observations")
	}

	var observations []*big.Int
	for i := 0; i < num; i++ {
		idx := prefixSizeBytes + observationSizeBytes*i
		o := junotypes.BytesToFelt(report[idx:(idx + observationSizeBytes)]).Big()
		observations = append(observations, o)
	}

	return observations[num/2], nil
}

func (c reportCodec) MaxReportLength(n int) int {
	return prefixSizeBytes + (n * observationSizeBytes) + juelsPerFeeCoinSizeBytes
}
