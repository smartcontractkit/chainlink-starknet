package ocr2

import (
	"context"
	"encoding/json"
	"time"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	"github.com/pkg/errors"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type OCR2Reader interface {
	LatestConfigDetails(context.Context, string) (ContractConfigDetails, error)
	LatestTransmissionDetails(context.Context, string) (TransmissionDetails, error)
	LatestRoundData(context.Context, string) (RoundData, error)
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

	// [0] - observation payment, [1] - transmission payment, [2] - gas base, [3] - gas per signature
	if len(res) != 4 {
		return bd, errors.New("unexpected result length")
	}

	observationPaymentFelt := junotypes.HexToFelt(res[0])
	transmissionPaymentFelt := junotypes.HexToFelt(res[1])

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

	blockNum := junotypes.HexToFelt(res[1])
	configDigest := junotypes.HexToFelt(res[2])

	ccd, err = NewContractConfigDetails(blockNum, configDigest)
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't initialize config details")
	}

	return
}

func (c *Client) LatestTransmissionDetails(ctx context.Context, address string) (td TransmissionDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_transmission_details",
	}

	res, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return td, errors.Wrap(err, "couldn't call the contract")
	}

	// [0] - config digest, [1] - epoch and round, [2] - latest answer, [3] - latest timestamp
	digest := junotypes.HexToFelt(res[0])
	configDigest := types.ConfigDigest{}
	digest.Big().FillBytes(configDigest[:])

	epoch, round := parseEpochAndRound(junotypes.HexToFelt(res[1]).Big())

	latestAnswer := starknet.HexToSignedBig(res[2])

	timestampFelt := junotypes.HexToFelt(res[3])
	// TODO: Int64() can return invalid data if int is too big
	unixTime := timestampFelt.Big().Int64()
	latestTimestamp := time.Unix(unixTime, 0)

	td = TransmissionDetails{
		Digest:          configDigest,
		Epoch:           epoch,
		Round:           round,
		LatestAnswer:    latestAnswer,
		LatestTimestamp: latestTimestamp,
	}

	return td, nil
}

func (c *Client) LatestRoundData(ctx context.Context, address string) (round RoundData, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_round_data",
	}

	results, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return round, errors.Wrap(err, "couldn't call the contract with selector latest_round_data")
	}
	felts := []junotypes.Felt{}
	for _, result := range results {
		felts = append(felts, junotypes.HexToFelt(result))
	}

	round, err = NewRoundData(felts)
	if err != nil {
		return round, errors.Wrap(err, "unable to decode RoundData")
	}
	return round, nil
}

func (c *Client) ConfigFromEventAt(ctx context.Context, address string, blockNum uint64) (cc ContractConfig, err error) {
	block, err := c.r.BlockByNumberGateway(ctx, blockNum)
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

			if starknet.IsEventFromContract(&decodedEvent, address, "ConfigSet") {
				config, err := ParseConfigSetEvent(decodedEvent.Data)
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
