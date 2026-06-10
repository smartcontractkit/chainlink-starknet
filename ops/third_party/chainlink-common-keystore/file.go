package keystore

import (
	"bytes"
	"context"
	"os"

	"github.com/smartcontractkit/chainlink-common/keystore/internal/atomicfile"
)

var _ Storage = &FileStorage{}

// FileStorage implements Storage using a file
type FileStorage struct {
	name string
}

func NewFileStorage(name string) *FileStorage {
	return &FileStorage{
		name: name,
	}
}

func (f *FileStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	return os.ReadFile(f.name)
}

func (f *FileStorage) PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) error {
	return atomicfile.WriteFile(f.name, bytes.NewReader(encryptedKeystore), 0600)
}
