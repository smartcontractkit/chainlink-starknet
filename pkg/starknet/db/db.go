package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ChainCfg struct {
	// chain configuration values in db
}

func (c *ChainCfg) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("config is not []byte")
	}

	return json.Unmarshal(b, c)
}

func (c ChainCfg) Value() (driver.Value, error) {
	return json.Marshal(c)
}
