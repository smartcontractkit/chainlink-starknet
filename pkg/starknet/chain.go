package starknet

import (
	"context"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/client"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/config"
	"github.com/smartcontractkit/chainlink/core/services"
)

type ChainSet interface {
	services.ServiceCtx

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	services.ServiceCtx

	ID() string
	Config() config.Config
	Reader() (client.Reader, error)
}
