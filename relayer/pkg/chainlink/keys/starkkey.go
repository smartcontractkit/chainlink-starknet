package keys

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	starksig "github.com/NethermindEth/juno/pkg/crypto/signature"
	"github.com/NethermindEth/juno/pkg/crypto/weierstrass"
)

var (
	curve   = weierstrass.Stark()
	byteLen = 32

	// note: the contract hash must match the corresponding OZ gauntlet command hash - otherwise addresses will not correspond
	defaultContractHash, _ = new(big.Int).SetString("0x726edb35cc732c1b3661fd837592033bd85ae8dde31533c35711fb0422d8993", 0)
	defaultSalt            = big.NewInt(100)
)

// Raw represents the Stark private key
type Raw []byte

// Key gets the Key
func (raw Raw) Key() Key {
	privKey := starksig.PrivateKey{}
	privKey.D = new(big.Int).SetBytes(raw)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(raw)

	return Key{
		privkey: privKey,
	}
}

// String returns description
func (raw Raw) String() string {
	return "<StarkNet Raw Private Key>"
}

// GoString wraps String()
func (raw Raw) GoString() string {
	return raw.String()
}

var _ fmt.GoStringer = &Key{}

// Key represents StarkNet key
type Key struct {
	privkey starksig.PrivateKey
}

// New creates new Key
func New() (Key, error) {
	return newFrom(crypto_rand.Reader)
}

// MustNewInsecure return Key if no error
func MustNewInsecure(reader io.Reader) Key {
	key, err := newFrom(reader)
	if err != nil {
		panic(err)
	}
	return key
}

func newFrom(reader io.Reader) (Key, error) {
	privKey, err := starksig.GenerateKey(curve, reader)
	if err != nil {
		return Key{}, err
	}
	return Key{
		privkey: *privKey,
	}, nil
}

// ID gets Key ID
func (key Key) ID() string {
	return key.PublicKeyStr()
}

// PublicKeyStr
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
}

// Raw from private key
func (key Key) Raw() Raw {
	return key.privkey.D.Bytes()
}

// String is the print-friendly format of the Key
func (key Key) String() string {
	return fmt.Sprintf("StarkNetKey{PrivateKey: <redacted>, Public Key: %s}", key.PublicKeyStr())
}

// GoString wraps String()
func (key Key) GoString() string {
	return key.String()
}

// ToPrivKey returns the key usable for signing.
func (key Key) ToPrivKey() starksig.PrivateKey {
	return key.privkey
}

// PublicKey copies public key object
func (key Key) PublicKey() starksig.PublicKey {
	return key.privkey.PublicKey
}
