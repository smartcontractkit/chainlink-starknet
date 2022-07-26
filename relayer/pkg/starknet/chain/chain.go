package starknet

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet/db"

	// unused module to keep in go.mod and prevent ambiguous import
	_ "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type ChainSet interface {
	types.Service

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	types.Service

	Config() starknet.Config
	UpdateConfig(*db.ChainCfg)

	TxManager() txm.TxManager
	Reader() (starknet.Reader, error)
}
