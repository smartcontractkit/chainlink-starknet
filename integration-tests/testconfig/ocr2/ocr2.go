package ocr2

import (
	"errors"
)

type Config struct {
	NumberOfRounds *int    `toml:"number_of_rounds"`
	NodeCount      *int    `toml:"node_count"`
	TestDuration   *string `toml:"test_duration"`
}

func (o *Config) Validate() error {
	if o.NodeCount != nil && *o.NodeCount < 3 {
		return errors.New("node_count must be set and cannot be less than 3")
	}

	if o.TestDuration == nil {
		return errors.New("test_duration must be set")
	}

	if o.NumberOfRounds == nil {
		return errors.New("number_of_rounds must be set")
	}

	return nil
}
