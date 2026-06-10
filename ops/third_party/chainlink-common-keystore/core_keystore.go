package keystore

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// CoreKeystore implements the core.Keystore interface for backwards compatibility with the old keystore
// https://github.com/smartcontractkit/chainlink-common/blob/main/pkg/types/core/keystore.go#L23
var _ core.Keystore = (*CoreKeystore)(nil)

type CoreKeystore struct {
	ks Keystore
}

func NewCoreKeystore(ks Keystore) *CoreKeystore {
	return &CoreKeystore{ks: ks}
}

func (c *CoreKeystore) Accounts(ctx context.Context) ([]string, error) {
	// List all the keys in the keystore.
	keys, err := c.ks.GetKeys(ctx, GetKeysRequest{})
	if err != nil {
		return nil, err
	}
	accounts := make([]string, 0, len(keys.Keys))
	for _, key := range keys.Keys {
		accounts = append(accounts, key.KeyInfo.Name)
	}
	return accounts, nil
}

func (c *CoreKeystore) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	resp, err := c.ks.Sign(ctx, SignRequest{
		KeyName: account,
		Data:    data,
	})
	if err != nil {
		return nil, err
	}
	return resp.Signature, nil
}

func (c *CoreKeystore) Decrypt(ctx context.Context, account string, data []byte) ([]byte, error) {
	resp, err := c.ks.Decrypt(ctx, DecryptRequest{
		KeyName:       account,
		EncryptedData: data,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
