package starknet

import "math/big"

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
