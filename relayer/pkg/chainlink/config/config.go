package config

import (
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/utils"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm"
)

var DefaultConfigSet = ConfigSet{
	OCR2CachePollPeriod: 5 * time.Second,
	OCR2CacheTTL:        time.Minute,
	RequestTimeout:      10 * time.Second,
	TxTimeout:           10 * time.Second,
	ConfirmationPoll:    5 * time.Second,
}

type ConfigSet struct {
	OCR2CachePollPeriod time.Duration
	OCR2CacheTTL        time.Duration

	// client config
	RequestTimeout time.Duration

	// txm config
	TxTimeout        time.Duration
	ConfirmationPoll time.Duration
}

type Config interface {
	txm.Config // txm config

	// ocr2 config
	ocr2.Config

	// client config
	RequestTimeout() time.Duration
}

type Chain struct {
	OCR2CachePollPeriod *utils.Duration
	OCR2CacheTTL        *utils.Duration
	RequestTimeout      *utils.Duration
	TxTimeout           *utils.Duration
	ConfirmationPoll    *utils.Duration
}

func (c *Chain) SetDefaults() {
	if c.OCR2CachePollPeriod == nil {
		c.OCR2CachePollPeriod = utils.MustNewDuration(DefaultConfigSet.OCR2CachePollPeriod)
	}
	if c.OCR2CacheTTL == nil {
		c.OCR2CacheTTL = utils.MustNewDuration(DefaultConfigSet.OCR2CacheTTL)
	}
	if c.RequestTimeout == nil {
		c.RequestTimeout = utils.MustNewDuration(DefaultConfigSet.RequestTimeout)
	}
	if c.TxTimeout == nil {
		c.TxTimeout = utils.MustNewDuration(DefaultConfigSet.TxTimeout)
	}
	if c.ConfirmationPoll == nil {
		c.ConfirmationPoll = utils.MustNewDuration(DefaultConfigSet.ConfirmationPoll)
	}
}

type Node struct {
	Name *string
	URL  *utils.URL
}
