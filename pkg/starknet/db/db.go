package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	// "github.com/smartcontractkit/chainlink/core/chains"
)

// todo: uncomment when interface released
// var _ chains.Config = (*ChainCfg)(nil)

type ChainCfg struct {
	// chain configuration values in db
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
