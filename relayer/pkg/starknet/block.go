package starknet

import starknetrpc "github.com/NethermindEth/starknet.go/rpc"

// BlockTagLatest is the RPC 0.9+ block tag for the latest block finalized by L2
// consensus. Used for read-only contract calls (monitoring, OCR2 cache reads).
const BlockTagLatest = "latest"

// BlockTagPreConfirmed is the RPC 0.9+ block tag for the block currently being
// built (height latest + 1). Replaces the deprecated "pending" tag (Starknet v0.14.x).
// Used for TXM nonce and fee estimation where in-flight state matters.
const BlockTagPreConfirmed = "pre_confirmed"

// LatestBlockID returns a block ID for the latest finalized block tag.
func LatestBlockID() starknetrpc.BlockID {
	return starknetrpc.WithBlockTag(BlockTagLatest)
}

// PreConfirmedBlockID returns a block ID for the pre_confirmed block tag.
func PreConfirmedBlockID() starknetrpc.BlockID {
	return starknetrpc.WithBlockTag(BlockTagPreConfirmed)
}
