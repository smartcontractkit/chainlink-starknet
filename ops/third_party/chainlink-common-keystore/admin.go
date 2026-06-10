package keystore

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"golang.org/x/crypto/curve25519"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/serialization"
)

const (
	MaxKeyNameLength  = 1000
	MaxMetadataLength = 1024 * 1024 // 1mb
)

var (
	ErrKeyAlreadyExists   = fmt.Errorf("key already exists")
	ErrInvalidKeyName     = fmt.Errorf("invalid key name")
	ErrKeyNotFound        = fmt.Errorf("key not found")
	ErrUnsupportedKeyType = fmt.Errorf("unsupported key type")
)

type CreateKeysRequest struct {
	Keys []CreateKeyRequest
}

type CreateKeyRequest struct {
	KeyName string
	KeyType KeyType
}

type CreateKeysResponse struct {
	Keys []CreateKeyResponse
}

type CreateKeyResponse struct {
	KeyInfo KeyInfo
}

type DeleteKeysRequest struct {
	KeyNames []string
}

type DeleteKeysResponse struct{}

type ImportKeysRequest struct {
	Keys []ImportKeyRequest
}

type ImportKeyRequest struct {
	NewKeyName string
	Data       []byte
	Password   string
}

type ImportKeysResponse struct{}

type ExportKeyParam struct {
	KeyName string
	Enc     EncryptionParams
}

type ExportKeysRequest struct {
	Keys []ExportKeyParam
}

type ExportKeysResponse struct {
	Keys []ExportKeyResponse
}

type ExportKeyResponse struct {
	KeyName string
	Data    []byte
}

type SetMetadataRequest struct {
	Updates []SetMetadataUpdate
}

type SetMetadataUpdate struct {
	KeyName  string
	Metadata []byte
}

type SetMetadataResponse struct{}

type RenameKeyRequest struct {
	OldName string
	NewName string
}

type RenameKeyResponse struct{}

type Admin interface {
	// CreateKeys creates multiple keys in a single atomic operation.
	// The response preserves the order of the keys in the request.
	// Returns ErrKeyAlreadyExists if any key name already exists,
	// ErrInvalidKeyName if any key name is invalid,
	// ErrUnsupportedKeyType if any key type is not supported.
	CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error)

	// DeleteKeys deletes multiple keys in a single atomic operation.
	// Returns ErrKeyNotFound if any key does not exist.
	DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error)

	// ImportKeys imports multiple encrypted keys in a single atomic operation.
	// Keys can be renamed during import using NewKeyName.
	// Returns ErrKeyAlreadyExists if a key with the target name exists,
	// ErrInvalidKeyName if the target name is invalid.
	ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error)

	// ExportKeys exports multiple keys in encrypted format.
	// Each key is encrypted using the parameters specified in the request.
	// Returns ErrKeyNotFound if any key does not exist.
	ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error)

	// SetMetadata updates metadata for multiple keys in a single atomic operation.
	// Returns ErrKeyNotFound if any key does not exist.
	SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error)

	// RenameKey renames an existing key.
	// Returns ErrKeyNotFound if the old key name does not exist,
	// ErrKeyAlreadyExists if the new key name exists,
	// ErrInvalidKeyName if the new key name is invalid.
	RenameKey(ctx context.Context, req RenameKeyRequest) (RenameKeyResponse, error)
}

// UnimplementedAdmin returns ErrUnimplemented for all Admin methods.
type UnimplementedAdmin struct{}

