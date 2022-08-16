package starknet

import (
	"math/big"
)

// convert big into padded bytes
func PadBytesBigInt(a *big.Int, length int) []byte {
	return PadBytes(a.Bytes(), length)
}

// padd bytes to specific length
func PadBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}

// convert 32 byte to "0" + 31 bytes
func EnsureFelt(b [32]byte) (out []byte) {
	out = make([]byte, 32)
	copy(out[:], b[:])
	out[0] = 0
	return out
}