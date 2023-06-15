package ocr2

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"

	caigorpc "github.com/smartcontractkit/caigo/rpcv02"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

//go:generate mockery --name OCR2Reader --output ./mocks/

type OCR2Reader interface {
	LatestConfigDetails(context.Context, caigotypes.Felt) (ContractConfigDetails, error)
	LatestTransmissionDetails(context.Context, caigotypes.Felt) (TransmissionDetails, error)
	LatestRoundData(context.Context, caigotypes.Felt) (RoundData, error)
	LinkAvailableForPayment(context.Context, caigotypes.Felt) (*big.Int, error)
	ConfigFromEventAt(context.Context, caigotypes.Felt, uint64) (ContractConfig, error)
	NewTransmissionsFromEventsAt(context.Context, caigotypes.Felt, uint64) ([]NewTransmissionEvent, error)
	BillingDetails(context.Context, caigotypes.Felt) (BillingDetails, error)

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

func (c *Client) BillingDetails(ctx context.Context, address caigotypes.Felt) (bd BillingDetails, err error) {
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

	observationPayment := caigotypes.StrToFelt(res[0]).Big()
	transmissionPayment := caigotypes.StrToFelt(res[1]).Big()

	bd, err = NewBillingDetails(observationPayment, transmissionPayment)
	if err != nil {
		return bd, errors.Wrap(err, "couldn't initialize billing details")
	}

	return
}

func (c *Client) LatestConfigDetails(ctx context.Context, address caigotypes.Felt) (ccd ContractConfigDetails, err error) {
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

	blockNum := caigotypes.StrToFelt(res[1])
	configDigest := caigotypes.StrToFelt(res[2])

	ccd, err = NewContractConfigDetails(blockNum.Big(), configDigest.Bytes())
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't initialize config details")
	}

	return
}

func (c *Client) LatestTransmissionDetails(ctx context.Context, address caigotypes.Felt) (td TransmissionDetails, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_transmission_details",
	}

	res, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return td, errors.Wrap(err, "couldn't call the contract")
	}

	// [0] - config digest, [1] - epoch and round, [2] - latest answer, [3] - latest timestamp
	if len(res) != 4 {
		return td, errors.New("unexpected result length")
	}

	digest := caigotypes.StrToFelt(res[0])
	configDigest := types.ConfigDigest{}
	digest.Big().FillBytes(configDigest[:])

	epoch, round := parseEpochAndRound(caigotypes.StrToFelt(res[1]).Big())

	latestAnswer, err := starknet.HexToUnsignedBig(res[2])
	if err != nil {
		return td, errors.Wrap(err, "latestAnswer invalid")
	}

	timestampFelt := caigotypes.StrToFelt(res[3])
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

func (c *Client) LatestRoundData(ctx context.Context, address caigotypes.Felt) (round RoundData, err error) {
	ops := starknet.CallOps{
		ContractAddress: address,
		Selector:        "latest_round_data",
	}

	results, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return round, errors.Wrap(err, "couldn't call the contract with selector latest_round_data")
	}
	felts := []caigotypes.Felt{}
	for _, result := range results {
		felts = append(felts, caigotypes.StrToFelt(result))
	}

	round, err = NewRoundData(felts)
	if err != nil {
		return round, errors.Wrap(err, "unable to decode RoundData")
	}
	return round, nil
}

func (c *Client) LinkAvailableForPayment(ctx context.Context, address caigotypes.Felt) (*big.Int, error) {
	results, err := c.r.CallContract(ctx, starknet.CallOps{
		ContractAddress: address,
		Selector:        "link_available_for_payment",
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call the contract with selector 'link_available_for_payment'")
	}
	if len(results) != 1 {
		return nil, errors.Wrap(err, "insufficient data from selector 'link_available_for_payment'")
	}
	return caigotypes.HexToBN(results[0]), nil
}

func (c *Client) fetchEventsFromBlock(ctx context.Context, address caigotypes.Felt, eventType string, blockNum uint64) (eventsAsFeltArrs [][]caigotypes.Felt, err error) {
	block := caigorpc.WithBlockNumber(blockNum)

	eventKey := caigotypes.BigToHex(caigotypes.GetSelectorFromName(eventType))

	input := caigorpc.EventsInput{
		EventFilter: caigorpc.EventFilter{
			FromBlock: block,
			ToBlock:   block,
			Address:   address,
			Keys:      [][]string{{eventKey}}, // skip other event types
			// PageSize:   0,
			// PageNumber: 0,
		},
		ResultPageRequest: caigorpc.ResultPageRequest{
			// ContinuationToken: ,
			ChunkSize: 10,
		},
	}
	events, err := c.r.Events(ctx, input)

	// TODO: check events.isLastPage, query more if needed

	if err != nil {
		return eventsAsFeltArrs, errors.Wrap(err, "couldn't fetch events for block")
	}

	for _, event := range events.Events {
		// convert to felts
		felts := []caigotypes.Felt{}
		for _, felt := range event.Data {
			felts = append(felts, caigotypes.StrToFelt(felt))
		}
		eventsAsFeltArrs = append(eventsAsFeltArrs, felts)
	}
	if len(eventsAsFeltArrs) == 0 {
		return nil, errors.New("events not found in the block")
	}
	return eventsAsFeltArrs, nil
}

func (c *Client) ConfigFromEventAt(ctx context.Context, address caigotypes.Felt, blockNum uint64) (cc ContractConfig, err error) {
	eventsAsFeltArrs, err := c.fetchEventsFromBlock(ctx, address, "ConfigSet", blockNum)
	if err != nil {
		return cc, errors.Wrap(err, "failed to fetch config_set events")
	}
	if len(eventsAsFeltArrs) != 1 {
		return cc, fmt.Errorf("expected to find one config_set event in block %d for address %s but found %d", blockNum, address, len(eventsAsFeltArrs))
	}
	configAtEvent := eventsAsFeltArrs[0]
	config, err := ParseConfigSetEvent(configAtEvent)
	if err != nil {
		return cc, errors.Wrap(err, "couldn't parse config event")
	}
	return ContractConfig{
		Config:      config,
		ConfigBlock: blockNum,
	}, nil
}

// NewTransmissionsFromEventsAt finds events of type new_transmission emitted by the contract address in a given block number.
func (c *Client) NewTransmissionsFromEventsAt(ctx context.Context, address caigotypes.Felt, blockNum uint64) (events []NewTransmissionEvent, err error) {
	eventsAsFeltArrs, err := c.fetchEventsFromBlock(ctx, address, "NewTransmission", blockNum)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch new_transmission events")
	}
	if len(eventsAsFeltArrs) == 0 {
		return nil, fmt.Errorf("expected to find at least one new_transmission event in block %d for address %s but found %d", blockNum, address, len(eventsAsFeltArrs))
	}
	events = []NewTransmissionEvent{}
	for _, felts := range eventsAsFeltArrs {
		event, err := ParseNewTransmissionEvent(felts)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't parse new_transmission event")
		}
		events = append(events, event)
	}
	return events, nil
}
