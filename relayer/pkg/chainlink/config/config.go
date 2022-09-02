package config

import (
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/db"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/core"
)

var DefaultConfigSet = ConfigSet{
	OCR2CachePollPeriod: 5 * time.Second,
	OCR2CacheTTL:        time.Minute,
	RequestTimeout:      10 * time.Second,
	TxTimeout:           time.Minute,
	TxConfirmFrequency:  5 * time.Second,
	TxRetryFrequency:    5 * time.Second,
}

type ConfigSet struct {
	OCR2CachePollPeriod time.Duration
	OCR2CacheTTL        time.Duration

	// client config
	RequestTimeout time.Duration

	// txm config
	TxTimeout          time.Duration
	TxConfirmFrequency time.Duration
	TxRetryFrequency   time.Duration
}

type Config interface {
	core.Config // txm config

	// ocr2 config
	ocr2.Config

	// client config
	RequestTimeout() time.Duration

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

func (c *config) TxConfirmFrequency() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.TxConfirmFrequency
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.TxConfirmFrequency
}

func (c *config) TxRetryFrequency() time.Duration {
	c.dbCfgLock.RLock()
	ch := c.dbCfg.TxRetryFrequency
	c.dbCfgLock.RUnlock()
	if ch != nil {
		return ch.Duration()
	}
	return c.defaults.TxRetryFrequency
}
