package starknet

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/config"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	// unused module to keep in go.mod and prevent ambiguous import
	_ "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type ChainSet = types.ChainSet[string, Chain]

type Chain interface {
	types.ChainService

	ID() string
	Config() config.Config

	TxManager() txm.TxManager
	Reader() (starknet.Reader, error)
}
