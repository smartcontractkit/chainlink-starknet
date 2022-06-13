package starknet

import (
	"sync"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/relay/starknet/db"
)

var DefaultConfigSet = ConfigSet{
	// defaults
}

type ConfigSet struct {
	// chain configuration parameters
}

type Config interface {
	// interface to interact with chain configuration
	Update(db.ChainCfg)
}

var _ Config = (*config)(nil)

type config struct {
	defaults ConfigSet
	chain    db.ChainCfg
	chainMu  sync.RWMutex
	lggr     logger.Logger
}

func NewConfig(dbCfg db.ChainCfg, lggr logger.Logger) *config {
	return &config{
		defaults: DefaultConfigSet,
		chain:    dbCfg,
		lggr:     lggr,
	}
}

func (c *config) Update(dbCfg db.ChainCfg) {
	c.chainMu.Lock()
	c.chain = dbCfg
	c.chainMu.Unlock()
}
