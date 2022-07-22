package keys

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	starksig "github.com/NethermindEth/juno/pkg/crypto/signature"
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

// Not actually public key, this is the derived contract address
func (key StarkKey) PublicKeyStr() string {
	return "0x" + hex.EncodeToString(PubKeyToContract(key.privkey.PublicKey, defaultContractHash, defaultSalt))
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
