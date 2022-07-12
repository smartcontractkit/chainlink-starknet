package starknet

import (
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet/db"
)

var DefaultConfigSet = ConfigSet{
	OCR2CachePollPeriod: 5 * time.Second,
	OCR2CacheTTL:        time.Minute,
	RequestTimeout:      10 * time.Second,
	TxTimeout:           time.Minute,
	TxSendFrequency:     15 * time.Second,
	TxMaxBatchSize:      100,
}

type ConfigSet struct {
	OCR2CachePollPeriod time.Duration
	OCR2CacheTTL        time.Duration

	// client config
	RequestTimeout time.Duration

	// txm config
	TxTimeout       time.Duration
	TxSendFrequency time.Duration
	TxMaxBatchSize  int
}

type Config interface {
	OCR2CachePollPeriod() time.Duration
	OCR2CacheTTL() time.Duration
	RequestTimeout() time.Duration
	TxTimeout() time.Duration
	TxSendFrequency() time.Duration
	TxMaxBatchSize() int

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

func (c *config) RequestTimeout() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.RequestTimeout
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.RequestTimeout
}

func (c *config) TxTimeout() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.TxTimeout
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.TxTimeout
}

func (c *config) TxSendFrequency() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.TxSendFrequency
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.TxSendFrequency
}

func (c *config) TxMaxBatchSize() int {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.TxMaxBatchSize
	c.dbCfgLock.RUnlock()
	if ch.Valid {
		return int(ch.Int64)
	}
	return c.defaults.TxMaxBatchSize
}
