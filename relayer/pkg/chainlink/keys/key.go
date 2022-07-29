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
	return key.AccountAddressStr()
}

// this is the derived contract address, the contract is deployed using the StarkKeyStr
// This is the primary identifier for onchain interactions
// the private key is identified by this
func (key Key) AccountAddressStr() string {
	return "0x" + hex.EncodeToString(PubKeyToAccount(key.privkey.PublicKey, defaultContractHash, defaultSalt))
}

// StarkKeyStr is the starknet public key associated to the private key
// it is the X component of the ECDSA pubkey and used in the deployment of the account contract
// this func is used in exporting it via CLI and API
func (key Key) StarkKeyStr() string {
	return "0x" + hex.EncodeToString(PubKeyToStarkKey(key.privkey.PublicKey))
}

// Raw from private key
func (key Key) Raw() Raw {
	return key.privkey.D.Bytes()
}

// String is the print-friendly format of the Key
func (key Key) String() string {
	return fmt.Sprintf("StarkNetKey{PrivateKey: <redacted>, Contract Address: %s}", key.AccountAddressStr())
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
