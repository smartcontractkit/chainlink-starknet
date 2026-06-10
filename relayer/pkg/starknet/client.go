package starknet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/client"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

//go:generate mockery --name Reader --output ./mocks/

type Reader interface {
	CallContract(context.Context, CallOps) ([]*felt.Felt, error)
	LatestBlockHeight(context.Context) (uint64, error)

	// provider interface
	BlockWithTxHashes(ctx context.Context, blockID starknetrpc.BlockID) (*starknetrpc.Block, error)
	Call(context.Context, starknetrpc.FunctionCall, starknetrpc.BlockID) ([]*felt.Felt, error)
	Events(ctx context.Context, input starknetrpc.EventsInput) (*starknetrpc.EventChunk, error)
	AccountNonce(context.Context, *felt.Felt) (*felt.Felt, error)
}

type Writer interface {
}

type ReaderWriter interface {
	Reader
	Writer
}

var _ ReaderWriter = (*Client)(nil)

type Client struct {
	Provider       starknetrpc.RPCProvider
	EthClient      *ethrpc.Client
	lggr           logger.Logger
	defaultTimeout time.Duration
}

// pass nil or 0 to timeout to not use built in default timeout
func NewClient(chainID string, baseURL string, apiKey string, lggr logger.Logger, timeout *time.Duration) (*Client, error) {
	// TODO: chainID now unused

	options := []client.ClientOption{}
	if strings.TrimSpace(apiKey) != "" {
		options = append(options, client.WithHeader("x-apikey", apiKey))
	}

	provider, err := starknetrpc.NewProvider(context.Background(), baseURL, options...)
	if err != nil && !errors.Is(err, starknetrpc.ErrIncompatibleVersion) {
		return nil, err
	}
	if err != nil {
		// starknet.go v0.17.x targets RPC 0.9.0; production nodes on 0.10.x still work.
		lggr.Warnw("starknet RPC spec version mismatch", "error", err)
	}

	c, err := ethrpc.DialContext(context.Background(), baseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Provider:  provider,
		EthClient: c,
		lggr:      lggr,
	}

	if timeout == nil {
		client.defaultTimeout = 0
	} else {
		client.defaultTimeout = *timeout
	}

	return client, nil
}

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.defaultTimeout == 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.defaultTimeout)
}

func (c *Client) CallContract(ctx context.Context, ops CallOps) (data []*felt.Felt, err error) {
	tx := starknetrpc.FunctionCall{
		ContractAddress:    ops.ContractAddress,
		EntryPointSelector: ops.Selector,
		Calldata:           ops.Calldata,
	}

	res, err := c.Call(ctx, tx, LatestBlockID())
	if err != nil {
		return nil, fmt.Errorf("error in client.CallContract: %w", err)
	}

	return res, nil
}

func (c *Client) LatestBlockHeight(ctx context.Context) (uint64, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	blockNum, err := c.Provider.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("error in client.LatestBlockHeight: %w", err)
	}

	return blockNum, nil
}

func (c *Client) BlockWithTxHashes(ctx context.Context, blockID starknetrpc.BlockID) (*starknetrpc.Block, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	out, err := c.Provider.BlockWithTxHashes(ctx, blockID)
	if err != nil {
		return out.(*starknetrpc.Block), fmt.Errorf("error in client.BlockWithTxHashes: %w", err)
	}
	return out.(*starknetrpc.Block), nil
}

func (c *Client) Call(ctx context.Context, calls starknetrpc.FunctionCall, blockHashOrTag starknetrpc.BlockID) ([]*felt.Felt, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	out, err := c.Provider.Call(ctx, calls, blockHashOrTag)
	if err != nil {
		return out, fmt.Errorf("error in client.Call: %w", err)
	}
	if out == nil {
		return out, NilResultError("client.Call")
	}
	return out, nil
}

func (c *Client) Events(ctx context.Context, input starknetrpc.EventsInput) (*starknetrpc.EventChunk, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	out, err := c.Provider.Events(ctx, input)
	if err != nil {
		return out, fmt.Errorf("error in client.Events: %w", err)
	}
	if out == nil {
		return out, NilResultError("client.Events")
	}
	return out, nil
}

func (c *Client) AccountNonce(ctx context.Context, accountAddress *felt.Felt) (*felt.Felt, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.Provider.Nonce(ctx, PreConfirmedBlockID(), accountAddress)
}

func (c *Client) AccountNonceLatest(ctx context.Context, accountAddress *felt.Felt) (*felt.Felt, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.Provider.Nonce(ctx, LatestBlockID(), accountAddress)
}

// EstimateFeeAtPreConfirmed estimates fees against the pre_confirmed block state.
func (c *Client) EstimateFeeAtPreConfirmed(
	ctx context.Context,
	txns []starknetrpc.BroadcastTxn,
	flags []starknetrpc.SimulationFlag,
) ([]starknetrpc.FeeEstimation, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.Provider.EstimateFee(ctx, txns, flags, PreConfirmedBlockID())
}
