package starknet

import (
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/relay/starknet/db"
)

var DefaultConfigSet = ConfigSet{
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

	Update(db.ChainCfg)
}

var _ Config = (*config)(nil)

type config struct {
	defaults  ConfigSet
	dbCfg     db.ChainCfg
	dbCfgLock sync.RWMutex
	lggr      logger.Logger
}

func NewConfig(dbCfg db.ChainCfg, lggr logger.Logger) *config {
	return &config{
		defaults: DefaultConfigSet,
		dbCfg:    dbCfg,
		lggr:     lggr,
	}
}

func (c *config) Update(dbCfg db.ChainCfg) {
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
}
