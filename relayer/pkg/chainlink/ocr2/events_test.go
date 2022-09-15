package ocr2

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var (
	newTransmissionEventRaw = []string{
		"0x1",
		"0x63",
		"0x2c0dd77ce74b1667dc6fa782bbafaef5becbe2d04b052726ab236daeb52ac5d",
		"0x1",
		"0x10203000000000000000000000000000000000000000000000000000000",
		"0x4",
		"0x63",
		"0x63",
		"0x63",
		"0x63",
		"0x1",
		"0x1",
		"0x485341c18461d70eac6ded4b8b17147f173308ddd56216a86f9ec4d994453",
		"0x1",
		"0x0",
	}
	configSetEventRaw = []string{
		"0x0",
		"0x485341c18461d70eac6ded4b8b17147f173308ddd56216a86f9ec4d994453",
		"0x1",
		"0x4",
		"0x21e867aa6e6c545949a9c6f9f5401b70007bd93675857a0a7d5345b8bffcbf0",
		"0x2c0dd77ce74b1667dc6fa782bbafaef5becbe2d04b052726ab236daeb52ac5d",
		"0x64642f34e68436f45757b920f4cdfbdff82728844d740bac672a19ad72011ca",
		"0x2de61335d8f1caa7e9df54486f016ded83d0e02fde4c12280f4b898720b0e2b",
		"0x3fad2efda193b37e4e526972d9613238b9ff993e1e3d3b1dd376d7b8ceb7acd",
		"0x2f14e18cc198dd5133c8a9aa92992fc1a462f703401716f402d0ee383b54faa",
		"0x4fcf11b05ebd00a207030c04836defbec3d37a3f77e581f2d0962a20a55adcd",
		"0x5c35686f78db31d9d896bb425b3fd99be19019f8aeaf0f7a8767867903341d4",
		"0x1",
		"0x3",
		"0x1",
		"0x800000000000010fffffffffffffffffffffffffffffffffffffffffffffff7",
		"0x3b9aca00",
		"0x2",
		"0x2",
		"0x1",
		"0x1",
	}
)

func TestNewTransmissionEvent_Parse(t *testing.T) {
	eventData := starknet.StringsToFelt(newTransmissionEventRaw)
	require.Equal(t, len(newTransmissionEventRaw), len(eventData))

	e, err := ParseNewTransmissionEvent(eventData)
	assert.NoError(t, err)

	require.Equal(t, e.RoundId, uint32(1))
	require.Equal(t, e.LatestAnswer, big.NewInt(99))
	require.Equal(t, e.LatestTimestamp, time.Unix(1, 0))
	require.Equal(t, e.Epoch, uint32(0))
	require.Equal(t, e.Round, uint8(1))
	require.Equal(t, e.Reimbursement, big.NewInt(0))

	require.Equal(t, e.JuelsPerFeeCoin, big.NewInt(1))
	require.Equal(t, e.GasPrice, big.NewInt(1))

	transmitterHex := "0x2c0dd77ce74b1667dc6fa782bbafaef5becbe2d04b052726ab236daeb52ac5d"
	require.Equal(t, len(transmitterHex), int(2+31.5*2)) // len('0x') + len(max_felt_len)
	require.Equal(t, e.Transmitter, caigotypes.StrToFelt(transmitterHex))

	require.Equal(t, e.Observers, []uint8{0, 1, 2, 3})
	require.Equal(t, e.ObservationsLen, uint32(4))
	require.Equal(t, e.ObservationsLen, uint32(len(e.Observers)))

	configDigest := XXXMustBytesToConfigDigest(starknet.XXXMustHexDecodeString("000485341c18461d70eac6ded4b8b17147f173308ddd56216a86f9ec4d994453"))
	require.Equal(t, len(configDigest), 32) // padded to 32 bytes
	require.Equal(t, e.ConfigDigest, configDigest)
}

func TestConfigSetEvent_Parse(t *testing.T) {
	eventData := starknet.StringsToFelt(configSetEventRaw)
	require.Equal(t, len(configSetEventRaw), len(eventData))

	e, err := ParseConfigSetEvent(eventData)
	assert.NoError(t, err)

	configDigest := XXXMustBytesToConfigDigest(starknet.XXXMustHexDecodeString("000485341c18461d70eac6ded4b8b17147f173308ddd56216a86f9ec4d994453"))
	require.Equal(t, len(configDigest), 32) // padded to 32 bytes
	require.Equal(t, e.ConfigDigest, configDigest)
	require.Equal(t, e.ConfigCount, uint64(1))

	oraclesLen := 4
	require.Equal(t, len(e.Signers), oraclesLen)
	signersExpected := []types.OnchainPublicKey{
		starknet.XXXMustHexDecodeString("021e867aa6e6c545949a9c6f9f5401b70007bd93675857a0a7d5345b8bffcbf0"),
		starknet.XXXMustHexDecodeString("064642f34e68436f45757b920f4cdfbdff82728844d740bac672a19ad72011ca"),
		starknet.XXXMustHexDecodeString("03fad2efda193b37e4e526972d9613238b9ff993e1e3d3b1dd376d7b8ceb7acd"),
		starknet.XXXMustHexDecodeString("04fcf11b05ebd00a207030c04836defbec3d37a3f77e581f2d0962a20a55adcd"),
	}
	require.Equal(t, e.Signers, signersExpected)

	transmittersExpected := []types.Account{
		"0x02c0dd77ce74b1667dc6fa782bbafaef5becbe2d04b052726ab236daeb52ac5d",
		"0x02de61335d8f1caa7e9df54486f016ded83d0e02fde4c12280f4b898720b0e2b",
		"0x02f14e18cc198dd5133c8a9aa92992fc1a462f703401716f402d0ee383b54faa",
		"0x05c35686f78db31d9d896bb425b3fd99be19019f8aeaf0f7a8767867903341d4",
	}
	require.Equal(t, e.Transmitters, transmittersExpected)
	require.Equal(t, len(e.Transmitters), oraclesLen)

	require.Equal(t, e.F, uint8(1))

	onchainConfig, err := medianreport.OnchainConfigCodec{}.EncodeFromBigInt(
		big.NewInt(medianreport.OnchainConfigVersion), // version
		big.NewInt(-10),        // min
		big.NewInt(1000000000), // max
	)
	assert.NoError(t, err)
	require.Equal(t, e.OnchainConfig, onchainConfig)

	require.Equal(t, e.OffchainConfigVersion, uint64(2))
	require.Equal(t, e.OffchainConfig, []uint8{0x1}) // dummy config
}

func TestNewTransmissionEventSelector(t *testing.T) {
	bytes, err := hex.DecodeString(NewTransmissionEventSelector)
	require.NoError(t, err)
	eventKey := new(big.Int)
	eventKey.SetBytes(bytes)
	assert.Equal(t, caigo.GetSelectorFromName("NewTransmission").Cmp(eventKey), 0)
}

func TestConfigSetEventSelector(t *testing.T) {
	bytes, err := hex.DecodeString(ConfigSetEventSelector)
	require.NoError(t, err)
	eventKey := new(big.Int)
	eventKey.SetBytes(bytes)
	assert.Equal(t, caigo.GetSelectorFromName("ConfigSet").Cmp(eventKey), 0)
}
