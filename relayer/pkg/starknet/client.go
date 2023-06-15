package starknet

import (
	"context"
	"math/big"
	"time"

	"github.com/pkg/errors"

	ethrpc "github.com/ethereum/go-ethereum/rpc"
	caigo "github.com/smartcontractkit/caigo"
	caigorpc "github.com/smartcontractkit/caigo/rpcv02"
	caigotypes "github.com/smartcontractkit/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

//go:generate mockery --name Reader --output ./mocks/

type Reader interface {
	CallContract(context.Context, CallOps) ([]string, error)
	LatestBlockHeight(context.Context) (uint64, error)

	// provider interface
	BlockWithTxHashes(ctx context.Context, blockID caigorpc.BlockID) (*caigorpc.Block, error)
	Call(context.Context, caigotypes.FunctionCall, caigorpc.BlockID) ([]string, error)
	Events(ctx context.Context, input caigorpc.EventsInput) (*caigorpc.EventsOutput, error)
	TransactionByHash(context.Context, caigotypes.Felt) (caigorpc.Transaction, error)
	TransactionReceipt(context.Context, caigotypes.Felt) (caigorpc.TransactionReceipt, error)
	AccountNonce(context.Context, caigotypes.Felt) (*big.Int, error)
}

type Writer interface {
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
	ChainID        string
}

// pass nil or 0 to timeout to not use built in default timeout
func NewClient(_chainID string, baseURL string, lggr logger.Logger, timeout *time.Duration) (*Client, error) {
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

	// cache chainID on the provider to avoid repeated calls
	chainID, err := client.Provider.ChainID(context.TODO())
	if err != nil {
		return nil, err
	}
	client.ChainID = chainID

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

func (c *Client) TransactionByHash(ctx context.Context, hash caigotypes.Felt) (caigorpc.Transaction, error) {
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

func (c *Client) TransactionReceipt(ctx context.Context, hash caigotypes.Felt) (caigorpc.TransactionReceipt, error) {
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

func (c *Client) Events(ctx context.Context, input caigorpc.EventsInput) (*caigorpc.EventsOutput, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	out, err := c.Provider.Events(ctx, input)
	if err != nil {
		return out, errors.Wrap(err, "error in client.Events")
	}
	if out == nil {
		return out, NilResultError("client.Events")
	}
	return out, nil

}
func (c *Client) AccountNonce(ctx context.Context, accountAddress caigotypes.Felt) (*big.Int, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	sender := caigotypes.BigToFelt(big.NewInt((0))) // not actually used in account.Nonce()
	account, err := caigo.NewRPCAccount(sender, accountAddress, nil, c.Provider, caigo.AccountVersion1)
	if err != nil {
		return nil, errors.Wrap(err, "error in client.AccountNonce")
	}
	return account.Nonce(ctx)
}
