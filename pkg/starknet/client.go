package starknet

import (
<<<<<<< HEAD
	"context"
	"github.com/pkg/errors"

	caigogw "github.com/dontpanicdao/caigo/gateway"
	caigotypes "github.com/dontpanicdao/caigo/types"
=======
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"
>>>>>>> af017e4 (Revert /relayer subdirectory)

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type Reader interface {
<<<<<<< HEAD
	ChainID(context.Context) (string, error)
	LatestBlockHeight(context.Context) (uint64, error)
	BlockByNumber(context.Context, uint64) (*caigogw.Block, error)

	CallContract(context.Context, CallOps) ([]string, error)
}

type Writer interface {
=======
	ocr2.Reader
	// RPC read interface
}

type Writer interface {
	// RPC write interface
>>>>>>> af017e4 (Revert /relayer subdirectory)
}

type ReaderWriter interface {
	Reader
	Writer
}

<<<<<<< HEAD
var _ ReaderWriter = (*Client)(nil)

type Client struct {
	gw   *caigogw.Gateway
	lggr logger.Logger
}

func NewClient(chainID string, lggr logger.Logger) (*Client, error) {
	return &Client{
		gw:   caigogw.NewClient(caigogw.WithChain(chainID)),
		lggr: lggr,
	}, nil
}

func (c *Client) CallContract(ctx context.Context, ops CallOps) (res []string, err error) {
	tx := caigotypes.Transaction{
		ContractAddress:    ops.ContractAddress,
		EntryPointSelector: ops.Selector,
		Calldata:           ops.Calldata,
	}

	res, err = c.gw.Call(ctx, tx, nil)
	if err != nil {
		return res, errors.Wrap(err, "couldn't call the contract")
	}

	return
}

func (c *Client) ChainID(ctx context.Context) (id string, err error) {
	id, err = c.gw.ChainID(ctx)
	if err != nil {
		return id, errors.Wrap(err, "couldn't get chain id")
	}

	return id, nil
}

func (c *Client) LatestBlockHeight(ctx context.Context) (height uint64, err error) {
	block, err := c.gw.Block(ctx, nil)
	if err != nil {
		return height, errors.Wrap(err, "couldn't get latest block height")
	}

	return uint64(block.BlockNumber), nil
}

func (c *Client) BlockByNumber(ctx context.Context, blockNum uint64) (block *caigogw.Block, err error) {
	block, err = c.gw.Block(ctx, &caigogw.BlockOptions{
		BlockNumber: blockNum,
	})
	if err != nil {
		return block, errors.Wrap(err, "couldn't get block by number")
	}

	return block, nil
}
=======
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
>>>>>>> af017e4 (Revert /relayer subdirectory)
