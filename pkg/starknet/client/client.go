package client

import (
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/config"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/logger"
)

type Reader interface {
	// RPC read interface
}

type Writer interface {
	// RPC write interface
}

type ReaderWriter interface {
	Reader
	Writer
}

// verify Client implements ReaderWriter
var _ ReaderWriter = (*Client)(nil)

type Client struct {
	rpc  *interface{} // todo: replace with RPC client
	cfg  config.Config
	lggr logger.Logger
}

func NewClient(endpoint string, cfg config.Config, lggr logger.Logger) (*Client, error) {
	var rpc interface{}
	return &Client{
		rpc:  &rpc,
		cfg:  cfg,
		lggr: lggr,
	}, nil
}
