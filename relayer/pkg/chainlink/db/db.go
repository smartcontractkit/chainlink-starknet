package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	// "github.com/smartcontractkit/chainlink/core/chains"
)

// todo: uncomment when interface is decoupled from core
// var _ chains.Config = (*ChainCfg)(nil)

type ChainCfg struct {
	OCR2CachePollPeriod *utils.Duration
	OCR2CacheTTL        *utils.Duration
	RequestTimeout      *utils.Duration
	TxTimeout           *utils.Duration
	TxSendFrequency     *utils.Duration
	TxMaxBatchSize      null.Int
}

func (c *ChainCfg) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, c)
}

func (c ChainCfg) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type Node struct {
	ID        int32
	Name      string
	ChainID   string `db:"starknet_chain_id"`
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
