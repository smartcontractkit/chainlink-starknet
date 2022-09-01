package monitoring

import (
	"bytes"
	"encoding/json"

	"github.com/dontpanicdao/caigo"
	caigogw "github.com/dontpanicdao/caigo/gateway"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func FilterEvents(block *caigogw.Block, address, eventName string) ([][]string, error) {
	eventKey := caigo.GetSelectorFromName(eventName)
	resultss := [][]string{}
	for _, txReceipt := range block.TransactionReceipts {
		for _, event := range txReceipt.Events {
			buf, err := json.Marshal(event)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't marshal event")
			}

			var decodedEvent caigotypes.Event
			if err := json.Unmarshal(buf, &decodedEvent); err != nil {
				return nil, errors.Wrap(err, "couldn't unmarshal event")
			}

			isEventFromContract := compareAddress(decodedEvent.FromAddress, address)
			isEventType := decodedEvent.Keys[0].Cmp(eventKey) == 0
			if isEventFromContract && isEventType {
				results := []string{}
				for _, felt := range decodedEvent.Data {
					results = append(results, felt.String())
				}
				resultss = append(resultss, results)
			}
		}
	}
	return resultss, nil
}

// compareAddress compares different hex starknet addresses with potentially different 0 padding
func compareAddress(a, b string) bool {
	aBytes, err := keys.HexToBytes(a)
	if err != nil {
		return false
	}

	bBytes, err := keys.HexToBytes(b)
	if err != nil {
		return false
	}

	return bytes.Compare(starknet.PadBytes(aBytes, 32), starknet.PadBytes(bBytes, 32)) == 0
}
