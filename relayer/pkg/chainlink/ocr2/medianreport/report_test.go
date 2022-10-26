package medianreport

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
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
	totalLen := prefixSizeBytes + observationSizeBytes*n + juelsPerFeeCoinSizeBytes + gasPriceSizeBytes
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
		idx := prefixSizeBytes + observationSizeBytes*i
		assert.Equal(t, oo[0].Value.FillBytes(make([]byte, observationSizeBytes)), []byte(report[idx:idx+observationSizeBytes]), fmt.Sprintf("validate median observation #%d", i))
	}

	// validate juelsPerFeeCoin
	index = prefixSizeBytes + observationSizeBytes*n
	assert.Equal(t, v.FillBytes(make([]byte, juelsPerFeeCoinSizeBytes)), []byte(report[index:index+juelsPerFeeCoinSizeBytes]), "validate juelsPerFeeCoin")

	// validate gasPrice
	index += juelsPerFeeCoinSizeBytes
	expectedGasPrice := big.NewInt(1)
	assert.Equal(t, expectedGasPrice.FillBytes(make([]byte, gasPriceSizeBytes)), []byte(report[index:index+gasPriceSizeBytes]), "validate gasPrice")
}

type medianTest struct {
	name           string
	obs            []*big.Int
	expectedMedian *big.Int
}

func TestMedianFromReport(t *testing.T) {
	cdc := ReportCodec{}
	// Requires at least one obs
	_, err := cdc.BuildReport(nil)
	require.Error(t, err)
	var tt = []medianTest{
		{
			name:           "2 positive one zero",
			obs:            []*big.Int{big.NewInt(0), big.NewInt(10), big.NewInt(20)},
			expectedMedian: big.NewInt(10),
		},
		{
			name:           "one zero",
			obs:            []*big.Int{big.NewInt(0)},
			expectedMedian: big.NewInt(0),
		},
		{
			name:           "two equal",
			obs:            []*big.Int{big.NewInt(1), big.NewInt(1)},
			expectedMedian: big.NewInt(1),
		},
		{
			name: "one negative one positive",
			obs:  []*big.Int{big.NewInt(-1), big.NewInt(1)},
			// sorts to -1, 1
			expectedMedian: big.NewInt(1),
		},
		{
			name: "two negative",
			obs:  []*big.Int{big.NewInt(-2), big.NewInt(-1)},
			// will sort to -2, -1
			expectedMedian: big.NewInt(-1),
		},
		{
			name: "three negative",
			obs:  []*big.Int{big.NewInt(-5), big.NewInt(-3), big.NewInt(-1)},
			// will sort to -5, -3, -1
			expectedMedian: big.NewInt(-3),
		},
	}

	// add cases for observation number from [1..31]
	for i := 1; i < 32; i++ {
		test := medianTest{
			name:           fmt.Sprintf("observations=%d", i),
			obs:            []*big.Int{},
			expectedMedian: big.NewInt(1),
		}
		for j := 0; j < i; j++ {
			test.obs = append(test.obs, big.NewInt(1))
		}
		tt = append(tt, test)
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var pos []median.ParsedAttributedObservation
			for i, obs := range tc.obs {
				pos = append(pos, median.ParsedAttributedObservation{
					Value:           obs,
					JuelsPerFeeCoin: obs,
					Observer:        commontypes.OracleID(uint8(i))},
				)
			}
			report, err := cdc.BuildReport(pos)
			require.NoError(t, err)
			assert.Equal(t, len(report), cdc.MaxReportLength(len(tc.obs)))
			med, err := cdc.MedianFromReport(report)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMedian.String(), med.String())
		})
	}

}
