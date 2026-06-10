package keystore

import (
	"context"
	"fmt"
	"sort"
)

type GetKeysRequest struct {
	KeyNames []string
}

type GetKeysResponse struct {
	Keys []GetKeyResponse
}

type GetKeyRequest struct {
	KeyName string
}

type GetKeyResponse struct {
	KeyInfo KeyInfo
}

// Reader is the interface for reading keys from the keystore.
// GetKeys returns all keys in the keystore if no names are provided, or the keys with the given names.
// Keys are sorted by name in lexicographic order.
type Reader interface {
	GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error)
}

// UnimplementedReader returns ErrUnimplemented for all Reader methods.
type UnimplementedReader struct{}

func (UnimplementedReader) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	return GetKeysResponse{}, fmt.Errorf("Reader.GetKeys: %w", ErrUnimplemented)
}

func (k *keystore) GetKeys(ctx context.Context, req GetKeysRequest) (GetKeysResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	// If no names provided, return all keys
	if len(req.KeyNames) == 0 {
		responses := make([]GetKeyResponse, 0, len(k.keystore))
		for name, key := range k.keystore {
			responses = append(responses, GetKeyResponse{
				KeyInfo: NewKeyInfo(name, key.keyType, key.createdAt, key.publicKey, key.metadata),
			})
		}
		sort.Slice(responses, func(i, j int) bool { return responses[i].KeyInfo.Name < responses[j].KeyInfo.Name })
		return GetKeysResponse{Keys: responses}, nil
	}

	responses := make([]GetKeyResponse, 0, len(req.KeyNames))
	seen := make(map[string]bool)
	for _, name := range req.KeyNames {
		key, ok := k.keystore[name]
		if !ok {
			return GetKeysResponse{}, fmt.Errorf("key not found: %s", name)
		}
		if seen[name] {
			return GetKeysResponse{}, fmt.Errorf("key %s provided multiple times", name)
		}
		seen[name] = true
		responses = append(responses, GetKeyResponse{
			KeyInfo: NewKeyInfo(name, key.keyType, key.createdAt, key.publicKey, key.metadata),
		})
	}
	sort.Slice(responses, func(i, j int) bool { return responses[i].KeyInfo.Name < responses[j].KeyInfo.Name })
	return GetKeysResponse{Keys: responses}, nil
}
