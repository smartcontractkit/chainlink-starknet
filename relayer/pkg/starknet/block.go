package starknet

import starknetrpc "github.com/NethermindEth/starknet.go/rpc"

// BlockTagPreConfirmed is the RPC 0.9+ block tag for the block currently being
// built (height latest + 1). Replaces the deprecated "pending" tag (Starknet v0.14.x).
const BlockTagPreConfirmed = "pre_confirmed"

// PreConfirmedBlockID returns a block ID for the pre_confirmed block tag.
func PreConfirmedBlockID() starknetrpc.BlockID {
	return starknetrpc.WithBlockTag(BlockTagPreConfirmed)
}
