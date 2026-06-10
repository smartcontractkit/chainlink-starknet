package kms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"

	"crypto/ed25519"
	"crypto/x509"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

type keystoreSignerReader struct {
	client Client
}

func NewKeystore(client Client) (interface {
	keystore.Reader
	keystore.Signer
}, error) {
	return &keystoreSignerReader{
		client: client,
	}, nil
}

// keySpecToKeyType converts an AWS KMS KeySpec to a keystore KeyType.
// AWS KMS supports:
//   - ECC_SECG_P256K1 (secp256k1) -> ECDSA_S256
//   - ECC_NIST_EDWARDS25519 (Ed25519) -> Ed25519
func keySpecToKeyType(keySpec kmstypes.KeySpec) (keystore.KeyType, error) {
	switch keySpec {
	case kmstypes.KeySpecEccSecgP256k1:
		return keystore.ECDSA_S256, nil
	case kmstypes.KeySpecEccNistEdwards25519:
		return keystore.Ed25519, nil
	default:
		return "", fmt.Errorf("unsupported KMS key spec: %s (supported: ECC_SECG_P256K1, ECC_NIST_EDWARDS25519)", keySpec)
	}
}

// GetKeys lists all keys in the KMS keystore.
// Note: possible that ListKeys is not supported for a given AWS role
// but specific GetPublicKey/DescribeKey are supported. We transparently
// let the user use whatever they are permitted to given their AWS perms.
func (k *keystoreSignerReader) GetKeys(ctx context.Context, req keystore.GetKeysRequest) (keystore.GetKeysResponse, error) {
	if len(req.KeyNames) == 0 {
		listResp, err := k.client.ListKeys(ctx, &kms.ListKeysInput{})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to list KMS keys: %w", err)
		}
		req.KeyNames = make([]string, 0, len(listResp.Keys))
		for _, key := range listResp.Keys {
			if key.KeyId != nil {
				req.KeyNames = append(req.KeyNames, *key.KeyId)
			}
		}
	}

	keys := make([]keystore.GetKeyResponse, 0, len(req.KeyNames))
	for _, keyID := range req.KeyNames {
		// Get public key
		key, err := k.client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
			KeyId: &keyID,
		})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to get public key for key %s: %w", keyID, err)
		}

		// Get key metadata to determine key type and creation date
		describeKey, err := k.client.DescribeKey(ctx, &kms.DescribeKeyInput{
			KeyId: &keyID,
		})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to describe key %s: %w", keyID, err)
		}
		createdAt := time.Unix(describeKey.KeyMetadata.CreationDate.Unix(), 0)

		// Convert KMS KeySpec to keystore KeyType
		keyType, err := keySpecToKeyType(describeKey.KeyMetadata.KeySpec)
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("key %s: %w", keyID, err)
		}
		var publicKeyBytes []byte
		switch keyType {
		case keystore.ECDSA_S256:
			publicKeyBytes, err = ASN1ToSEC1PublicKey(key.PublicKey)
			if err != nil {
				return keystore.GetKeysResponse{}, fmt.Errorf("failed to convert public key for key %s: %w", keyID, err)
			}
		case keystore.Ed25519:
			// ed25519 supported by standard libraries unlike secp256k1.
			pubKey, err := x509.ParsePKIXPublicKey(key.PublicKey)
			if err != nil {
				return keystore.GetKeysResponse{}, fmt.Errorf("failed to convert Ed25519 public key for key %s: %w", keyID, err)
			}
			ed25519PubKey, ok := pubKey.(ed25519.PublicKey)
			if !ok {
				return keystore.GetKeysResponse{}, fmt.Errorf("failed to convert Ed25519 public key for key %s to ed25519.PublicKey: %w", keyID, err)
			}
			publicKeyBytes = ed25519PubKey
		default:
			return keystore.GetKeysResponse{}, fmt.Errorf("unsupported key type: %s", keyType)
		}

		keys = append(keys, keystore.GetKeyResponse{
			KeyInfo: keystore.NewKeyInfo(keyID, keyType, createdAt, publicKeyBytes, []byte{}),
		})
	}
	return keystore.GetKeysResponse{Keys: keys}, nil
}

// Sign signs data using the KMS key specified by the key name.
func (k *keystoreSignerReader) Sign(ctx context.Context, req keystore.SignRequest) (keystore.SignResponse, error) {
	key, err := k.client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: &req.KeyName,
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to get public key for key %s: %w", req.KeyName, err)
	}
	describeKey, err := k.client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: &req.KeyName,
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to describe key %s: %w", req.KeyName, err)
	}
	keyType, err := keySpecToKeyType(describeKey.KeyMetadata.KeySpec)
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("key %s: %w", req.KeyName, err)
	}

	switch keyType {
	case keystore.ECDSA_S256:
		if len(req.Data) != 32 {
			return keystore.SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidSignRequest)
		}
		pubKeyBytes, err := ASN1ToSEC1PublicKey(key.PublicKey)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to convert public key for KeyId=%s: %w", req.KeyName, err)
		}

		// MessageType is digest because its prehashed.
		sig, err := k.client.Sign(ctx, &kms.SignInput{
			KeyId:            &req.KeyName,
			Message:          req.Data,
			SigningAlgorithm: kmstypes.SigningAlgorithmSpecEcdsaSha256,
			MessageType:      kmstypes.MessageTypeDigest,
		})
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to sign data: %w", err)
		}
		signature, err := ASN1ToSEC1Sig(sig.Signature, pubKeyBytes, req.Data)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to convert KMS signature to SEC1 signature: %w", err)
		}
		return keystore.SignResponse{
			Signature: signature,
		}, nil
	case keystore.Ed25519:
		// Ed25519 can sign arbitrary length messages, uses RAW message type
		sig, err := k.client.Sign(ctx, &kms.SignInput{
			KeyId:            &req.KeyName,
			Message:          req.Data,
			SigningAlgorithm: kmstypes.SigningAlgorithmSpecEd25519Sha512,
			MessageType:      kmstypes.MessageTypeRaw,
		})
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to sign data: %w", err)
		}
		// Ed25519 signatures from KMS are already in the correct format (64 bytes)
		if len(sig.Signature) != 64 {
			return keystore.SignResponse{}, fmt.Errorf("invalid Ed25519 signature length: expected 64 bytes, got %d", len(sig.Signature))
		}
		return keystore.SignResponse{
			Signature: sig.Signature,
		}, nil
	default:
		return keystore.SignResponse{}, fmt.Errorf("key %s: %w", req.KeyName, keystore.ErrInvalidSignRequest)
	}
}

func (k *keystoreSignerReader) Verify(ctx context.Context, req keystore.VerifyRequest) (keystore.VerifyResponse, error) {
	return keystore.Verify(ctx, req)
}
