package encoding

import (
	"fmt"
)

func (l *LatestConfigDetails) Unmarshal(results []string) error {
	if len(results) != 3 {
		return fmt.Errorf("expected results size to be 3 but got %v", results)
	}
	var err error
	var index int
	if l.ConfigCount, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigCount: %w", err)
	}
	index++
	if l.BlockNumber, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding BlockNumber: %w", err)
	}
	index++
	if l.ConfigDigest, err = DecodeConfigDigest(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigDigest: %w", err)
	}
	return nil
}

func (l *LatestTransmissionDetails) Unmarshal(results []string) error {
	if len(results) != 4 {
		return fmt.Errorf("expected results size to be 4 but got %v", results)
	}
	var err error
	var index int
	if l.ConfigDigest, err = DecodeConfigDigest(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigCount: %w", err)
	}
	index++
	if l.Epoch, l.Round, err = DecodeEpochAndRound(results[index]); err != nil {
		return fmt.Errorf("error decoding Epoch and Round: %w", err)
	}
	index++
	if l.LatestAnswer, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding LatestAnswer: %w", err)
	}
	index++
	if l.LatestTimestamp, err = DecodeTime(results[index]); err != nil {
		return fmt.Errorf("error decoding LatestTimestamp: %w", err)
	}
	return nil
}

func (r *RoundData) Unmarshal(results []string) error {
	if len(results) != 5 {
		return fmt.Errorf("expected results size to be 5 but got %v", results)
	}
	var err error
	var index int
	if r.RoundID, err = DecodeUint32(results[index]); err != nil {
		return fmt.Errorf("error decoding RoundID: %w", err)
	}
	index++
	if r.Answer, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding Answer: %w", err)
	}
	index++
	if r.BlockNumber, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding BlockNumber: %w", err)
	}
	index++
	if r.StartedAt, err = DecodeTime(results[index]); err != nil {
		return fmt.Errorf("error decoding StartedAt: %w", err)
	}
	index++
	if r.UpdatedAt, err = DecodeTime(results[index]); err != nil {
		return fmt.Errorf("error decoding UpdatedAt: %w", err)
	}
	return nil
}

func (l *LinkAvailableForPayment) Unmarshal(results []string) error {
	if len(results) != 1 {
		return fmt.Errorf("expected results size to be 1 but got %v", results)
	}
	var err error
	var index int
	if l.Available, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding Available: %w", err)
	}
	return nil
}

func (n *NewTransmisisonEvent) Unmarshal(results []string) error {
	if len(results) < 10 {
		return fmt.Errorf("expected results size to be at least 10 but got %v", results)
	}
	var err error
	var index int
	if n.RoundID, err = DecodeUint32(results[index]); err != nil {
		return fmt.Errorf("error decoding RoundID: %w", err)
	}
	index++
	if n.Answer, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding Answer: %w", err)
	}
	index++
	if n.Transmitter, err = DecodeAccount(results[index]); err != nil {
		return fmt.Errorf("error decoding Transmitter: %w", err)
	}
	index++
	if n.ObservationTimestamp, err = DecodeTime(results[index]); err != nil {
		return fmt.Errorf("error decoding ObservationTimestamp: %w", err)
	}
	index++
	if n.Observers, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding Observers: %w", err)
	}
	index++
	if n.Observations, err = DecodeObservations(results, index); err != nil {
		return fmt.Errorf("error decoding Observations: %w", err)
	}
	index += len(n.Observations) + 1
	if n.JuelsPerFeeCoin, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding JuelsPerFeeCoin: %w", err)
	}
	index++
	if n.ConfigDigest, err = DecodeConfigDigest(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigDigest: %w", err)
	}
	index++
	if n.Epoch, n.Round, err = DecodeEpochAndRound(results[index]); err != nil {
		return fmt.Errorf("error decoding Epoch and Round: %w", err)
	}
	index++
	if n.Reimbursement, err = DecodeBigInt(results[index]); err != nil {
		return fmt.Errorf("error decoding Reimbursement: %w", err)
	}
	return nil
}

func (c *ConfigSetEvent) Unmarshal(results []string) error {
	if len(results) < 10 {
		return fmt.Errorf("expected results size to be at least 10 but got %v", results)
	}
	var err error
	var index int
	var numSlots uint64
	if c.PreviousConfigBlockNumber, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding PreviousConfigBlockNumber: %w", err)
	}
	index++
	if c.LatestConfigDigest, err = DecodeConfigDigest(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigDigest: %w", err)
	}
	index++
	if c.ConfigCount, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding ConfigCount: %w", err)
	}
	index++
	if c.Signers, c.Transmitters, err = DecodeOracles(results, index); err != nil {
		return fmt.Errorf("error decoding Signers and Transmitters: %w", err)
	}
	index += len(c.Signers) + len(c.Transmitters) + 1
	if c.F, err = DecodeUint8(results[index]); err != nil {
		return fmt.Errorf("error decoding F: %w", err)
	}
	index += 1
	c.OnchainConfig, numSlots, err = DecodeOnchainConfig(results, index)
	if err != nil {
		return fmt.Errorf("error decoding OnchainConfig: %w", err)
	}
	index += int(numSlots) + 1
	if c.OffchainConfigVersion, err = DecodeUint64(results[index]); err != nil {
		return fmt.Errorf("error decoding OffchainConfigVersion: %w", err)
	}
	c.OffchainConfig, numSlots, err = DecodeOffchainConfig(results, index)
	if err != nil {
		return fmt.Errorf("error decoding OnchainConfig: %w", err)
	}
	return nil
}
