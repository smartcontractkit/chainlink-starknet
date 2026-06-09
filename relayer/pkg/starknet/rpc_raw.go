package starknet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
)

// blockTagParam converts a BlockID to a JSON-RPC block parameter. starknet.go
// v0.9.0 only marshals "pending" and "latest"; pre_confirmed requires a raw call.
func blockTagParam(blockID starknetrpc.BlockID) (interface{}, error) {
	switch blockID.Tag {
	case BlockTagPreConfirmed, BlockTagLatest, BlockTagPending:
		return blockID.Tag, nil
	case "":
		if blockID.Number != nil {
			return map[string]uint64{"block_number": *blockID.Number}, nil
		}
		if blockID.Hash != nil && blockID.Hash.BigInt(big.NewInt(0)).BitLen() != 0 {
			return map[string]string{"block_hash": blockID.Hash.String()}, nil
		}
	}
	return nil, fmt.Errorf("unsupported block id: %+v", blockID)
}

func isPreConfirmedBlock(blockID starknetrpc.BlockID) bool {
	return blockID.Tag == BlockTagPreConfirmed
}

// isUnsupportedPreConfirmedBlockTagErr reports whether an RPC node rejected
// pre_confirmed (legacy RPC 0.8 / starknet-devnet-rs still expect "pending").
func isUnsupportedPreConfirmedBlockTagErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "pre_confirmed") &&
		(strings.Contains(msg, "unknown variant") || strings.Contains(msg, "Invalid block ID"))
}

func (c *Client) rawRPC(ctx context.Context, method string, params []interface{}, result interface{}) error {
	if c.defaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.defaultTimeout)
		defer cancel()
	}

	if err := c.EthClient.CallContext(ctx, result, method, params...); err != nil {
		return fmt.Errorf("error in rawRPC %s: %w", method, err)
	}
	return nil
}

func hexStringsToFelts(hexResults []string) ([]*felt.Felt, error) {
	felts := make([]*felt.Felt, len(hexResults))
	for i, hexVal := range hexResults {
		f, err := starknetutils.HexToFelt(hexVal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse felt %q: %w", hexVal, err)
		}
		felts[i] = f
	}
	return felts, nil
}

func (c *Client) callAtBlock(ctx context.Context, call starknetrpc.FunctionCall, blockID starknetrpc.BlockID) ([]*felt.Felt, error) {
	blockParam, err := blockTagParam(blockID)
	if err != nil {
		return nil, err
	}

	var hexResults []string
	params := []interface{}{call, blockParam}
	if err := c.rawRPC(ctx, "starknet_call", params, &hexResults); err != nil {
		return nil, err
	}
	return hexStringsToFelts(hexResults)
}

func (c *Client) nonceAtBlock(ctx context.Context, blockID starknetrpc.BlockID, accountAddress *felt.Felt) (*felt.Felt, error) {
	blockParam, err := blockTagParam(blockID)
	if err != nil {
		return nil, err
	}

	var nonceHex string
	params := []interface{}{blockParam, accountAddress}
	if err := c.rawRPC(ctx, "starknet_getNonce", params, &nonceHex); err != nil {
		return nil, err
	}
	return starknetutils.HexToFelt(nonceHex)
}

func (c *Client) nonceAtPreConfirmed(ctx context.Context, accountAddress *felt.Felt) (*felt.Felt, error) {
	nonce, err := c.nonceAtBlock(ctx, PreConfirmedBlockID(), accountAddress)
	if err != nil && isUnsupportedPreConfirmedBlockTagErr(err) {
		return c.nonceAtBlock(ctx, starknetrpc.WithBlockTag(BlockTagPending), accountAddress)
	}
	return nonce, err
}

func (c *Client) eventsAtBlock(ctx context.Context, input starknetrpc.EventsInput) (*starknetrpc.EventChunk, error) {
	chunk, err := c.eventsAtBlockOnce(ctx, input)
	if err != nil && isUnsupportedPreConfirmedBlockTagErr(err) && eventsInputUsesPreConfirmed(input) {
		fallback := input
		if isPreConfirmedBlock(fallback.FromBlock) {
			fallback.FromBlock = starknetrpc.WithBlockTag(BlockTagPending)
		}
		if isPreConfirmedBlock(fallback.ToBlock) {
			fallback.ToBlock = starknetrpc.WithBlockTag(BlockTagPending)
		}
		return c.eventsAtBlockOnce(ctx, fallback)
	}
	return chunk, err
}

func (c *Client) eventsAtBlockOnce(ctx context.Context, input starknetrpc.EventsInput) (*starknetrpc.EventChunk, error) {
	fromParam, err := blockTagParam(input.FromBlock)
	if err != nil {
		return nil, err
	}
	toParam, err := blockTagParam(input.ToBlock)
	if err != nil {
		return nil, err
	}

	rawInput := map[string]interface{}{
		"event_filter": map[string]interface{}{
			"from_block": fromParam,
			"to_block":   toParam,
			"address":    input.Address.String(),
			"keys":       input.Keys,
		},
		"result_page_request": map[string]interface{}{
			"chunk_size": input.ChunkSize,
		},
	}
	if input.ContinuationToken != "" {
		rawInput["result_page_request"].(map[string]interface{})["continuation_token"] = input.ContinuationToken
	}

	var chunk starknetrpc.EventChunk
	if err := c.rawRPC(ctx, "starknet_getEvents", []interface{}{rawInput}, &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (c *Client) estimateFeeAtBlock(
	ctx context.Context,
	txns []starknetrpc.BroadcastTxn,
	flags []starknetrpc.SimulationFlag,
	blockTag string,
) ([]starknetrpc.FeeEstimation, error) {
	params := []interface{}{txns, flags, blockTag}
	var estimates []starknetrpc.FeeEstimation
	if err := c.rawRPC(ctx, "starknet_estimateFee", params, &estimates); err != nil {
		return nil, err
	}
	return estimates, nil
}

// EstimateFeeAtPreConfirmed estimates fees against the pre_confirmed block state.
func (c *Client) EstimateFeeAtPreConfirmed(
	ctx context.Context,
	txns []starknetrpc.BroadcastTxn,
	flags []starknetrpc.SimulationFlag,
) ([]starknetrpc.FeeEstimation, error) {
	estimates, err := c.estimateFeeAtBlock(ctx, txns, flags, BlockTagPreConfirmed)
	if err != nil && isUnsupportedPreConfirmedBlockTagErr(err) {
		return c.estimateFeeAtBlock(ctx, txns, flags, BlockTagPending)
	}
	return estimates, err
}

// eventsInputUsesPreConfirmed reports whether an events query targets pre_confirmed.
func eventsInputUsesPreConfirmed(input starknetrpc.EventsInput) bool {
	return isPreConfirmedBlock(input.FromBlock) || isPreConfirmedBlock(input.ToBlock)
}

// blockIDFromJSON is used in tests to verify pre_confirmed marshaling.
func blockIDFromJSON(tag string) (json.RawMessage, error) {
	return json.Marshal(tag)
}
