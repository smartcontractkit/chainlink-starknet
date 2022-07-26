package starknet

import (
	"context"
	"math/big"
	"time"

	"github.com/pkg/errors"

	caigogw "github.com/dontpanicdao/caigo/gateway"
	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type Reader interface {
	CallContract(context.Context, CallOps) ([]string, error)
	LatestBlockHeight(context.Context) (uint64, error)
	BlockByNumberGateway(context.Context, uint64) (*caigogw.Block, error)

	// provider interface
	BlockByHash(context.Context, string, string) (*caigotypes.Block, error)
	BlockByNumber(context.Context, *big.Int, string) (*caigotypes.Block, error)
	Call(context.Context, caigotypes.FunctionCall, string) ([]string, error)
	ChainID(context.Context) (string, error)
}

type Writer interface {
	AccountNonce(context.Context, string) (*big.Int, error)
	Invoke(context.Context, caigotypes.Transaction) (*caigotypes.AddTxResponse, error)
	TransactionByHash(context.Context, string) (*caigotypes.Transaction, error)
	TransactionReceipt(context.Context, string) (*caigotypes.TransactionReceipt, error)
	EstimateFee(context.Context, caigotypes.Transaction) (*caigotypes.FeeEstimate, error)
}

type Unimplemented interface {
	Class(context.Context, string) (*caigotypes.ContractClass, error)
	ClassHashAt(context.Context, string) (*caigotypes.Felt, error)
	ClassAt(context.Context, string) (*caigotypes.ContractClass, error)
}

type ReaderWriter interface {
	Reader
	Writer
	Unimplemented
}

var _ ReaderWriter = (*Client)(nil)

var _ caigotypes.Provider = (*Client)(nil)

type Client struct {
	gw             *caigogw.GatewayProvider
	lggr           logger.Logger
	defaultTimeout *time.Duration
}

// pass nil to timeout to not use built in default timeout
func NewClient(chainID string, baseURL string, lggr logger.Logger, timeout *time.Duration) (*Client, error) {
	client := &Client{
		gw:             caigogw.NewProvider(caigogw.WithChain(chainID)),
		lggr:           lggr,
		defaultTimeout: timeout,
	}
	client.setURL(baseURL) // hack: change the base URL (not supported in caigo)

	return client, nil
}

func (c *Client) setURL(baseURL string) {
	if baseURL == "" {
		return // if empty, use default from caigo
	}

	c.gw.Gateway.Base = baseURL
	c.gw.Gateway.Feeder = baseURL + "/feeder_gateway"
	c.gw.Gateway.Gateway = baseURL + "/gateway"
}

// -- Custom Wrapped Func --

func (c *Client) CallContract(ctx context.Context, ops CallOps) (res []string, err error) {
	tx := caigotypes.FunctionCall{
		ContractAddress:    ops.ContractAddress,
		EntryPointSelector: ops.Selector,
		Calldata:           ops.Calldata,
	}

	res, err = c.Call(ctx, tx, "")
	if err != nil {
		return res, errors.Wrap(err, "error in client.CallContract")
	}

	return
}

func (c *Client) LatestBlockHeight(ctx context.Context) (height uint64, err error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	block, err := c.gw.Block(ctx, nil)
	if err != nil {
		return height, errors.Wrap(err, "error in client.LatestBlockHeight")
	}

	return uint64(block.BlockNumber), nil
}

func (c *Client) BlockByNumberGateway(ctx context.Context, blockNum uint64) (block *caigogw.Block, err error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	block, err = c.gw.Block(ctx, &caigogw.BlockOptions{
		BlockNumber: blockNum,
	})
	if err != nil {
		return block, errors.Wrap(err, "couldn't get block by number")
	}

	return block, nil
}

// -- caigo.Provider interface --

func (c *Client) BlockByHash(ctx context.Context, hash string, _ string) (*caigotypes.Block, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.BlockByHash(ctx, hash, "")
	if err != nil {
		return out, errors.Wrap(err, "error in client.BlockByHash")
	}
	return out, nil

}

func (c *Client) BlockByNumber(ctx context.Context, num *big.Int, _ string) (*caigotypes.Block, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.BlockByNumber(ctx, num, "")
	if err != nil {
		return out, errors.Wrap(err, "error in client.BlockByNumber")
	}
	return out, nil

}

func (c *Client) Call(ctx context.Context, calls caigotypes.FunctionCall, blockHashOrTag string) ([]string, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.Call(ctx, calls, blockHashOrTag)
	if err != nil {
		return out, errors.Wrap(err, "error in client.Call")
	}
	return out, nil

}

func (c *Client) ChainID(ctx context.Context) (string, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.ChainID(ctx)
	if err != nil {
		return out, errors.Wrap(err, "error in client.ChainID")
	}
	return out, nil

}

func (c *Client) AccountNonce(ctx context.Context, address string) (*big.Int, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.AccountNonce(ctx, address)
	if err != nil {
		return out, errors.Wrap(err, "error in client.AccountNonce")
	}
	return out, nil

}

func (c *Client) Invoke(ctx context.Context, txs caigotypes.Transaction) (*caigotypes.AddTxResponse, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.Invoke(ctx, txs)
	if err != nil {
		return out, errors.Wrap(err, "error in client.Invoke")
	}
	return out, nil

}

func (c *Client) TransactionByHash(ctx context.Context, hash string) (*caigotypes.Transaction, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.TransactionByHash(ctx, hash)
	if err != nil {
		return out, errors.Wrap(err, "error in client.TransactionByHash")
	}
	return out, nil

}

func (c *Client) TransactionReceipt(ctx context.Context, hash string) (*caigotypes.TransactionReceipt, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.TransactionReceipt(ctx, hash)
	if err != nil {
		return out, errors.Wrap(err, "error in client.TransactionReceipt")
	}
	return out, nil

}

func (c *Client) EstimateFee(ctx context.Context, txs caigotypes.Transaction) (*caigotypes.FeeEstimate, error) {
	if c.defaultTimeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *c.defaultTimeout)
		defer cancel()
	}

	out, err := c.gw.EstimateFee(ctx, txs)
	if err != nil {
		return out, errors.Wrap(err, "error in client.EstimateFee")
	}
	return out, nil

}

// -- unimplemented provider interface --

func (c *Client) Class(context.Context, string) (*caigotypes.ContractClass, error) {
	return nil, errors.New("client.Class is not implemented")
}

func (c *Client) ClassHashAt(context.Context, string) (*caigotypes.Felt, error) {
	return nil, errors.New("client.ClassHashAt is not implemented")
}

func (c *Client) ClassAt(context.Context, string) (*caigotypes.ContractClass, error) {
	return nil, errors.New("client.ClassAt is not implemented")
}
