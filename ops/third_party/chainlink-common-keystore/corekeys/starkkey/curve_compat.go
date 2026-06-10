package starkkey

import "math/big"

// STARK curve subgroup order.
var curveOrder = func() *big.Int {
	n, ok := new(big.Int).SetString("3618502788666131213697322783095070105526743751716087489154079457884512865583", 10)
	if !ok {
		panic("invalid stark curve order")
	}
	return n
}()
