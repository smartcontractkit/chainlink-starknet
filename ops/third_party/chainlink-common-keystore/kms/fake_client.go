package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/x509"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

type FakeKMSClient struct {
	keys []Key

	createdAt time.Time
}

type Key struct {
	KeyType    keystore.KeyType
	KeyID      string
	PrivateKey internal.Raw
}

func NewFakeKMSClient(keys []Key) (*FakeKMSClient, error) {
	return &FakeKMSClient{
		keys:      keys,
		createdAt: time.Now(),
	}, nil
}

func (m *FakeKMSClient) GetPublicKey(ctx context.Context, input *kms.GetPublicKeyInput, opts ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			var asn1PubKey []byte
			var err error
			switch key.KeyType {
			case keystore.ECDSA_S256:
				var ecdsaKey *ecdsa.PrivateKey
				ecdsaKey, err = crypto.ToECDSA(internal.Bytes(key.PrivateKey))
				if err != nil {
					return nil, err
				}
				asn1PubKey, err = SEC1ToASN1PublicKey(crypto.FromECDSAPub(&ecdsaKey.PublicKey))
				if err != nil {
					return nil, err
				}
			case keystore.Ed25519:
				// Ed25519 private key is 64 bytes, public key is the last 32 bytes
				ed25519PrivKey := ed25519.PrivateKey(internal.Bytes(key.PrivateKey))
				if len(ed25519PrivKey) != 64 {
					return nil, errors.New("invalid Ed25519 private key length")
				}
				pubKey := ed25519.PublicKey(ed25519PrivKey[32:])
				asn1PubKey, err = x509.MarshalPKIXPublicKey(pubKey)
				if err != nil {
					return nil, err
				}
			default:
				return nil, errors.New("unsupported key type")
			}
			return &kms.GetPublicKeyOutput{
				KeyId:     &key.KeyID,
				PublicKey: asn1PubKey,
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

func (m *FakeKMSClient) Sign(ctx context.Context, input *kms.SignInput, opts ...func(*kms.Options)) (*kms.SignOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			var signature []byte
			switch key.KeyType {
			case keystore.ECDSA_S256:
				ecdsaKey, err := crypto.ToECDSA(internal.Bytes(key.PrivateKey))
				if err != nil {
					return nil, err
				}
				sig, err := crypto.Sign(input.Message, ecdsaKey)
				if err != nil {
					return nil, err
				}
				signature, err = SEC1ToASN1Sig(sig)
				if err != nil {
					return nil, err
				}
			case keystore.Ed25519:
				// Ed25519 signatures are 64 bytes and don't need ASN.1 encoding
				ed25519PrivKey := ed25519.PrivateKey(internal.Bytes(key.PrivateKey))
				signature = ed25519.Sign(ed25519PrivKey, input.Message)
			default:
				return nil, errors.New("unsupported key type")
			}
			return &kms.SignOutput{
				KeyId:     &key.KeyID,
				Signature: signature,
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

// DescribeKey returns metadata about the key.
func (m *FakeKMSClient) DescribeKey(ctx context.Context, input *kms.DescribeKeyInput, opts ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			var keySpec kmstypes.KeySpec
			switch key.KeyType {
			case keystore.ECDSA_S256:
				keySpec = kmstypes.KeySpecEccSecgP256k1
			case keystore.Ed25519:
				keySpec = kmstypes.KeySpecEccNistEdwards25519
			default:
				return nil, errors.New("unsupported key type")
			}
			return &kms.DescribeKeyOutput{
				KeyMetadata: &kmstypes.KeyMetadata{
					KeyId:        &key.KeyID,
					KeySpec:      keySpec,
					CreationDate: &m.createdAt,
				},
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

// ListKeys returns a list of key IDs.
func (m *FakeKMSClient) ListKeys(ctx context.Context, input *kms.ListKeysInput, opts ...func(*kms.Options)) (*kms.ListKeysOutput, error) {
	keys := make([]kmstypes.KeyListEntry, 0, len(m.keys))
	for _, key := range m.keys {
		keyID := key.KeyID
		keys = append(keys, kmstypes.KeyListEntry{
			KeyId: &keyID,
		})
	}
	return &kms.ListKeysOutput{
		Keys: keys,
	}, nil
}
