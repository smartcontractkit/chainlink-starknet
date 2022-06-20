package ocr2

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type OCR2Reader interface {
	LatestConfigDetails(context.Context, string) (ContractConfigDetails, error)
	ConfigFromEventAt(context.Context, string, uint64) (ContractConfig, error)
	BillingDetails(context.Context, string) (BillingDetails, error)

	BaseClient() *starknet.Client
}

var _ OCR2Reader = (*Client)(nil)

type Client struct {
	starknetClient *starknet.Client
	lggr           logger.Logger
}

func NewClient(chainID string, lggr logger.Logger) (*Client, error) {
	client, err := starknet.NewClient(chainID, lggr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initialize starknet client")
	}

	return &Client{
		starknetClient: client,
		lggr:           lggr,
	}, nil
}

func (c *Client) BaseClient() *starknet.Client {
	return c.starknetClient
}

func (c *Client) BillingDetails(ctx context.Context, address string) (bd BillingDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "billing",
	}

	res, err := c.starknetClient.CallContract(ctx, ops)
	if err != nil {
		return bd, errors.Wrap(err, "couldn't call the contract")
	}

	if len(res) != 2 {
		return bd, errors.New("unexpected result length")
	}

	bd, err = NewBillingDetails(res[0], res[1])
	if err != nil {
		return bd, errors.Wrap(err, "couldn't initialize billing details")
	}

	return
}

func (c *Client) LatestConfigDetails(ctx context.Context, address string) (ccd ContractConfigDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_config_details",
	}

	res, err := c.starknetClient.CallContract(ctx, ops)
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't call the contract")
	}

	// [0] - config count, [1] - block number, [2] - config digest
	if len(res) != 3 {
		return ccd, errors.New("unexpected result length")
	}

	ccd, err = NewContractConfigDetails(res[1], res[2])
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't initialize config details")
	}

	return
}

func (c *Client) ConfigFromEventAt(ctx context.Context, address string, blockNum uint64) (cc ContractConfig, err error) {
	block, err := c.starknetClient.BlockByNumber(ctx, blockNum)
	if err != nil {
		return cc, errors.Wrap(err, "couldn't fetch block by number")
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

				return ContractConfig{
					Config:      config,
					ConfigBlock: blockNum,
				}, nil
			}
		}
	}

	return cc, errors.New("config not found in the block")
}
