package client

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
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
	lggr logger.Logger
}

func NewClient(endpoint string, lggr logger.Logger) (*Client, error) {
	var rpc interface{}
	return &Client{
		rpc:  &rpc,
		lggr: lggr,
	}, nil
}
