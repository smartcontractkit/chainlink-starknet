package keys

import "math/big"

func (key *Key) Set(hash, salt *big.Int) {
	key.hash = hash
	key.salt = salt
}