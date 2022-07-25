package ocr2

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	junotypes "github.com/NethermindEth/juno/pkg/types"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/stretchr/testify/assert"
)

func TestBuildReport(t *testing.T) {
	c := ReportCodec{}
	oo := []median.ParsedAttributedObservation{}

	// expected outputs
	n := 4
	observers := make([]byte, 32)
	v := big.NewInt(0)
	v.SetString("1000000000000000000", 10)

	for i := 0; i < n; i++ {
		oo = append(oo, median.ParsedAttributedObservation{
			Timestamp:       uint32(time.Now().Unix()),
			Value:           big.NewInt(1234567890),
			JuelsPerFeeCoin: v,
			Observer:        commontypes.OracleID(i),
		})

		// create expected outputs
		observers[i] = uint8(i)
	}

	report, err := c.BuildReport(oo)
	assert.NoError(t, err)

	// validate length
	totalLen := prefixSizeBytes + observationSizeBytes*n + juelsPerFeeCoinSizeBytes
	assert.Equal(t, totalLen, len(report), "validate length")

	// validate timestamp
	timestamp := junotypes.BytesToFelt(report[0:timestampSizeBytes]).Big()
	assert.Equal(t, uint64(oo[0].Timestamp), timestamp.Uint64(), "validate timestamp")

	// validate observers
	index := timestampSizeBytes
	assert.Equal(t, observers, []byte(report[index:index+observersSizeBytes]), "validate observers")

	// validate observer count
	index += observersSizeBytes
	count := junotypes.BytesToFelt(report[index : index+observationsLenBytes]).Big()
	assert.Equal(t, uint8(n), uint8(count.Uint64()), "validate observer count")

	// validate observations
	for i := 0; i < n; i++ {
		index := prefixSizeBytes + observationSizeBytes*i
		assert.Equal(t, oo[0].Value.FillBytes(make([]byte, observationSizeBytes)), []byte(report[index:index+observationSizeBytes]), fmt.Sprintf("validate median observation #%d", i))
	}

	// validate juelsToEth
	assert.Equal(t, v.FillBytes(make([]byte, juelsPerFeeCoinSizeBytes)), []byte(report[totalLen-juelsPerFeeCoinSizeBytes:totalLen]), "validate juelsToEth")
}