package starknet

import (
	caigo "github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"
)

func IsEventFromContract(event *caigotypes.Event, address string, eventName string) bool {
	eventKey := caigo.GetSelectorFromName(eventName)
	// encoded event name guaranteed to be at index 0
	return CompareAddress(event.FromAddress, address) && event.Keys[0].Cmp(eventKey) == 0
}
