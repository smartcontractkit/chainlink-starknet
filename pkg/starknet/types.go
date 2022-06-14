<<<<<<< HEAD
package starknet

type CallOps struct {
	ContractAddress string
	Selector        string
	Calldata        []string
=======
package relay

// [relayConfig] member of Chainlink's job spec v2 (OCR2 only currently)
type RelayConfig struct {
	ChainID  string `json:"chainID"`
	NodeName string `json:"nodeName"` // optional, defaults to random node with 'chainID'
>>>>>>> af017e4 (Revert /relayer subdirectory)
}
