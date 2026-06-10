package keystore

import (
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/curve25519"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/serialization"
)

type KeyPath []string

func (k KeyPath) String() string {
	return joinKeySegments(k...)
}

func (k KeyPath) Base() string {
	return k[len(k)-1]
}

// HasPrefix returns true if k starts with the given prefix path.
// For example, KeyPath{"solana", "tx", "my-key"}.HasPrefix(KeyPath{"solana", "tx"}) returns true.
func (k KeyPath) HasPrefix(prefix KeyPath) bool {
	if len(prefix) > len(k) {
		return false
	}
	for i := range prefix {
		if k[i] != prefix[i] {
			return false
		}
	}
	return true
}

func NewKeyPath(segments ...string) KeyPath {
	return segments
}

func NewKeyPathFromString(fullName string) KeyPath {
	return strings.Split(fullName, "/")
}

// joinKeySegments joins path-like key name segments using "/" and avoids double slashes.
// Empty segments are skipped so joinKeySegments("evm", "tx", "my-key") => "evm/tx/my-key".
func joinKeySegments(segments ...string) string {
	cleaned := make([]string, 0, len(segments))
	for _, s := range segments {
		s = strings.Trim(s, "/")
		if s == "" {
			continue
		}
		cleaned = append(cleaned, s)
	}
	return strings.Join(cleaned, "/")
}

type KeyType string

func (k KeyType) String() string {
	return string(k)
}

func (k KeyType) IsEncryptionKeyType() bool {
	return slices.Contains(AllEncryptionKeyTypes, k)
}

func (k KeyType) IsDigitalSignatureKeyType() bool {
	return slices.Contains(AllDigitalSignatureKeyTypes, k)
}

const (
	// Hybrid encryption (key exchange + encryption) key types.
	// Naming schema is generally <key exchange algorithm><encryption algorithm>.
	// Except for widely used/commonly paired encryption algorithms, we
	// omit the encryption algorithm. So for example X25519 with ChaCha20Poly1305
	// (via box) is specified just as X25519.
	// X25519:
	// - X25519 for ECDH key exchange.
	// - Box for encryption (ChaCha20Poly1305)
	X25519 KeyType = "X25519"
	// ECDH_P256:
	// - ECDH on P-256
	// - Encryption with AES-GCM and HKDF-SHA256
	ECDH_P256 KeyType = "ECDH_P256"

	// Digital signature key types.
	// Ed25519:
	// - Ed25519 for digital signatures.
	// - Supports arbitrary messages sizes, no hashing required.
	Ed25519 KeyType = "Ed25519"
	// ECDSA_S256:
	// - ECDSA on secp256k1 for digital signatures.
	// - Only signs 32 byte digests. Caller must hash the data before signing.
	ECDSA_S256 KeyType = "ECDSA_S256"
)

type KeyTypeList []KeyType

func (k KeyTypeList) String() string {
	types := make([]string, 0, len(k))
	for _, k := range k {
		types = append(types, k.String())
	}
	return strings.Join(types, ", ")
}

var AllKeyTypes = KeyTypeList{X25519, ECDH_P256, Ed25519, ECDSA_S256}
var AllEncryptionKeyTypes = KeyTypeList{X25519, ECDH_P256}
var AllDigitalSignatureKeyTypes = KeyTypeList{Ed25519, ECDSA_S256}

type ScryptParams struct {
	N int
	P int
}

var (
	DefaultScryptParams = ScryptParams{
		N: gethkeystore.StandardScryptN,
		P: gethkeystore.StandardScryptP,
	}
	FastScryptParams ScryptParams = ScryptParams{
		N: 1 << 14,
		P: 1,
	}
)

// KeyInfo is the information about a key in the keystore.
// Public key may be empty for non-asymmetric key types.
type KeyInfo struct {
	Name      string
	KeyType   KeyType
	CreatedAt time.Time
	PublicKey []byte
	Metadata  []byte
}

// NewKeyInfo is a constructor that ensures all fields are set explicitly.
func NewKeyInfo(name string, keyType KeyType, createdAt time.Time, publicKey []byte, metadata []byte) KeyInfo {
	return KeyInfo{
		Name:      name,
		KeyType:   keyType,
		CreatedAt: createdAt,
		PublicKey: publicKey,
		Metadata:  metadata,
	}
}

type Keystore interface {
	Admin
	Reader
	Signer
	Encryptor
}

var ErrUnimplemented = errors.New("unimplemented")

// UnimplementedKeystore provides a no-op implementation of Keystore.
// It composes the specific unimplemented stubs for each interface.
// Clients should embed this struct to ensure forward compatibility with changes to the Keystore interface.
type UnimplementedKeystore struct {
	UnimplementedAdmin
	UnimplementedReader
	UnimplementedSigner
	UnimplementedEncryptor
}

type key struct {
	keyType    KeyType
	privateKey internal.Raw
	createdAt  time.Time
	metadata   []byte

	// Derived from private key on loading from storage.
	// Cached here for convenience.
	publicKey []byte
}

// newKey is a private constructor ensuring the internal key struct is fully and correctly populated.
func newKey(keyType KeyType, privateKey internal.Raw, publicKey []byte, createdAt time.Time, metadata []byte) key {
	return key{
		keyType:    keyType,
		privateKey: privateKey,
		publicKey:  publicKey,
		createdAt:  createdAt,
		metadata:   metadata,
	}
}

// EncryptionParams controls password-based encryption.
// Password is the secret used to derive the encryption key.
// ScryptParams control CPU/memory cost.
type EncryptionParams struct {
	Password     string
	ScryptParams ScryptParams
}

