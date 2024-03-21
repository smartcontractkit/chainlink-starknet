package txm

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/NethermindEth/juno/core/felt"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	"golang.org/x/exp/maps"
)

// TxStore tracks broadcast & unconfirmed txs per account address per chain id
type TxStore struct {
	lock        sync.RWMutex
	nonceToHash map[felt.Felt]string // map nonce to txhash
	hashToNonce map[string]felt.Felt // map hash to nonce
	hashToCalls map[string][]starknetrpc.FunctionCall
	hashToKey   map[string]felt.Felt
}

func NewTxStore() *TxStore {
	return &TxStore{
		nonceToHash: map[felt.Felt]string{},
		hashToNonce: map[string]felt.Felt{},
		hashToCalls: map[string][]starknetrpc.FunctionCall{},
		hashToKey:   map[string]felt.Felt{},
	}
}

func deepCopy(nonce *felt.Felt, calls []starknetrpc.FunctionCall, publicKey *felt.Felt) (newNonce *felt.Felt, newCalls []starknetrpc.FunctionCall, newPublicKey *felt.Felt) {
	newNonce = new(felt.Felt).Set(nonce)
	newPublicKey = new(felt.Felt).Set(publicKey)
	// TODO: make new felts?
	newCalls = calls
	return
}

func copyCall(call *starknetrpc.FunctionCall) *starknetrpc.FunctionCall {
	copyCall := starknetrpc.FunctionCall{
		ContractAddress:    new(felt.Felt).Set(call.ContractAddress),
		EntryPointSelector: new(felt.Felt).Set(call.EntryPointSelector),
		Calldata:           []*felt.Felt{},
	}
	for i := 0; i < len(call.Calldata); i++ {
		copyCall.Calldata = append(copyCall.Calldata, new(felt.Felt).Set(call.Calldata[i]))
	}
	return &copyCall
}

func (s *TxStore) Save(nonce *felt.Felt, hash string, calls []starknetrpc.FunctionCall, publicKey *felt.Felt) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if h, exists := s.nonceToHash[*nonce]; exists {
		return fmt.Errorf("nonce used: tried to use nonce (%s) for tx (%s), already used by (%s)", nonce, hash, h)
	}
	if n, exists := s.hashToNonce[hash]; exists {
		return fmt.Errorf("hash used: tried to use tx (%s) for nonce (%s), already used nonce (%s)", hash, nonce, &n)
	}

	newNonce, newCalls, newPublicKey := deepCopy(nonce, calls, publicKey)

	// store hash
	s.nonceToHash[*newNonce] = hash
	s.hashToNonce[hash] = *newNonce
	s.hashToCalls[hash] = newCalls
	s.hashToKey[hash] = *newPublicKey

	return nil
}

func (s *TxStore) Confirm(hash string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if nonce, exists := s.hashToNonce[hash]; exists {
		delete(s.nonceToHash, nonce)

		delete(s.hashToNonce, hash)
		delete(s.hashToCalls, hash)
		delete(s.hashToKey, hash)
		return nil
	}
	return fmt.Errorf("tx hash does not exist - it may already be confirmed: %s", hash)
}

func (s *TxStore) GetUnconfirmed() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return maps.Values(s.nonceToHash)
}

func (s *TxStore) GetUnconfirmedNonces() []felt.Felt {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return maps.Keys(s.nonceToHash)
}

func (s *TxStore) GetSmallestUnconfirmedNonce() *felt.Felt {
	nonces := s.GetUnconfirmedNonces()
	smallest := &nonces[0]
	for i := 1; i < len(nonces); i++ {
		current := &nonces[i]
		if smallest.Cmp(current) > 0 {
			smallest = current
		}
	}
	return smallest
}

type UnconfirmedTx struct {
	PublicKey *felt.Felt
	Hash      string
	Nonce     *felt.Felt
	Calls     []starknetrpc.FunctionCall
}

