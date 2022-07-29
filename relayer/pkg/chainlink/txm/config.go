package txm

import "time"

//go:generate mockery --name TxConfig --output ./mocks/ --case=underscore --filename txconfig.go

// txm config
type TxConfig interface {
	TxTimeout() time.Duration
	TxSendFrequency() time.Duration
	TxMaxBatchSize() int
}
