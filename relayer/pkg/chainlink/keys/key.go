package keys

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/caigo"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

// Raw represents the Stark private key
type Raw []byte

// Key gets the Key
func (raw Raw) Key() Key {
	k := Key{}
	var err error

	k.priv = new(big.Int).SetBytes(raw)
	k.pub.X, k.pub.Y, err = caigo.Curve.PrivateToPoint(k.priv)
	if err != nil {
		panic(err) // key not generated
	}
	return k
}

// String returns description
func (raw Raw) String() string {
	return "<Starknet Raw Private Key>"
}

// GoString wraps String()
func (raw Raw) GoString() string {
	return raw.String()
}

var _ fmt.GoStringer = &Key{}

type PublicKey struct {
	X, Y *big.Int
}

// Key represents Starknet key
type Key struct {
	priv *big.Int
	pub  PublicKey
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
	return GenerateKey(reader)
}

// ID gets Key ID
func (key Key) ID() string {
	return key.StarkKeyStr()
}

// StarkKeyStr is the starknet public key associated to the private key
// it is the X component of the ECDSA pubkey and used in the deployment of the account contract
// this func is used in exporting it via CLI and API
func (key Key) StarkKeyStr() string {
	return "0x" + hex.EncodeToString(PubKeyToStarkKey(key.pub))
}

// Raw from private key
func (key Key) Raw() Raw {
	return key.priv.Bytes()
}

// String is the print-friendly format of the Key
func (key Key) String() string {
	return fmt.Sprintf("StarknetKey{PrivateKey: <redacted>, StarkKey: %s}", key.StarkKeyStr())
}

// GoString wraps String()
func (key Key) GoString() string {
	return key.String()
}

// ToPrivKey returns the key usable for signing.
func (key Key) ToPrivKey() *big.Int {
	return key.priv
}

// PublicKey copies public key object
func (key Key) PublicKey() PublicKey {
	return key.pub
}

// Sign creates a signature by concat'ing the public key, and the curve parameters r,s.
//
//	public key (32 bytes) + r (32 bytes) + s (32 bytes)
func Sign(hash *big.Int, key Key) ([]byte, error) {

	r, s, err := caigo.Curve.Sign(hash, key.ToPrivKey())
	if err != nil {
		return []byte{}, err
	}

	// enforce s <= N/2 to prevent signature malleability
	if s.Cmp(new(big.Int).Rsh(caigo.Curve.N, 1)) > 0 {
		s.Sub(caigo.Curve.N, s)
	}

	// encoding: public key (32 bytes) + r (32 bytes) + s (32 bytes)
	expectedLen := 3 * byteLen
	buff := bytes.NewBuffer([]byte(PubKeyToStarkKey(key.pub)))
	if _, err := buff.Write(starknet.PadBytes(r.Bytes(), byteLen)); err != nil {
		return []byte{}, err
	}
	if _, err := buff.Write(starknet.PadBytes(s.Bytes(), byteLen)); err != nil {
		return []byte{}, err
	}

	out := buff.Bytes()
	if len(out) != expectedLen {
		return []byte{}, errors.Errorf("unexpected signature size, got %d want %d", len(out), expectedLen)
	}
	return out, nil
}
