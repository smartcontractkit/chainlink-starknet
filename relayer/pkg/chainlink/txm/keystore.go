package txm

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/caigo"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
)

// LooppKeystore implements [loop.Keystore] interface and the requirements
// of signature d/encoding of the [KeystoreAdapter]
type LooppKeystore struct {
	Get func(id string) (*big.Int, error)
}

func NewLooppKeystore(get func(id string) (*big.Int, error)) *LooppKeystore {
	return &LooppKeystore{
		Get: get,
	}
}

var _ loop.Keystore = &LooppKeystore{}

// Sign implements [loop.Keystore]
// hash is expected to be the byte representation of big.Int
// the return []byte is encodes a starknet signature per [signature.bytes]
func (lk *LooppKeystore) Sign(ctx context.Context, id string, hash []byte) ([]byte, error) {

	k, err := lk.Get(id)
	if err != nil {
		return nil, err
	}
	// loopp spec requires passing nil hash to check existence of id
	if hash == nil {
		return nil, nil
	}

	starkHash := new(big.Int).SetBytes(hash)
	x, y, err := caigo.Curve.Sign(starkHash, k)
	if err != nil {
		return nil, fmt.Errorf("error signing data with curve: %w", err)
	}

	s := &signature{
		x: x,
		y: y,
	}
	return s.bytes()
}

// TODO what is this supposed to return for starknet?
func (lk *LooppKeystore) Accounts(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

const (
	maxPointByteLen = 32 // stark curve max is 252 bits
	signatureLen    = 2 * maxPointByteLen
)

// signature is an intermediate representation for translating between a raw-bytes signature and a caigo
// signature comprised of big.Ints
type signature struct {
	x, y *big.Int
}

// encodes x,y into []byte slice
// the first [maxPointByteLen] are the padded bytes of x
// the second [maxPointByteLen] are the padded bytes of y
func (s *signature) bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	n, err := buf.Write(padBytes(s.x.Bytes(), maxPointByteLen))
	if err != nil {
		return nil, fmt.Errorf("error writing 'x' component of signature: %w", err)
	}
	if n != maxPointByteLen {
		return nil, fmt.Errorf("unexpected write length of 'x' component of signature: wrote %d expected %d", n, maxPointByteLen)
	}

	n, err = buf.Write(padBytes(s.y.Bytes(), maxPointByteLen))
	if err != nil {
		return nil, fmt.Errorf("error writing 'y' component of signature: %w", err)
	}
	if n != maxPointByteLen {
		return nil, fmt.Errorf("unexpected write length of 'y' component of signature: wrote %d expected %d", n, maxPointByteLen)
	}

	if buf.Len() != signatureLen {
		return nil, fmt.Errorf("error in signature length")
	}
	return buf.Bytes(), nil
}

// b is expected to encode x,y components in accordance with [signature.bytes]
func signatureFromBytes(b []byte) (*signature, error) {
	if len(b) != signatureLen {
		return nil, fmt.Errorf("expected signature length %d got %d", signatureLen, len(b))
	}

	x := b[:maxPointByteLen]
	y := b[maxPointByteLen:]

	return &signature{
		x: new(big.Int).SetBytes(x),
		y: new(big.Int).SetBytes(y),
	}, nil
}

// pad bytes to specific length
func padBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}

// KeystoreAdapter is a starknet-specific adaption layer to translate between the generic Loop Keystore (bytes) and
// the type specific caigo Keystore (big.Int)
// The loop.Keystore must be produce a byte representation that can be parsed by the Decode func implementation
// Users of the interface are responsible to ensure this compatibility.
type KeystoreAdapter interface {
	caigo.Keystore
	// Loopp must return a LOOPp Keystore implementation whose Sign func
	// is compatible with the [Decode] func implementation
	Loopp() loop.Keystore
	// Decode translates from the raw signature of the LOOPp Keystore to that of the Caigo Keystore
	Decode(ctx context.Context, rawSignature []byte) (*big.Int, *big.Int, error)
}

// keystoreAdapter implements [KeystoreAdapter]
type keystoreAdapter struct {
	looppKs loop.Keystore
}

// NewKeystoreAdapter instantiates the KeystoreAdapter interface
// Callers are responsible for ensuring that the given LOOPp Keystore encodes
// signatures that can be parsed by the Decode function
func NewKeystoreAdapter(lk loop.Keystore) KeystoreAdapter {
	return &keystoreAdapter{looppKs: lk}
}

// Sign implements the caigo Keystore Sign func.
func (ca *keystoreAdapter) Sign(ctx context.Context, senderAddress string, hash *big.Int) (*big.Int, *big.Int, error) {
	raw, err := ca.looppKs.Sign(ctx, senderAddress, hash.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("error computing loopp keystore signature: %w", err)
	}
	return ca.Decode(ctx, raw)
}

func (ca *keystoreAdapter) Decode(ctx context.Context, rawSignature []byte) (x *big.Int, y *big.Int, err error) {

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		starknetSig, serr := signatureFromBytes(rawSignature)
		if serr != nil {
			return nil, nil, fmt.Errorf("error creating starknet signature from raw signature: %w", serr)
		}
		x, y = starknetSig.x, starknetSig.y
	}
	return x, y, err
}

func (ca *keystoreAdapter) Loopp() loop.Keystore {
	return ca.looppKs
}
