package medianreport

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/NethermindEth/starknet.go/curve"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestBuildReportWithNegativeValues(t *testing.T) {
	ctx := tests.Context(t)
	c := ReportCodec{}
	oo := []median.ParsedAttributedObservation{}

	oo = append(oo, median.ParsedAttributedObservation{
		Timestamp:        uint32(time.Now().Unix()),
		Value:            big.NewInt(-10),
		JuelsPerFeeCoin:  big.NewInt(10),
		GasPriceSubunits: big.NewInt(10),
		Observer:         commontypes.OracleID(1),
	})

	_, err := c.BuildReport(ctx, oo)
	assert.ErrorContains(t, err, "starknet does not support negative values: value = (-10), fee = (10), gas = (10)")

	oo = []median.ParsedAttributedObservation{}
	oo = append(oo, median.ParsedAttributedObservation{
		Timestamp:        uint32(time.Now().Unix()),
		Value:            big.NewInt(10),
		JuelsPerFeeCoin:  big.NewInt(-10),
		GasPriceSubunits: big.NewInt(10),
		Observer:         commontypes.OracleID(1),
	})

	_, err = c.BuildReport(ctx, oo)
	assert.ErrorContains(t, err, "starknet does not support negative values: value = (10), fee = (-10), gas = (10)")

	oo = []median.ParsedAttributedObservation{}
	oo = append(oo, median.ParsedAttributedObservation{
		Timestamp:        uint32(time.Now().Unix()),
		Value:            big.NewInt(10),
		JuelsPerFeeCoin:  big.NewInt(10),
		GasPriceSubunits: big.NewInt(-10),
		Observer:         commontypes.OracleID(1),
	})

	_, err = c.BuildReport(ctx, oo)
	assert.ErrorContains(t, err, "starknet does not support negative values: value = (10), fee = (10), gas = (-10)")
}

func TestBuildReportNoObserversOverflow(t *testing.T) {
	ctx := tests.Context(t)
	c := ReportCodec{}
	oo := []median.ParsedAttributedObservation{}
	fmt.Println("hello")
	v := big.NewInt(0)
	v.SetString("1000000000000000000", 10)

	// test largest possible encoded observers byte array
	for i := 30; i >= 0; i-- {
		oo = append(oo, median.ParsedAttributedObservation{
			Timestamp:        uint32(time.Now().Unix()),
			Value:            big.NewInt(1234567890),
			GasPriceSubunits: big.NewInt(2),
			JuelsPerFeeCoin:  v,
			Observer:         commontypes.OracleID(i),
		})
	}

	report, err := c.BuildReport(ctx, oo)
	assert.Nil(t, err)

	index := timestampSizeBytes
	observersBytes := []byte(report[index : index+observersSizeBytes])
	observersBig := starknetutils.BytesToBig(observersBytes)

	// encoded observers felt is less than max felt
	assert.Equal(t, -1, observersBig.Cmp(curve.Curve.P), "observers should be less than max felt")
}

func TestBuildReport(t *testing.T) {
	ctx := tests.Context(t)
	c := ReportCodec{}
	oo := []median.ParsedAttributedObservation{}

	// expected outputs
	n := 4
	observers := make([]byte, 32)
	v := big.NewInt(0)
	v.SetString("1000000000000000000", 10)

	// 0x01 pad the first byte
	observers[0] = uint8(1)
	for i := 0; i < n; i++ {
		oo = append(oo, median.ParsedAttributedObservation{
			Timestamp:        uint32(time.Now().Unix()),
			Value:            big.NewInt(1234567890),
			GasPriceSubunits: big.NewInt(2),
			JuelsPerFeeCoin:  v,
			Observer:         commontypes.OracleID(i),
		})

		// create expected outputs
		// remember to add 1 byte offset to avoid felt overflow
		observers[i+1] = uint8(i)
	}

	report, err := c.BuildReport(ctx, oo)
	assert.NoError(t, err)

	// validate length
	totalLen := prefixSizeBytes + observationSizeBytes*n + juelsPerFeeCoinSizeBytes + gasPriceSizeBytes
	assert.Equal(t, totalLen, len(report), "validate length")

	// validate timestamp
	timestamp := new(big.Int).SetBytes(report[0:timestampSizeBytes])
	assert.Equal(t, uint64(oo[0].Timestamp), timestamp.Uint64(), "validate timestamp")

	// validate observers
	index := timestampSizeBytes
	assert.Equal(t, observers, []byte(report[index:index+observersSizeBytes]), "validate observers")

	// validate observer count
	index += observersSizeBytes
	count := new(big.Int).SetBytes(report[index : index+observationsLenBytes])
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
	expectedGasPrice := big.NewInt(2)
	assert.Equal(t, expectedGasPrice.FillBytes(make([]byte, gasPriceSizeBytes)), []byte(report[index:index+gasPriceSizeBytes]), "validate gasPrice")
}

type medianTest struct {
	name           string
	obs            []*big.Int
	expectedMedian *big.Int
}

func TestMedianFromReport(t *testing.T) {
	ctx := tests.Context(t)
	cdc := ReportCodec{}
	// Requires at least one obs
	_, err := cdc.BuildReport(ctx, nil)
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
			ctx := tests.Context(t)
			var pos []median.ParsedAttributedObservation
			for i, obs := range tc.obs {
				pos = append(pos, median.ParsedAttributedObservation{
					Value:            obs,
					JuelsPerFeeCoin:  obs,
					GasPriceSubunits: obs,
					Observer:         commontypes.OracleID(uint8(i))},
				)
			}
			report, err := cdc.BuildReport(ctx, pos)
			require.NoError(t, err)
			max, err := cdc.MaxReportLength(ctx, len(tc.obs))
			require.NoError(t, err)
			assert.Equal(t, len(report), max)
			med, err := cdc.MedianFromReport(ctx, report)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMedian.String(), med.String())
		})
	}
}
