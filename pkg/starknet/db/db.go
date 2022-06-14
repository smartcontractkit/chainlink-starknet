package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	// "github.com/smartcontractkit/chainlink/core/chains"
)

// todo: uncomment when interface is decoupled from core
// var _ chains.Config = (*ChainCfg)(nil)

type ChainCfg struct {
<<<<<<< HEAD
	OCR2CachePollPeriod *utils.Duration
	OCR2CacheTTL        *utils.Duration
=======
	OCR2CachePollPeriod   *utils.Duration
	OCR2CacheTTL          *utils.Duration
>>>>>>> af017e4 (Revert /relayer subdirectory)
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
