package keys

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	starksig "github.com/NethermindEth/juno/pkg/crypto/signature"
<<<<<<< HEAD
	"github.com/NethermindEth/juno/pkg/crypto/weierstrass"
)

var (
	curve   = weierstrass.Stark()
	byteLen = 32

	// note: the contract hash must match the corresponding OZ gauntlet command hash - otherwise addresses will not correspond
	defaultContractHash, _ = new(big.Int).SetString("0x726edb35cc732c1b3661fd837592033bd85ae8dde31533c35711fb0422d8993", 0)
	defaultSalt            = big.NewInt(100)
=======
>>>>>>> rename keys + remove redundant funcs
)

// Raw represents the Stark private key
type StarkRaw []byte

// Key gets the Key
func (raw StarkRaw) Key() StarkKey {
	privKey := starksig.PrivateKey{}
	privKey.D = new(big.Int).SetBytes(raw)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(raw)

	return StarkKey{
		privkey: privKey,
	}
}

// String returns description
func (raw StarkRaw) String() string {
	return "<StarkNet Raw Private Key>"
}

// GoString wraps String()
func (raw StarkRaw) GoString() string {
	return raw.String()
}

var _ fmt.GoStringer = &StarkKey{}

// Key represents StarkNet key
type StarkKey struct {
	privkey starksig.PrivateKey
}

// New creates new Key
func New() (StarkKey, error) {
	return newFrom(crypto_rand.Reader)
}

// MustNewInsecure return Key if no error
func MustNewInsecure(reader io.Reader) StarkKey {
	key, err := newFrom(reader)
	if err != nil {
		panic(err)
	}
	return key
}

func newFrom(reader io.Reader) (StarkKey, error) {
	privKey, err := starksig.GenerateKey(curve, reader)
	if err != nil {
		return StarkKey{}, err
	}
	return StarkKey{
		privkey: *privKey,
	}, nil
}

// ID gets Key ID
func (key StarkKey) ID() string {
	return key.PublicKeyStr()
}

// PublicKeyStr
<<<<<<< HEAD
// Not actually public key, this is the derived contract address
func (key Key) PublicKeyStr() string {
	return "0x" + hex.EncodeToString(PubToStarkKey(key.privkey.PublicKey, defaultContractHash, defaultSalt))
}

// PubToStarkKey implements the pubkey to deployed account given contract hash + salt
func PubToStarkKey(pubkey starksig.PublicKey, classHash, salt *big.Int) []byte {
	hash := pedersen.ArrayDigest(
		new(big.Int).SetBytes([]byte("STARKNET_CONTRACT_ADDRESS")),
		big.NewInt(0),
		salt,      // salt
		classHash, // classHash
		pedersen.ArrayDigest(pubkey.X),
	)

	// pad big.Int to 32 bytes if needed
	if len(hash.Bytes()) < byteLen {
		out := make([]byte, byteLen)
		return hash.FillBytes(out)
	}

	return hash.Bytes()
=======
func (key StarkKey) PublicKeyStr() string {

	// TODO: what if this is different per network?
	// https://github.com/Shard-Labs/starknet-devnet/blob/master/starknet_devnet/account.py
	classHash, _ := new(big.Int).SetString("1803505466663265559571280894381905521939782500874858933595227108099796801620", 10)
	salt := big.NewInt(20)

	return "0x" + hex.EncodeToString(PubKeyToContract(key.privkey.PublicKey, classHash, salt))
>>>>>>> rename keys + remove redundant funcs
}

// Raw from private key
func (key StarkKey) Raw() StarkRaw {
	return key.privkey.D.Bytes()
}

// String is the print-friendly format of the Key
func (key StarkKey) String() string {
	return fmt.Sprintf("StarkNetKey{PrivateKey: <redacted>, Public Key: %s}", key.PublicKeyStr())
}

// GoString wraps String()
func (key StarkKey) GoString() string {
	return key.String()
}

// ToPrivKey returns the key usable for signing.
func (key StarkKey) ToPrivKey() starksig.PrivateKey {
	return key.privkey
}

// PublicKey copies public key object
func (key StarkKey) PublicKey() starksig.PublicKey {
	return key.privkey.PublicKey
}
