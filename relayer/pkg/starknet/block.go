package starknet

import starknetrpc "github.com/NethermindEth/starknet.go/rpc"

// LatestBlockID returns a block ID for the latest finalized block tag (RPC 0.9+).
// Used for read-only contract calls (monitoring, OCR2 cache reads).
func LatestBlockID() starknetrpc.BlockID {
	return starknetrpc.WithBlockTag(starknetrpc.BlockTagLatest)
}

// PreConfirmedBlockID returns a block ID for the pre_confirmed block tag (RPC 0.9+).
// Replaces the deprecated "pending" tag. Used for TXM nonce and fee estimation.
func PreConfirmedBlockID() starknetrpc.BlockID {
	return starknetrpc.WithBlockTag(starknetrpc.BlockTagPreConfirmed)
}
