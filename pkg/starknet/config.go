package starknet

import (
	"sync"
<<<<<<< HEAD
	"time"
=======
>>>>>>> af017e4 (Revert /relayer subdirectory)

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/db"
)

var DefaultConfigSet = ConfigSet{
<<<<<<< HEAD
	OCR2CachePollPeriod: 5 * time.Second,
	OCR2CacheTTL:        time.Minute,
}

type ConfigSet struct {
	OCR2CachePollPeriod time.Duration
	OCR2CacheTTL        time.Duration
}

type Config interface {
	OCR2CachePollPeriod() time.Duration
	OCR2CacheTTL() time.Duration

=======
	// defaults
}

type ConfigSet struct {
	// chain configuration parameters
}

type Config interface {
	// interface to interact with chain configuration
>>>>>>> af017e4 (Revert /relayer subdirectory)
	Update(db.ChainCfg)
}

var _ Config = (*config)(nil)

type config struct {
<<<<<<< HEAD
	defaults  ConfigSet
	dbCfg     db.ChainCfg
	dbCfgLock sync.RWMutex
	lggr      logger.Logger
=======
	defaults ConfigSet
	chain    db.ChainCfg
	chainMu  sync.RWMutex
	lggr     logger.Logger
>>>>>>> af017e4 (Revert /relayer subdirectory)
}

func NewConfig(dbCfg db.ChainCfg, lggr logger.Logger) *config {
	return &config{
		defaults: DefaultConfigSet,
<<<<<<< HEAD
		dbCfg:    dbCfg,
=======
		chain:    dbCfg,
>>>>>>> af017e4 (Revert /relayer subdirectory)
		lggr:     lggr,
	}
}

func (c *config) Update(dbCfg db.ChainCfg) {
<<<<<<< HEAD
	c.dbCfgLock.Lock()
	c.dbCfg = dbCfg
	c.dbCfgLock.Unlock()
}

func (c *config) OCR2CachePollPeriod() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.OCR2CachePollPeriod
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.OCR2CachePollPeriod
}

func (c *config) OCR2CacheTTL() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.OCR2CacheTTL
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.OCR2CacheTTL
=======
	c.chainMu.Lock()
	c.chain = dbCfg
	c.chainMu.Unlock()
>>>>>>> af017e4 (Revert /relayer subdirectory)
}
