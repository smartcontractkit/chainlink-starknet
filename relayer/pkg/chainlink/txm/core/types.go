package core

import (
	"context"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

//go:generate mockery --name Config --output ./mocks/ --case=underscore --filename config.go

// ----------- Config -----------------
type Config interface {
	TxTimeout() time.Duration
	TxConfirmFrequency() time.Duration
	TxRetryFrequency() time.Duration
}

// ----------- Generic -----------------
// T - base transcation type to be sent in the queue
// K - base key type
// N - nonce type
// E - estimate/simulate type

type TxQueue[T any] interface {
	Enqueue(T) error
}

type TxManager[T any] interface {
	types.Service
	TxQueue[T]
	TxCount(Status) int
}

type Keystore[K any] interface {
	Get(id string) (K, error)
}

type TxStatuses[T any] interface {
	Get(status Status) []Tx[T]
	Exists(id string) bool
	Queued(tx T) (Tx[T], error)
	Retry(id string) (Tx[T], error)
	Broadcast(id, hash string) error
	Confirmed(id string) error
	Errored(id, err string) error
	Fatal(id string) error
}

type Tx[T any] interface {
	Sender() string
	ID() string
	Tx() T
	Hash() string
	Status() Status
	Err() string
}

type ChainClient[K any, T any, N any, E any] interface {
	GetNonce(context.Context, K, T) (N, error)
	EstimateTx(context.Context, K, T, N) (E, error)
	SendTx(context.Context, K, T, N, E) (string, error)
	TxStatus(context.Context, string) (Status, string, error)
	IsFatalError(string) bool
}