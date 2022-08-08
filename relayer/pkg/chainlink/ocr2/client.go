package ocr2

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	caigo "github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type OCR2Reader interface {
	LatestConfigDetails(context.Context, string) (ContractConfigDetails, error)
	LatestTransmissionDetails(context.Context, string) (TransmissionDetails, error)
	ConfigFromEventAt(context.Context, string, uint64) (ContractConfig, error)
	BillingDetails(context.Context, string) (BillingDetails, error)

	BaseReader() starknet.Reader
}

var _ OCR2Reader = (*Client)(nil)

type Client struct {
	r    starknet.Reader
	lggr logger.Logger
}

func NewClient(reader starknet.Reader, lggr logger.Logger) (*Client, error) {
	return &Client{
		r:    reader,
		lggr: lggr,
	}, nil
}

func (c *Client) BaseReader() starknet.Reader {
	return c.r
}

func (c *Client) BillingDetails(ctx context.Context, address string) (bd BillingDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "billing",
	}

	res, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return bd, errors.Wrap(err, "couldn't call the contract")
	}

	if len(res) != 2 {
		return bd, errors.New("unexpected result length")
	}

	observationPaymentFelt, err := caigoStringToJunoFelt(res[0])
	if err != nil {
		return bd, errors.Wrap(err, "couldn't decode observation payment")
	}

	transmissionPaymentFelt, err := caigoStringToJunoFelt(res[1])
	if err != nil {
		return bd, errors.Wrap(err, "couldn't decode transmission payment")
	}

	bd, err = NewBillingDetails(observationPaymentFelt, transmissionPaymentFelt)
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

	res, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't call the contract")
	}

	// [0] - config count, [1] - block number, [2] - config digest
	if len(res) != 3 {
		return ccd, errors.New("unexpected result length")
	}

	blockNum, err := caigoStringToJunoFelt(res[1])
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't decode block num")
	}

	configDigest, err := caigoStringToJunoFelt(res[2])
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't decode config digest")
	}

	ccd, err = NewContractConfigDetails(blockNum, configDigest)
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't initialize config details")
	}

	return
}

func (c *Client) LatestTransmissionDetails(ctx context.Context, address string) (td TransmissionDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_round_data",
	}

	res, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return td, errors.Wrap(err, "couldn't call the contract")
	}

	// [0] - round_id, [1] - answer, [2] - block_num,
	// [3] - observation_timestamp, [4] - transmission_timestamp
	blockNumFelt, err := caigoStringToJunoFelt(res[2])
	if err != nil {
		return td, errors.Wrap(err, "couldn't decode block num")
	}

	blockNum := uint64(blockNumFelt.Big().Int64())
	if blockNum == 0 {
		c.lggr.Warn("No transmissions found")
		return td, nil
	}

	block, err := c.r.BlockByNumberGateway(ctx, blockNum)
	if err != nil {
		return td, errors.Wrap(err, "couldn't fetch block by number")
	}

	for _, txReceipt := range block.TransactionReceipts {
		for _, event := range txReceipt.Events {
			var decodedEvent caigotypes.Event

			m, err := json.Marshal(event)
			if err != nil {
				return td, errors.Wrap(err, "couldn't marshal event")
			}

			err = json.Unmarshal(m, &decodedEvent)
			if err != nil {
				return td, errors.Wrap(err, "couldn't unmarshal event")
			}

			if isEventFromContract(&decodedEvent, address, "new_transmission") {
				return parseTransmissionEventData(decodedEvent.Data)
			}
		}
	}

	return td, errors.New("transmission details not found")
}

func (c *Client) ConfigFromEventAt(ctx context.Context, address string, blockNum uint64) (cc ContractConfig, err error) {
	block, err := c.r.BlockByNumberGateway(ctx, blockNum+1) // HAXX: temporary workaround for devnet 0.2.8
	if err != nil {
		return cc, errors.Wrap(err, "couldn't fetch block by number")
	}

	c.lggr.Errorf("Fetching block number %v", blockNum)

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

			eventKey := caigo.GetSelectorFromName("config_set")
			c.lggr.Errorf("Checking %v if address=%v selector %v", decodedEvent, address, eventKey)

			if isEventFromContract(&decodedEvent, address, "config_set") {
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
