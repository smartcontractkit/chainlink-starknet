package keystore

import (
	"context"
	"sync"
)

var _ Storage = &MemoryStorage{}

// MemoryStorage implements Storage using in-memory storage
type MemoryStorage struct {
	mu   sync.RWMutex
	data []byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: []byte{}, // Initialize with empty byte slice instead of nil
	}
}

func (m *MemoryStorage) GetEncryptedKeystore(ctx context.Context) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the data
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *MemoryStorage) PutEncryptedKeystore(ctx context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy of the data, treating nil as empty bytes
	if data == nil {
		m.data = []byte{}
	} else {
		m.data = make([]byte, len(data))
		copy(m.data, data)
	}

	return nil
}
