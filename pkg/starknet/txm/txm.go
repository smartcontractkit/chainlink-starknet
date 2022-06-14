package txm

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

type TransactionManager interface {
	types.Service
	Enqueue() error
}

type starktxm struct {
	starter utils.StartStopOnce
}

func New() TransactionManager {
	return &starktxm{}
}

func (txm *starktxm) Start(ctx context.Context) error {
	return nil
}

func (txm *starktxm) Close() error {
	return nil
}

func (txm *starktxm) Healthy() error {
	return nil
}

func (txm *starktxm) Ready() error {
	return nil
}

func (txm starktxm) Enqueue() error {
  return nil
}
