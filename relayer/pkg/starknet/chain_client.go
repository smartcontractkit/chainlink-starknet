package starknet

import (
	"context"
	"fmt"

	"github.com/NethermindEth/juno/core/felt"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

// type alias for readibility
type FinalizedBlock = starknetrpc.Block

// used to create batch requests
type StarknetBatchBuilder interface {
	RequestBlockByHash(h *felt.Felt) (StarknetBatchBuilder, error)
	// RequestBlockByNumber(id uint64) (StarknetBatchBuilder, error)
	// RequestLatestPendingBlock() (StarknetBatchBuilder, error)
	// RequestLatestBlockHashAndNumber() (StarknetBatchBuilder, error)
	// RequestEventsByFilter(f starknetrpc.EventFilter) (StarknetBatchBuilder, error)
	// RequestTxReceiptByHash(h *felt.Felt) (StarknetBatchBuilder, error)
	Build() []gethrpc.BatchElem
}

var _ StarknetBatchBuilder = (*batchBuilder)(nil)

type batchBuilder struct {
	args []gethrpc.BatchElem
}

func NewBatchBuilder() StarknetBatchBuilder {
	return &batchBuilder{
		args: make([]gethrpc.BatchElem, 0),
	}
}

func (b *batchBuilder) RequestBlockByHash(h *felt.Felt) (StarknetBatchBuilder, error) {
	b.args = append(b.args, gethrpc.BatchElem{
		Method: "starknet_getBlockWithTxs",
		Args: []interface{}{
			starknetrpc.BlockID{Hash: h},
		},
		Result: &FinalizedBlock{},
	})
	return b, nil
}

func (b *batchBuilder) Build() []gethrpc.BatchElem {
	return b.args
}

type StarknetChainClient interface {
	// only finalized blocks have a block hashes
	BlockByHash(ctx context.Context, h *felt.Felt) (FinalizedBlock, error)
	// only finalized blocks have numbers
	BlockByNumber(ctx context.Context, id uint64) (FinalizedBlock, error)
	// only way to get the latest pending block (only 1 pending block exists at a time)
	LatestPendingBlock(ctx context.Context) (starknetrpc.PendingBlock, error)
	// returns block number and block has of latest finalized block
	LatestBlockHashAndNumber(ctx context.Context) (starknetrpc.BlockHashAndNumberOutput, error)
	// get block logs, event logs, etc.
	EventsByFilter(ctx context.Context, f starknetrpc.EventFilter) ([]starknetrpc.EmittedEvent, error)
	TxReceiptByHash(ctx context.Context, h *felt.Felt) (starknetrpc.TransactionReceipt, error)
	Batch(builder StarknetBatchBuilder) ([]gethrpc.BatchElem, error)
}

func (c *Client) BlockByHash(ctx context.Context, h *felt.Felt) (FinalizedBlock, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	block, err := c.Provider.BlockWithTxs(ctx, starknetrpc.BlockID{Hash: h})

	if err != nil {
		return FinalizedBlock{}, fmt.Errorf("error in BlockByHash: %w", err)
	}

	finalizedBlock, ok := block.(*FinalizedBlock)

	if !ok {
		return FinalizedBlock{}, fmt.Errorf("expected type Finalized block but found: %T", block)
	}

	return *finalizedBlock, nil
}

func (c *Client) Batch(ctx context.Context, builder StarknetBatchBuilder) ([]gethrpc.BatchElem, error) {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	args := builder.Build()

	err := c.EthClient.BatchCallContext(ctx, args)

	if err != nil {
		return nil, fmt.Errorf("error in Batch: %w", err)
	}

	return args, nil
}
