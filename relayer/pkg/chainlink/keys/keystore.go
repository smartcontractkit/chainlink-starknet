package keys

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/caigo"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	starknetadapter "github.com/smartcontractkit/chainlink-relay/pkg/loop/adapters/starknet"
)

// KeystoreAdapter is a starknet-specific adaption layer to translate between the generic Loop Keystore (bytes) and
// the type specific caigo Keystore (big.Int)
// The loop.Keystore must be produce a byte representation that can be parsed by the Decode func implementation
// Users of the interface are responsible to ensure this compatibility.
//
//go:generate mockery --name KeystoreAdapter --output ./mocks/ --case=underscore --filename keystore.go
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
		starknetSig, serr := starknetadapter.SignatureFromBytes(rawSignature)
		if serr != nil {
			return nil, nil, fmt.Errorf("error creating starknet signature from raw signature: %w", serr)
		}
		x, y, err = starknetSig.Ints()
	}
	return x, y, err
}

func (ca *keystoreAdapter) Loopp() loop.Keystore {
	return ca.looppKs
}

type KeyGetter interface {
	Get(id string) (*big.Int, error)
}

// LooppKeystore implements [loop.Keystore] interface and the requirements
// of signature d/encoding of the [KeystoreAdapter]
type LooppKeystore struct {
	KeyGetter
}

func NewLooppKeystore(getter KeyGetter) *LooppKeystore {
	return &LooppKeystore{
		KeyGetter: getter,
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

	s, err := starknetadapter.SignatureFromBigInts(x, y)
	if err != nil {
		return nil, fmt.Errorf("error creating signature from big ints: %w", err)
	}
	return s.Bytes()
}

// TODO what is this supposed to return for starknet?
func (lk *LooppKeystore) Accounts(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}