func (UnimplementedAdmin) CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error) {
	return CreateKeysResponse{}, fmt.Errorf("Admin.CreateKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error) {
	return DeleteKeysResponse{}, fmt.Errorf("Admin.DeleteKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error) {
	return ImportKeysResponse{}, fmt.Errorf("Admin.ImportKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) ExportKeys(ctx context.Context, req ExportKeysRequest) (ExportKeysResponse, error) {
	return ExportKeysResponse{}, fmt.Errorf("Admin.ExportKeys: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error) {
	return SetMetadataResponse{}, fmt.Errorf("Admin.SetMetadata: %w", ErrUnimplemented)
}

func (UnimplementedAdmin) RenameKey(ctx context.Context, req RenameKeyRequest) (RenameKeyResponse, error) {
	return RenameKeyResponse{}, fmt.Errorf("Admin.RenameKey: %w", ErrUnimplemented)
}

func ValidKeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	// Just a sanity bound.
	if len(name) > MaxKeyNameLength {
		return fmt.Errorf("key name cannot be longer than %d characters", MaxKeyNameLength)
	}
	return nil
}

// CreateKeys creates multiple keys in a single operation. The response preserves the order of the request.
// It's atomic - either all keys are created or none are created.
func (ks *keystore) CreateKeys(ctx context.Context, req CreateKeysRequest) (CreateKeysResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ksCopy := maps.Clone(ks.keystore)
	var responses []CreateKeyResponse
	for _, keyReq := range req.Keys {
		if _, ok := ksCopy[keyReq.KeyName]; ok {
			return CreateKeysResponse{}, fmt.Errorf("%w: %s", ErrKeyAlreadyExists, keyReq.KeyName)
		}
		if err := ValidKeyName(keyReq.KeyName); err != nil {
			return CreateKeysResponse{}, fmt.Errorf("%w: %s", ErrInvalidKeyName, err)
		}
		switch keyReq.KeyType {
		case Ed25519:
			_, privateKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey), publicKey, time.Now(), []byte{})
		case ECDSA_S256:
			privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate ECDSA_S256 key: %w", err)
			}
			privateKeyBytes := make([]byte, 32)
			privateKey.D.FillBytes(privateKeyBytes)
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKeyBytes), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKeyBytes), publicKey, time.Now(), []byte{})
		case X25519:
			privateKey := [curve25519.ScalarSize]byte{}
			_, err := rand.Read(privateKey[:])
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate Curve25519 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey[:]), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey[:]), publicKey, time.Now(), []byte{})
		case ECDH_P256:
			privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to generate ECDH_P256 key: %w", err)
			}
			publicKey, err := publicKeyFromPrivateKey(internal.NewRaw(privateKey.Bytes()), keyReq.KeyType)
			if err != nil {
				return CreateKeysResponse{}, fmt.Errorf("failed to get public key from private key: %w", err)
			}
			ksCopy[keyReq.KeyName] = newKey(keyReq.KeyType, internal.NewRaw(privateKey.Bytes()), publicKey, time.Now(), []byte{})
		default:
			return CreateKeysResponse{}, fmt.Errorf("%w: %s, available key types: %s", ErrUnsupportedKeyType, keyReq.KeyType, AllKeyTypes.String())
		}

		created := ksCopy[keyReq.KeyName].createdAt
		k := ksCopy[keyReq.KeyName]
		responses = append(responses, CreateKeyResponse{
			KeyInfo: NewKeyInfo(keyReq.KeyName, keyReq.KeyType, created, k.publicKey, k.metadata),
		})
	}

	// Persist it to storage.
	if err := ks.save(ctx, ksCopy); err != nil {
		return CreateKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	ks.keystore = ksCopy
	return CreateKeysResponse{Keys: responses}, nil
}

