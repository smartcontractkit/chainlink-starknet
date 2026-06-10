package keystore

import "context"

// Storage is where we store the encrypted key material.
type Storage interface {
	GetEncryptedKeystore(ctx context.Context) ([]byte, error)
	// PutEncryptedKeystore must atomically replace the entire encrypted keystore.
	PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error
}
