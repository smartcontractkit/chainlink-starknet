package starknet

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"

	caigogw "github.com/dontpanicdao/caigo/gateway"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type Reader interface {
	ocr2.Reader
	ChainID(context.Context) (string, error)

	callContract(context.Context, string, string, ...string) ([]string, error)
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

func (c *Client) callContract(ctx context.Context, address string, method string, params ...string) (res []string, err error) {
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

func (c *Client) ChainID(ctx context.Context) (id string, err error) {
	id, err = c.gw.ChainID(ctx)
	if err != nil {
		return
	}

	return id, nil
}

func (c *Client) LatestBlockHeight(ctx context.Context) (height uint64, err error) {
	block, err := c.gw.Block(ctx, nil)
	if err != nil {
		return
	}

	return uint64(block.BlockNumber), nil
}

func (c *Client) OCR2BillingDetails(ctx context.Context, address string) (bd ocr2.BillingDetails, err error) {
	res, err := c.callContract(ctx, address, "billing")
	if err != nil {
		return
	}

	if len(res) != 2 {
		return bd, errors.New("unexpected result length")
	}

	bd, err = ocr2.NewBillingDetails(res[0], res[1])
	return
}

func (c *Client) OCR2LatestConfigDetails(ctx context.Context, address string) (ccd ocr2.ContractConfigDetails, err error) {
	res, err := c.callContract(ctx, address, "latest_config_details")
	if err != nil {
		return
	}

	// [0] - config count, [1] - block number, [2] - config digest
	if len(res) != 3 {
		return ccd, errors.New("unexpected result length")
	}

	ccd, err = ocr2.NewContractConfigDetails(res[1], res[2])
	return
}

func (c *Client) OCR2LatestConfig(ctx context.Context, address string, blockNum uint64) (cc ocr2.ContractConfig, err error) {
	block, err := c.gw.Block(ctx, &caigogw.BlockOptions{
		BlockNumber: blockNum,
	})
	if err != nil {
		return
	}

	for _, txReceipt := range block.TransactionReceipts {
		for _, event := range txReceipt.Events {
			var decodedEvent caigotypes.Event

			m, err := json.Marshal(event)
			if err != nil {
				return cc, errors.Wrap(err, "couldn't marshal event")
			}

			err = json.Unmarshal(m, &decodedEvent)
			if err != nil {
				return cc, errors.Wrap(err, "couldn't unmarshal event")
			}

			if isSetConfigEventFromContract(&decodedEvent, address) {
				config, err := parseConfigEventData(decodedEvent.Data)
				if err != nil {
					return cc, errors.Wrap(err, "couldn't parse config event")
				}

				return ocr2.ContractConfig{
					Config:      config,
					ConfigBlock: blockNum,
				}, nil
			}
		}
	}

	return cc, errors.New("config not found in the block")
}