func (s *TxStore) GetSingleUnconfirmed(hash string) (tx UnconfirmedTx, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	n, hExists := s.hashToNonce[hash]
	c, cExists := s.hashToCalls[hash]
	k, kExists := s.hashToKey[hash]

	if !hExists || !cExists || !kExists {
		return tx, errors.New("datum not found in txstore")
	}

	newNonce, newCalls, newPublicKey := deepCopy(&n, c, &k)

	tx.Calls = newCalls
	tx.Nonce = newNonce
	tx.PublicKey = newPublicKey
	tx.Hash = hash

	return tx, nil
}

// Retrieve Unconfirmed Txs in their queued order (by nonce)
func (s *TxStore) GetUnconfirmedSorted() (txs []UnconfirmedTx) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	nonces := maps.Values(s.hashToNonce)
	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i].Cmp(&nonces[j]) == -1
	})

	for i := 0; i < len(nonces); i++ {
		n := nonces[i]
		h := s.nonceToHash[n]
		k := s.hashToKey[h]
		c := s.hashToCalls[h]

		newNonce, newCalls, newPublicKey := deepCopy(&n, c, &k)

		txs = append(txs, UnconfirmedTx{Hash: h, Nonce: newNonce, Calls: newCalls, PublicKey: newPublicKey})
	}

	return txs
}

func (s *TxStore) InflightCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.nonceToHash)
}

type ChainTxStore struct {
	store map[*felt.Felt]*TxStore // map account address to txstore
	lock  sync.RWMutex
}

func NewChainTxStore() *ChainTxStore {
	return &ChainTxStore{
		store: map[*felt.Felt]*TxStore{},
	}
}

func (c *ChainTxStore) Save(from *felt.Felt, nonce *felt.Felt, hash string, calls []starknetrpc.FunctionCall, publicKey *felt.Felt) error {
	// use write lock for methods that modify underlying data
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.validate(from); err != nil {
		// if does not exist, create a new store for the address
		c.store[from] = NewTxStore()
	}
	return c.store[from].Save(nonce, hash, calls, publicKey)
}

func (c *ChainTxStore) Confirm(from *felt.Felt, hash string) error {
	// use write lock for methods that modify underlying data
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.validate(from); err != nil {
		return err
	}
	return c.store[from].Confirm(hash)
}

func (c *ChainTxStore) GetUnconfirmedSorted(from *felt.Felt) []UnconfirmedTx {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// empty slice if address isn't found
	if err := c.validate(from); err != nil {
		return nil
	}

	return c.store[from].GetUnconfirmedSorted()
}

func (c *ChainTxStore) GetSmallestUnconfirmedNonce(from *felt.Felt) *felt.Felt {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// empty slice if address isn't found
	if err := c.validate(from); err != nil {
		return nil
	}

	return c.store[from].GetSmallestUnconfirmedNonce()
}

func (c *ChainTxStore) GetSingleUnconfirmed(from *felt.Felt, hash string) (tx UnconfirmedTx, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if err := c.validate(from); err != nil {
		return tx, err
	}

	return c.store[from].GetSingleUnconfirmed(hash)
}

func (c *ChainTxStore) GetAllInflightCount() map[*felt.Felt]int {
	// use read lock for methods that read underlying data
	c.lock.RLock()
	defer c.lock.RUnlock()

	list := map[*felt.Felt]int{}

	for i := range c.store {
		list[i] = c.store[i].InflightCount()
	}

	return list
}

func (c *ChainTxStore) GetAllUnconfirmed() map[*felt.Felt][]string {
	// use read lock for methods that read underlying data
	c.lock.RLock()
	defer c.lock.RUnlock()

	list := map[*felt.Felt][]string{}

	for i := range c.store {
		list[i] = c.store[i].GetUnconfirmed()
	}
	return list
}

func (c *ChainTxStore) validate(from *felt.Felt) error {
	if _, exists := c.store[from]; !exists {
		return fmt.Errorf("from address does not exist: %s", from)
	}
	return nil
}
