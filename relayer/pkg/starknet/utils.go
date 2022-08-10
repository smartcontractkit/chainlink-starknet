package starknet

import "math/big"

func BigIntPadBytes(a *big.Int, length int) []byte {
	return PadBytes(a.Bytes(), length)
}

func PadBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	return a
}