func (k *keystore) DeleteKeys(ctx context.Context, req DeleteKeysRequest) (DeleteKeysResponse, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	ksCopy := maps.Clone(k.keystore)
	for _, name := range req.KeyNames {
		if _, ok := ksCopy[name]; !ok {
			return DeleteKeysResponse{}, fmt.Errorf("%w: %s", ErrKeyNotFound, name)
		}
		delete(ksCopy, name)
	}
	if err := k.save(ctx, ksCopy); err != nil {
		return DeleteKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	k.keystore = ksCopy
	return DeleteKeysResponse{}, nil
}

func (ks *keystore) ImportKeys(ctx context.Context, req ImportKeysRequest) (ImportKeysResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ksCopy := maps.Clone(ks.keystore)
	for i, keyReq := range req.Keys {
		encData := gethkeystore.CryptoJSON{}
		err := json.Unmarshal(keyReq.Data, &encData)
		if err != nil {
			return ImportKeysResponse{}, fmt.Errorf("key num = %d, failed to unmarshal encrypted import data: %w", i, err)
		}
		decData, err := gethkeystore.DecryptDataV3(encData, keyReq.Password)
		if err != nil {
			return ImportKeysResponse{}, fmt.Errorf("key num = %d, failed to decrypt key: %w", i, err)
		}
		keypb := &serialization.Key{}
		err = proto.Unmarshal(decData, keypb)
		if err != nil {
			return ImportKeysResponse{}, fmt.Errorf("key num = %d, failed to unmarshal key: %w", i, err)
		}
		pkRaw := internal.NewRaw(keypb.PrivateKey)
		keyType := KeyType(keypb.KeyType)
		if !slices.Contains(AllKeyTypes, keyType) {
			return ImportKeysResponse{}, fmt.Errorf("%w: %s, available key types: %s", ErrUnsupportedKeyType, keyType, AllKeyTypes.String())
		}
		publicKey, err := publicKeyFromPrivateKey(pkRaw, keyType)
		if err != nil {
			return ImportKeysResponse{}, fmt.Errorf("key num = %d, failed to get public key from private key: %w", i, err)
		}
		metadata := keypb.Metadata
		// The proto compiler sets empty slices to nil during the serialization (https://github.com/golang/protobuf/issues/1348).
		// We set metadata back to empty slice to be consistent with the Create method which initializes it as such.
		if metadata == nil {
			metadata = []byte{}
		}
		if len(metadata) > MaxMetadataLength {
			return ImportKeysResponse{}, fmt.Errorf("key num = %d, metadata of length %d exceeds maximum length of %d bytes", i, len(metadata), MaxMetadataLength)
		}

		keyName := keyReq.NewKeyName
		if keyName == "" {
			keyName = keypb.Name
		}
		if err := ValidKeyName(keyName); err != nil {
			return ImportKeysResponse{}, fmt.Errorf("key name = %s, %w: %s", keyName, ErrInvalidKeyName, err)
		}
		if _, ok := ksCopy[keyName]; ok {
			return ImportKeysResponse{}, fmt.Errorf("key name = %s, %w: %s", keyName, ErrKeyAlreadyExists, keyName)
		}
		ksCopy[keyName] = newKey(keyType, pkRaw, publicKey, time.Unix(keypb.CreatedAt, 0), metadata)
	}
	// Persist it to storage.
	if err := ks.save(ctx, ksCopy); err != nil {
		return ImportKeysResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	ks.keystore = ksCopy
	return ImportKeysResponse{}, nil
}

func (ks *keystore) ExportKeys(_ context.Context, req ExportKeysRequest) (ExportKeysResponse, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	result := ExportKeysResponse{}
	for _, keyReq := range req.Keys {
		key, ok := ks.keystore[keyReq.KeyName]
		if !ok {
			return ExportKeysResponse{}, fmt.Errorf("%w: %s", ErrKeyNotFound, keyReq.KeyName)
		}
		keypb := &serialization.Key{
			Name:       keyReq.KeyName,
			KeyType:    string(key.keyType),
			PrivateKey: internal.Bytes(key.privateKey),
			CreatedAt:  key.createdAt.Unix(),
			Metadata:   key.metadata,
		}
		serialized, err := proto.Marshal(keypb)
		if err != nil {
			return ExportKeysResponse{}, fmt.Errorf("key = %s, failed to marshal key: %w", keyReq.KeyName, err)
		}
		encData, err := gethkeystore.EncryptDataV3(serialized, []byte(keyReq.Enc.Password), keyReq.Enc.ScryptParams.N, keyReq.Enc.ScryptParams.P)
		if err != nil {
			return ExportKeysResponse{}, fmt.Errorf("key = %s, failed to encrypt key: %w", keyReq.KeyName, err)
		}
		encDataBytes, err := json.Marshal(encData)
		if err != nil {
			return ExportKeysResponse{}, fmt.Errorf("key = %s, failed to marshal encrypted key: %w", keyReq.KeyName, err)
		}
		result.Keys = append(result.Keys, ExportKeyResponse{
			KeyName: keyReq.KeyName,
			Data:    encDataBytes,
		})
	}
	return result, nil
}

func (ks *keystore) SetMetadata(ctx context.Context, req SetMetadataRequest) (SetMetadataResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ksCopy := maps.Clone(ks.keystore)
	for _, metReq := range req.Updates {
		if len(metReq.Metadata) > MaxMetadataLength {
			return SetMetadataResponse{}, fmt.Errorf("metadata for key %s exceeds maximum length of %d bytes", metReq.KeyName, MaxMetadataLength)
		}
		key, ok := ksCopy[metReq.KeyName]
		if !ok {
			return SetMetadataResponse{}, fmt.Errorf("%w: %s", ErrKeyNotFound, metReq.KeyName)
		}
		key.metadata = metReq.Metadata
		ksCopy[metReq.KeyName] = key
	}
	// Persist it to storage.
	if err := ks.save(ctx, ksCopy); err != nil {
		return SetMetadataResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}
	// If we succeed to save, update the in memory keystore.
	ks.keystore = ksCopy
	return SetMetadataResponse{}, nil
}

func (ks *keystore) RenameKey(ctx context.Context, req RenameKeyRequest) (RenameKeyResponse, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if req.NewName == req.OldName {
		return RenameKeyResponse{}, nil
	}

	if err := ValidKeyName(req.NewName); err != nil {
		return RenameKeyResponse{}, fmt.Errorf("%w: %s", ErrInvalidKeyName, err)
	}

	if _, ok := ks.keystore[req.NewName]; ok {
		return RenameKeyResponse{}, fmt.Errorf("%w: %s", ErrKeyAlreadyExists, req.NewName)
	}

	k, ok := ks.keystore[req.OldName]
	if !ok {
		return RenameKeyResponse{}, fmt.Errorf("%w: %s", ErrKeyNotFound, req.OldName)
	}

	ksCopy := maps.Clone(ks.keystore)
	ksCopy[req.NewName] = k
	delete(ksCopy, req.OldName)

	if err := ks.save(ctx, ksCopy); err != nil {
		return RenameKeyResponse{}, fmt.Errorf("failed to save keystore: %w", err)
	}

	ks.keystore = ksCopy
	return RenameKeyResponse{}, nil
}
