package starknet

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
<<<<<<< HEAD
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/db"
=======
	"github.com/smartcontractkit/chainlink-starknet/pkg/relay/starknet/db"
>>>>>>> af017e4 (Revert /relayer subdirectory)
)

type ChainSet interface {
	types.Service

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	types.Service

	Config() Config
	UpdateConfig(*db.ChainCfg)
}
