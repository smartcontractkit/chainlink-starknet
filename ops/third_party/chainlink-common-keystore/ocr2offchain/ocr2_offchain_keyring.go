package ocr2offchain

import (
	"context"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/keystore"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"golang.org/x/crypto/curve25519"
)

const (
	OCR2OffchainSigning    = "ocr2_offchain_signing"
	OCR2OffchainEncryption = "ocr2_offchain_encryption"
	PrefixOCR2Offchain     = "ocr2_offchain"
)

func CreateOCR2OffchainKeyring(ctx context.Context, ks keystore.Keystore, keyringName string) (ocrtypes.OffchainKeyring, error) {
	signingKeyPath := keystore.NewKeyPath(PrefixOCR2Offchain, keyringName, OCR2OffchainSigning)
	encryptionKeyPath := keystore.NewKeyPath(PrefixOCR2Offchain, keyringName, OCR2OffchainEncryption)
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				KeyName: signingKeyPath.String(),
				KeyType: keystore.Ed25519,
			},
			{
				KeyName: encryptionKeyPath.String(),
				KeyType: keystore.X25519,
			},
		},
	}
	resp, err := ks.CreateKeys(ctx, createReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Keys) != 2 {
		return nil, fmt.Errorf("expected 2 keys, got %d", len(resp.Keys))
	}
	return &evmOffchainKeyring{
		ks:                    ks,
		signingKeyPath:        signingKeyPath,
		encryptionKeyPath:     encryptionKeyPath,
		offchainKey:           resp.Keys[0].KeyInfo,
		offchainEncryptionKey: resp.Keys[1].KeyInfo,
	}, nil
}

// ListOCR2OffchainKeyrings lists OCR2 offchain keyrings. If no local names provided, returns all OCR2 offchain keyrings.
func GetOCR2OffchainKeyrings(ctx context.Context, ks keystore.Keystore, keyRingNames []string) ([]ocrtypes.OffchainKeyring, error) {
	var names []string
	if len(keyRingNames) > 0 {
		for _, keyRingName := range keyRingNames {
			names = append(names, keystore.NewKeyPath(PrefixOCR2Offchain, keyRingName, OCR2OffchainSigning).String())
			names = append(names, keystore.NewKeyPath(PrefixOCR2Offchain, keyRingName, OCR2OffchainEncryption).String())
		}
	}

	getReq := keystore.GetKeysRequest{KeyNames: names}
	resp, err := ks.GetKeys(ctx, getReq)
	if err != nil {
		return nil, err
	}

	// Group by keyrings.
	keyRingMap := make(map[string][]keystore.KeyInfo)
	for _, key := range resp.Keys {
		if !strings.HasPrefix(key.KeyInfo.Name, keystore.NewKeyPath(PrefixOCR2Offchain).String()) {
			continue
		}
		keyPath := keystore.NewKeyPathFromString(key.KeyInfo.Name)
		// Example:
		// /ocr2_offchain/keyring_name/ocr2_offchain_signing
		// /ocr2_offchain/keyring_name/ocr2_offchain_encryption
		// Group by keyring name (first 3 segments: ocr2_offchain/keyring_name)
		keyRingMap[keyPath[:2].String()] = append(keyRingMap[keyPath[:2].String()], key.KeyInfo)
	}

	var keyrings []ocrtypes.OffchainKeyring
	for _, keyInfos := range keyRingMap {
		// Find signing and encryption keys
		var signingKey, encryptionKey keystore.KeyInfo
		for _, keyInfo := range keyInfos {
			if strings.HasSuffix(keyInfo.Name, OCR2OffchainSigning) {
				signingKey = keyInfo
			} else if strings.HasSuffix(keyInfo.Name, OCR2OffchainEncryption) {
				encryptionKey = keyInfo
			}
		}
		keyrings = append(keyrings, &evmOffchainKeyring{
			ks:                    ks,
			signingKeyPath:        keystore.NewKeyPathFromString(signingKey.Name),
			encryptionKeyPath:     keystore.NewKeyPathFromString(encryptionKey.Name),
			offchainKey:           signingKey,
			offchainEncryptionKey: encryptionKey,
		})
	}
	return keyrings, nil
}

var _ ocrtypes.OffchainKeyring = &evmOffchainKeyring{}

type evmOffchainKeyring struct {
	ks                    keystore.Keystore
	signingKeyPath        keystore.KeyPath
	encryptionKeyPath     keystore.KeyPath
	offchainKey           keystore.KeyInfo
	offchainEncryptionKey keystore.KeyInfo
}

func (k *evmOffchainKeyring) ConfigEncryptionKeyPath() keystore.KeyPath {
	return k.encryptionKeyPath
}

func (k *evmOffchainKeyring) ConfigSigningKeyPath() keystore.KeyPath {
	return k.signingKeyPath
}

func (k *evmOffchainKeyring) OffchainPublicKey() ocrtypes.OffchainPublicKey {
	var pubKey ocrtypes.OffchainPublicKey
	copy(pubKey[:], k.offchainKey.PublicKey)
	return pubKey
}

func (k *evmOffchainKeyring) ConfigEncryptionPublicKey() ocrtypes.ConfigEncryptionPublicKey {
	var pubKey ocrtypes.ConfigEncryptionPublicKey
	copy(pubKey[:], k.offchainEncryptionKey.PublicKey)
	return pubKey
}

func (k *evmOffchainKeyring) OffchainSign(msg []byte) ([]byte, error) {
	signResp, err := k.ks.Sign(context.Background(), keystore.SignRequest{
		KeyName: k.offchainKey.Name,
		Data:    msg,
	})
	return signResp.Signature, err
}

func (k *evmOffchainKeyring) ConfigDiffieHellman(point [curve25519.PointSize]byte) ([curve25519.PointSize]byte, error) {
	resp, err := k.ks.DeriveSharedSecret(context.Background(), keystore.DeriveSharedSecretRequest{
		KeyName:      k.offchainEncryptionKey.Name,
		RemotePubKey: point[:],
	})
	if err != nil {
		return [curve25519.PointSize]byte{}, err
	}

	var sharedPoint [curve25519.PointSize]byte
	copy(sharedPoint[:], resp.SharedSecret)
	return sharedPoint, nil
}
