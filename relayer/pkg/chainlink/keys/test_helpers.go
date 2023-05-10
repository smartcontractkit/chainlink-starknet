package keys

import (
	"encoding/hex"
	"math/big"
)

var (
	// devnet key derivation
	// https://github.com/Shard-Labs/starknet-devnet/blob/master/starknet_devnet/account.py
	DevnetClassHash, _ = new(big.Int).SetString("1803505466663265559571280894381905521939782500874858933595227108099796801620", 10)
	DevnetSalt         = big.NewInt(20)
)

func (key Key) DevnetAccountAddrStr() string {
	return "0x" + hex.EncodeToString(PubKeyToAccount(key.pub, DevnetClassHash, DevnetSalt))

}