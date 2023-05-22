package keys

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/smartcontractkit/caigo"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

//go:generate mockery --name KeystoreAdapter --output ./mocks/ --case=underscore --filename keystore.go

// KeystoreAdapter is a starknet-specific adaption layer to translate between the generic Loop Keystore (bytes) and
// the type specific caigo Keystore (big.Int)
// The loop.Keystore must be produce a byte representation that can be parsed by the caigo.Keystore implementation
// Users of the interface are responsible to ensure this compatibility.
type KeystoreAdapter interface {
	caigo.Keystore
	Loopp() loop.Keystore
}

// caigoAdapter implements [KeystoreAdapter]
type caigoAdapter struct {
	looppKs loop.Keystore
}

// Sign implements the caigo Keystore Sign func.
func (ca *caigoAdapter) Sign(senderAddress string, hash *big.Int) (*big.Int, *big.Int, error) {
	raw, err := ca.looppKs.Sign(context.Background(), senderAddress, hash.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("error computing loopp keystore signature: %w", err)
	}
	starknetSign, err := signatureFromBytes(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating starknet signature from raw signature: %w", err)
	}
	return starknetSign.x, starknetSign.y, nil
}

func (ca *caigoAdapter) Loopp() loop.Keystore {
	return ca.looppKs
}

func NewCaigoAdapter(looppKs loop.Keystore) KeystoreAdapter {
	return &caigoAdapter{looppKs: looppKs}
}

type NonceManager interface {
	types.Service

	Register(ctx context.Context, address caigotypes.Hash, chainId string, client NonceManagerClient) error

	NextSequence(address caigotypes.Hash, chainID string) (*big.Int, error)
	IncrementNextSequence(address caigotypes.Hash, chainID string, currentNonce *big.Int) error
}

const (
	maxPointByteLen = 32 // stark curve max is 252 bits
	signatureLen    = 2 * maxPointByteLen
)

// MemKeystore is an in-memory implementation of the LOOPp Keystore interface
type MemKeystore struct {
	mu   sync.Mutex
	keys map[string]*big.Int
}

func NewMemKeystore() *MemKeystore {
	return &MemKeystore{
		keys: make(map[string]*big.Int),
	}
}

func (ks *MemKeystore) Put(senderAddress string, k *big.Int) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.keys[senderAddress] = k
}

var ErrSenderNoExist = errors.New("sender does not exist")

func (ks *MemKeystore) Get(senderAddress string) (*big.Int, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, exists := ks.keys[senderAddress]
	if !exists {
		return nil, fmt.Errorf("error getting key for sender %s: %w", senderAddress, ErrSenderNoExist)
	}
	return k, nil
}

// Sign implements the LoopKeystore interface
// this implementation wraps starknet specific curve and expects
// hash: byte representation (big-endian) of *big.Int.
func (ks *MemKeystore) Sign(id string, hash []byte) ([]byte, error) {
	k, err := ks.Get(id)
	if err != nil {
		return nil, err
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

// signature is an intermediate representation for translating between a raw-bytes signature and a caigo
// signature comprised of big.Ints
type signature struct {
	x, y *big.Int
}

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
		return nil, fmt.Errorf("expected slice len %d got %d", signatureLen, len(b))
	}
	x := b[:maxPointByteLen]
	y := b[maxPointByteLen:]

	return &signature{
		x: new(big.Int).SetBytes(x),
		y: new(big.Int).SetBytes(y),
	}, nil
}

// pad bytes  to specific length
func padBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}
