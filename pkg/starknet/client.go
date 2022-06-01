package starknet

import (
	"context"
	"strconv"

	caigogw "github.com/dontpanicdao/caigo/gateway"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type Reader interface {
	ocr2.Reader
	ChainID(context.Context) (string, error)
	CallContract(context.Context, string, string, ...string) ([]string, error)
}

type Writer interface {
}

type ReaderWriter interface {
	Reader
	Writer
}

// verify Client implements ReaderWriter
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

func (c *Client) ChainID(ctx context.Context) (id string, err error) {
	id, err = c.gw.ChainID(ctx)
	if err != nil {
		return
	}

	return id, nil
}

func (c *Client) CallContract(ctx context.Context, address string, method string, params ...string) (res []string, err error) {
	tx := caigotypes.Transaction{
		ContractAddress:    address,
		EntryPointSelector: method,
		Calldata:           params,
	}

	res, err = c.gw.Call(ctx, tx, nil)
	if err != nil {
		return
	}

	return
}

func (c *Client) OCR2ReadLatestConfigDetails(ctx context.Context, address string) (details ocr2.ContractConfigDetails, err error) {
	res, err := c.CallContract(ctx, address, "latest_config_details")
	if err != nil {
		return
	}

	// todo: validate res
	// todo: proper felt convertion

	block, err := strconv.ParseUint(res[0], 10, 64)
	if err != nil {
		return
	}

	var digest [32]byte
	copy(digest[:], res[1])

	details = ocr2.ContractConfigDetails{
		Block:  block,
		Digest: digest,
	}

	return
}
