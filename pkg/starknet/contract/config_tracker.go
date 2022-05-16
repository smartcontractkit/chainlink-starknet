package contract

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractConfigTracker = (*ContractTracker)(nil)

func (c *ContractTracker) Notify() <-chan struct{} {
	return nil
}

func (c *ContractTracker) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	cfg, err := c.ReadCCFromCache()
	return cfg.configBlock, cfg.config.ConfigDigest, err
}

func (c *ContractTracker) LatestConfig(ctx context.Context, changedInBlock uint64) (types.ContractConfig, error) {
	cfg, err := c.ReadCCFromCache()
	if err != nil {
		return types.ContractConfig{}, err
	}
	return cfg.config, nil
}

func (c *ContractTracker) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
}
