package keystore

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

var (
	ErrInvalidSignRequest   = errors.New("invalid sign request")
	ErrInvalidVerifyRequest = errors.New("invalid verify request")
)

type SignRequest struct {
	KeyName string
	Data    []byte
}

type SignResponse struct {
	Signature []byte
}

type VerifyRequest struct {
	KeyType   KeyType
	PublicKey []byte
	Data      []byte
	Signature []byte
}

type VerifyResponse struct {
	Valid bool
}

type Signer interface {
	Sign(ctx context.Context, req SignRequest) (SignResponse, error)
	Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error)
}

// UnimplementedSigner returns ErrUnimplemented for all Signer methods.
type UnimplementedSigner struct{}

func (UnimplementedSigner) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	return SignResponse{}, fmt.Errorf("Signer.Sign: %w", ErrUnimplemented)
}

func (UnimplementedSigner) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	return VerifyResponse{}, fmt.Errorf("Signer.Verify: %w", ErrUnimplemented)
}

func (k *keystore) Sign(ctx context.Context, req SignRequest) (SignResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keystore[req.KeyName]
	if !ok {
		return SignResponse{}, fmt.Errorf("%s: %w", req.KeyName, ErrKeyNotFound)
	}
	switch key.keyType {
	case Ed25519:
		privateKey := ed25519.PrivateKey(internal.Bytes(key.privateKey))
		signature := ed25519.Sign(privateKey, req.Data)
		return SignResponse{
			Signature: signature,
		}, nil
	case ECDSA_S256:
		if len(req.Data) != 32 {
			return SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), ErrInvalidSignRequest)
		}
		privateKey, err := gethcrypto.ToECDSA(internal.Bytes(key.privateKey))
		if err != nil {
			return SignResponse{}, fmt.Errorf("failed to convert private key to ECDSA private key: %w", err)
		}
		signature, err := gethcrypto.Sign(req.Data, privateKey)
		if err != nil {
			return SignResponse{}, err
		}
		return SignResponse{
			Signature: signature,
		}, nil
	default:
		return SignResponse{}, fmt.Errorf("unsupported key type: %s, available key types: %s", key.keyType, AllDigitalSignatureKeyTypes.String())
	}
}

func (k *keystore) Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	// Note don't need the lock since this is a pure function.
	return Verify(ctx, req)
}

func Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error) {
	switch req.KeyType {
	case Ed25519:
		if len(req.Signature) != 64 {
			return VerifyResponse{}, fmt.Errorf("signature must be 64 bytes for Ed25519, got %d: %w", len(req.Signature), ErrInvalidVerifyRequest)
		}
		if len(req.PublicKey) != 32 {
			return VerifyResponse{}, fmt.Errorf("public key must be 32 bytes for Ed25519, got %d: %w", len(req.PublicKey), ErrInvalidVerifyRequest)
		}
		publicKey := ed25519.PublicKey(req.PublicKey)
		signature := ed25519.Verify(publicKey, req.Data, req.Signature)
		return VerifyResponse{
			Valid: signature,
		}, nil
	case ECDSA_S256:
		if len(req.Data) != 32 {
			return VerifyResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), ErrInvalidVerifyRequest)
		}
		// ECDSA_S256 public keys are in SEC1 (uncompressed) format
		if len(req.PublicKey) != 65 {
			return VerifyResponse{}, fmt.Errorf("public key must be 65 bytes for ECDSA_S256, got %d: %w", len(req.PublicKey), ErrInvalidVerifyRequest)
		}
		if len(req.Signature) != 65 {
			return VerifyResponse{}, fmt.Errorf("signature must be 65 bytes for ECDSA_S256, got %d: %w", len(req.Signature), ErrInvalidVerifyRequest)
		}
		// VerifySignature expects 64 bytes [R || S] without the V byte
		// Strip the V byte (last byte) from the 65-byte signature
		valid := gethcrypto.VerifySignature(req.PublicKey, req.Data, req.Signature[:64])
		return VerifyResponse{
			Valid: valid,
		}, nil
	default:
		return VerifyResponse{}, fmt.Errorf("unsupported key type: %s, available key types: %s", req.KeyType, AllDigitalSignatureKeyTypes.String())
	}
}
