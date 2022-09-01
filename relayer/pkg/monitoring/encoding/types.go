package encoding

import (
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

const (
	LatestRoundDataViewName           = "latest_round_dat"
	LatestConfigDetailsViewName       = "latest_config_details"
	LatestTransmissionDetailsViewName = "latest_transmission_details"
	LinkAvailableForPaymentViewName   = "link_available_for_payment"
	BalanceOfMethod                   = "balanceOf"

	NewTransmissionEventName = "new_transmission"
	ConfigSetEventName       = "config_set"
)

type LatestConfigDetails struct {
	ConfigCount  uint64
	BlockNumber  uint64
	ConfigDigest types.ConfigDigest
}

type LatestTransmissionDetails struct {
	ConfigDigest    types.ConfigDigest
	Epoch           uint32
	Round           uint8
	LatestAnswer    *big.Int
	LatestTimestamp time.Time
}

type RoundData struct {
	RoundID     uint32
	Answer      *big.Int
	BlockNumber uint64
	StartedAt   time.Time
	UpdatedAt   time.Time
}

type LinkAvailableForPayment struct {
	Available *big.Int
}

type NewTransmisisonEvent struct {
	RoundID              uint32
	Answer               *big.Int
	Transmitter          types.Account
	ObservationTimestamp time.Time
	Observers            *big.Int
	Observations         []*big.Int
	JuelsPerFeeCoin      *big.Int
	ConfigDigest         types.ConfigDigest
	Epoch                uint32
	Round                uint8
	Reimbursement        *big.Int
}

type ConfigSetEvent struct {
	PreviousConfigBlockNumber uint64
	LatestConfigDigest        types.ConfigDigest
	ConfigCount               uint64
	Signers                   []types.OnchainPublicKey
	Transmitters              []types.Account
	F                         uint8
	OnchainConfig             []byte
	OffchainConfigVersion     uint64
	OffchainConfig            []byte
}