func publicKeyFromPrivateKey(privateKeyBytes internal.Raw, keyType KeyType) ([]byte, error) {
	switch keyType {
	case Ed25519:
		if len(internal.Bytes(privateKeyBytes)) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size: %d", len(internal.Bytes(privateKeyBytes)))
		}
		privateKey := ed25519.PrivateKey(internal.Bytes(privateKeyBytes))
		publicKey := privateKey.Public().(ed25519.PublicKey)
		return publicKey, nil
	case ECDSA_S256:
		// Here we use SEC1 (uncompressed) format for ECDSA public keys.
		// Its commonly used and EVM addresses are derived from this format.
		// We use the geth crypto library for secp256k1 support
		// because stdlib doesn't support it.
		privateKey, err := gethcrypto.ToECDSA(internal.Bytes(privateKeyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to convert private key to ECDSA private key: %w", err)
		}
		pubKey := gethcrypto.FromECDSAPub(&privateKey.PublicKey)
		return pubKey, nil
	case X25519:
		if len(internal.Bytes(privateKeyBytes)) != curve25519.ScalarSize {
			return nil, fmt.Errorf("invalid X25519 private key size: %d", len(internal.Bytes(privateKeyBytes)))
		}
		pubKey, err := curve25519.X25519(internal.Bytes(privateKeyBytes)[:], curve25519.Basepoint)
		if err != nil {
			return nil, fmt.Errorf("failed to derive shared secret: %w", err)
		}
		return pubKey, nil
	case ECDH_P256:
		curve := ecdh.P256()
		priv, err := curve.NewPrivateKey(internal.Bytes(privateKeyBytes))
		if err != nil {
			return nil, fmt.Errorf("invalid P-256 private key: %w", err)
		}
		return priv.PublicKey().Bytes(), nil
	default:
		// Some types may not have a public key.
		return []byte{}, nil
	}
}

type keystore struct {
	mu       sync.RWMutex
	keystore map[string]key
	storage  Storage
	enc      EncryptionParams
	lggr     *slog.Logger
}

type Option func(*keystore)

func WithLogger(l *slog.Logger) Option {
	return func(k *keystore) {
		if l != nil {
			k.lggr = l
		}
	}
}

func WithScryptParams(sp ScryptParams) Option {
	return func(k *keystore) {
		k.enc.ScryptParams = sp
	}
}

// LoadKeystore constructs a keystore with required args (ctx, storage, password)
// and optional settings via functional options.
// Default logger is a discard logger and default scrypt params are the standard scrypt params.
func LoadKeystore(ctx context.Context, storage Storage, password string, opts ...Option) (Keystore, error) {
	ks := &keystore{
		storage: storage,
		enc: EncryptionParams{
			Password:     password,
			ScryptParams: DefaultScryptParams,
		},
	}
	for _, opt := range opts {
		opt(ks)
	}
	if ks.lggr == nil || !testing.Testing() { // logging is not allowed in production binaries
		ks.lggr = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}
	err := ks.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load keystore: %w", err)
	}
	return ks, nil
}

func (k *keystore) load(ctx context.Context) error {
	encryptedKeystore, err := k.storage.GetEncryptedKeystore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get encrypted keystore: %w", err)
	}

	// If no data exists, return empty keystore
	if len(encryptedKeystore) == 0 {
		k.keystore = make(map[string]key)
		return nil
	}

	encryptedSecrets := gethkeystore.CryptoJSON{}
	err = json.Unmarshal(encryptedKeystore, &encryptedSecrets)
	if err != nil {
		return fmt.Errorf("failed to unmarshal encrypted keystore: %w", err)
	}
	decryptedKeystore, err := gethkeystore.DecryptDataV3(encryptedSecrets, k.enc.Password)
	if err != nil {
		return fmt.Errorf("failed to decrypt keystore: %w", err)
	}
	keystorepb := &serialization.Keystore{}
	err = proto.Unmarshal(decryptedKeystore, keystorepb)
	if err != nil {
		return fmt.Errorf("failed to unmarshal keystore: %w", err)
	}
	keystore := make(map[string]key)
	for _, k := range keystorepb.Keys {
		pkRaw := internal.NewRaw(k.PrivateKey)
		keyType := KeyType(k.KeyType)
		publicKey, err := publicKeyFromPrivateKey(pkRaw, keyType)
		if err != nil {
			return fmt.Errorf("failed to get public key from private key: %w", err)
		}
		keystore[k.Name] = newKey(keyType, pkRaw, publicKey, time.Unix(k.CreatedAt, 0), k.Metadata)
	}
	k.keystore = keystore
	return nil
}

func (k *keystore) save(ctx context.Context, keystore map[string]key) error {
	keystorepb := serialization.Keystore{
		Keys: make([]*serialization.Key, 0),
	}
	for name, k := range keystore {
		keystorepb.Keys = append(keystorepb.Keys, &serialization.Key{
			Name:       name,
			KeyType:    string(k.keyType),
			PrivateKey: internal.Bytes(k.privateKey),
			CreatedAt:  k.createdAt.Unix(),
			Metadata:   k.metadata,
		})
	}
	rawKeystore, err := proto.Marshal(&keystorepb)
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}
	encryptedSecrets, err := gethkeystore.EncryptDataV3(rawKeystore, []byte(k.enc.Password), k.enc.ScryptParams.N, k.enc.ScryptParams.P)
	if err != nil {
		return fmt.Errorf("failed to encrypt keystore: %w", err)
	}
	encryptedSecretsBytes, err := json.Marshal(encryptedSecrets)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted keystore: %w", err)
	}
	err = k.storage.PutEncryptedKeystore(ctx, encryptedSecretsBytes)
	if err != nil {
		return fmt.Errorf("failed to put encrypted keystore: %w", err)
	}
	return nil
}
