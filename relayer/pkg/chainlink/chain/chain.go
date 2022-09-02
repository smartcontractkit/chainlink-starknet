package starknet

import (
	"context"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/config"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/db"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/core"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	// unused module to keep in go.mod and prevent ambiguous import
	_ "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type ChainSet interface {
	types.Service

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	types.Service

	Config() config.Config
	UpdateConfig(*db.ChainCfg)

	TxManager() core.TxQueue[caigotypes.Transaction]
	Reader() (starknet.Reader, error)
}
