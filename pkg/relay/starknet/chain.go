package starknet

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-starknet/pkg/relay/starknet/db"
)

type ChainSet interface {
	types.Service

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	types.Service

	ID() string
	Config() Config
	UpdateConfig(*db.ChainCfg)
}
