package keys

import (
	"math/big"

	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	starksig "github.com/NethermindEth/juno/pkg/crypto/signature"
	"github.com/NethermindEth/juno/pkg/crypto/weierstrass"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

// constants
var (
	curve   = weierstrass.Stark()
	byteLen = 32

	// note: the contract hash must match the corresponding OZ gauntlet command hash - otherwise addresses will not correspond
	defaultContractHash, _ = new(big.Int).SetString("0x726edb35cc732c1b3661fd837592033bd85ae8dde31533c35711fb0422d8993", 0)
	defaultSalt            = big.NewInt(100)
)

// PubKeyToContract implements the pubkey to deployed account given contract hash + salt
func PubKeyToAccount(pubkey starksig.PublicKey, classHash, salt *big.Int) []byte {
	hash := pedersen.ArrayDigest(
		new(big.Int).SetBytes([]byte("STARKNET_CONTRACT_ADDRESS")),
		big.NewInt(0),
		salt,      // salt
		classHash, // classHash
		pedersen.ArrayDigest(pubkey.X),
	)

	// pad big.Int to 32 bytes if needed
	return starknet.BigIntPadBytes(hash, byteLen)
}

// PubToStarkKey implements the pubkey to starkkey functionality: https://github.com/0xs34n/starknet.js/blob/cd61356974d355aa42f07a3d63f7ccefecbd913c/src/utils/ellipticCurve.ts#L49
func PubKeyToStarkKey(pubkey starksig.PublicKey) []byte {
	return starknet.BigIntPadBytes(pubkey.X, byteLen)
}
