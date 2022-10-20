package starknet

import (
	"context"
	"time"

	"github.com/pkg/errors"

	caigorpc "github.com/smartcontractkit/caigo/rpcv02"
	caigotypes "github.com/smartcontractkit/caigo/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

//go:generate mockery --name Reader --output ./mocks/

type Reader interface {
	CallContract(context.Context, CallOps) ([]string, error)
	LatestBlockHeight(context.Context) (uint64, error)

	// provider interface
	BlockWithTxHashes(ctx context.Context, blockID caigorpc.BlockID) (*caigorpc.Block, error)
	Call(context.Context, caigotypes.FunctionCall, caigorpc.BlockID) ([]string, error)
	ChainID(context.Context) (string, error)
	Events(ctx context.Context, filter caigorpc.EventFilter) (*caigorpc.EventsOutput, error)
}

type Writer interface {
	// Invoke(context.Context, caigotypes.FunctionInvoke) (*caigotypes.AddInvokeTransactionOutput, error)
	TransactionByHash(context.Context, caigotypes.Hash) (caigorpc.Transaction, error)
	TransactionReceipt(context.Context, caigotypes.Hash) (caigorpc.TransactionReceipt, error)
	EstimateFee(context.Context, caigotypes.FunctionInvoke, caigorpc.BlockID) (*caigotypes.FeeEstimate, error)
}

type ReaderWriter interface {
	Reader
	Writer
}

var _ ReaderWriter = (*Client)(nil)

// var _ caigotypes.Provider = (*Client)(nil)

type Client struct {
	Provider       *caigorpc.Provider
	lggr           logger.Logger
	defaultTimeout time.Duration
}

// pass nil or 0 to timeout to not use built in default timeout
func NewClient(chainID string, baseURL string, lggr logger.Logger, timeout *time.Duration) (*Client, error) {
	// TODO: chainID now unused
	c, err := ethrpc.DialContext(context.Background(), baseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Provider: caigorpc.NewProvider(c),
		lggr:     lggr,
	}

	// make copy to preserve value
	// defensive in case the timeout reference is ever garbage collected or removed
	if timeout == nil {
		client.defaultTimeout = 0
	} else {
		client.defaultTimeout = *timeout
	}

	return client, nil
}

// -- Custom Wrapped Func --

func (c *Client) CallContract(ctx context.Context, ops CallOps) (res []string, err error) {
	tx := caigotypes.FunctionCall{
		ContractAddress:    ops.ContractAddress,
		EntryPointSelector: ops.Selector,
		Calldata:           ops.Calldata,
	}

	res, err = c.Call(ctx, tx, caigorpc.WithBlockTag("pending"))
	if err != nil {
		return res, errors.Wrap(err, "error in client.CallContract")
	}

	return
}

func (c *Client) LatestBlockHeight(ctx context.Context) (height uint64, err error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	blockNum, err := c.Provider.BlockNumber(ctx)
	if err != nil {
		return height, errors.Wrap(err, "error in client.LatestBlockHeight")
	}

	return blockNum, nil
}

// -- caigo.Provider interface --

func (c *Client) BlockWithTxHashes(ctx context.Context, blockID caigorpc.BlockID) (*caigorpc.Block, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.BlockWithTxHashes(ctx, blockID)
	if err != nil {
		return out.(*caigorpc.Block), errors.Wrap(err, "error in client.BlockWithTxHashes")
	}
	return out.(*caigorpc.Block), nil
}

func (c *Client) Call(ctx context.Context, calls caigotypes.FunctionCall, blockHashOrTag caigorpc.BlockID) ([]string, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.Call(ctx, calls, blockHashOrTag)
	if err != nil {
		return out, errors.Wrap(err, "error in client.Call")
	}
	if out == nil {
		return out, NilResultError("client.Call")
	}
	return out, nil

}

func (c *Client) ChainID(ctx context.Context) (string, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.ChainID(ctx)
	if err != nil {
		return out, errors.Wrap(err, "error in client.ChainID")
	}
	return out, nil

}

func (c *Client) TransactionByHash(ctx context.Context, hash caigotypes.Hash) (caigorpc.Transaction, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.TransactionByHash(ctx, hash)
	if err != nil {
		return out, errors.Wrap(err, "error in client.TransactionByHash")
	}
	if out == nil {
		return out, NilResultError("client.TransactionByHash")
	}
	return out, nil

}

func (c *Client) TransactionReceipt(ctx context.Context, hash caigotypes.Hash) (caigorpc.TransactionReceipt, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.TransactionReceipt(ctx, hash)
	if err != nil {
		return out, errors.Wrap(err, "error in client.TransactionReceipt")
	}
	if out == nil {
		return out, NilResultError("client.TransactionReceipt")
	}
	return out, nil

}

func (c *Client) EstimateFee(ctx context.Context, call caigotypes.FunctionInvoke, blockID caigorpc.BlockID) (*caigotypes.FeeEstimate, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.EstimateFee(ctx, call, blockID)
	if err != nil {
		return out, errors.Wrap(err, "error in client.EstimateFee")
	}
	if out == nil {
		return out, NilResultError("client.EstimateFee")
	}
	return out, nil

}

func (c *Client) Events(ctx context.Context, filter caigorpc.EventFilter) (*caigorpc.EventsOutput, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.Events(ctx, caigorpc.EventsInput{
		EventFilter: filter,
		// TODO: ResultPageRequest: ResultPageRequest { ContinuationToken: , ChunkSize: }
	})
	if err != nil {
		return out, errors.Wrap(err, "error in client.Events")
	}
	if out == nil {
		return out, NilResultError("client.Events")
	}
	return out, nil

}
