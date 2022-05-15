package contract

const (
	// todo: set actual values
	MaxOracles           = 0
	MaxOffchainConfigLen = 0
)

// Contract State
type State struct {
	// todo: adjust according to the actual state structure
	Config         OnchainConfig
	OffchainConfig OffchainConfig
	Oracles        Oracles
}

type OnchainConfig struct {
	// todo: adjust data types
	Owner                     string
	ProposedOwner             string
	TokenMint                 string
	TokenVault                string
	RequesterAccessController string
	BillingAccessController   string
	MinAnswer                 int
	MaxAnswer                 int
	F                         uint8
	Round                     uint8
	Padding0                  uint16
	Epoch                     uint32
	LatestAggregatorRoundID   uint32
	LatestTransmitter         string
	ConfigCount               uint32
	LatestConfigDigest        [32]byte
	LatestConfigBlockNumber   uint64
	Billing                   Billing
}

type OffchainConfig struct {
	Version uint64
	Raw     [MaxOffchainConfigLen]byte
	Len     uint64
}

type Oracles struct {
	Raw [MaxOracles]Oracle
	Len uint64
}

type Oracle struct {
	// todo: adjust data types
	Transmitter   string
	Signer        string
	Payee         string
	ProposedPayee string
	FromRoundID   uint32
	Payment       uint64
}

type Billing struct {
	ObservationPayment  uint32
	TransmissionPayment uint32
}

type OCR2Spec struct {
	// todo: add spec
	ID      int32
	ChainID string
}
