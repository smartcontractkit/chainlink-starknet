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
	hashToCall  map[string]*starknetrpc.FunctionCall
	hashToKey   map[string]felt.Felt
}

func NewTxStore() *TxStore {
	return &TxStore{
		nonceToHash: map[felt.Felt]string{},
		hashToNonce: map[string]felt.Felt{},
		hashToCall:  map[string]*starknetrpc.FunctionCall{},
		hashToKey:   map[string]felt.Felt{},
	}
}

func deepCopy(nonce *felt.Felt, call *starknetrpc.FunctionCall, publicKey *felt.Felt) (newNonce *felt.Felt, newCall *starknetrpc.FunctionCall, newPublicKey *felt.Felt) {
	newNonce = new(felt.Felt).Set(nonce)
	newPublicKey = new(felt.Felt).Set(publicKey)
	newCall = copyCall(call)
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

func (s *TxStore) Save(nonce *felt.Felt, hash string, call *starknetrpc.FunctionCall, publicKey *felt.Felt) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if h, exists := s.nonceToHash[*nonce]; exists {
		return fmt.Errorf("nonce used: tried to use nonce (%s) for tx (%s), already used by (%s)", nonce, hash, h)
	}
	if n, exists := s.hashToNonce[hash]; exists {
		return fmt.Errorf("hash used: tried to use tx (%s) for nonce (%s), already used nonce (%s)", hash, nonce, &n)
	}

	newNonce, newCall, newPublicKey := deepCopy(nonce, call, publicKey)

	// store hash
	s.nonceToHash[*newNonce] = hash

	s.hashToNonce[hash] = *newNonce
	s.hashToCall[hash] = newCall
	s.hashToKey[hash] = *newPublicKey

	return nil
}

func (s *TxStore) Confirm(hash string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if nonce, exists := s.hashToNonce[hash]; exists {
		delete(s.nonceToHash, nonce)

		delete(s.hashToNonce, hash)
		delete(s.hashToCall, hash)
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

type UnconfirmedTx struct {
	PublicKey *felt.Felt
	Hash      string
	Nonce     *felt.Felt
	Call      *starknetrpc.FunctionCall
}

func (s *TxStore) GetSingleUnconfirmed(hash string) (tx UnconfirmedTx, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	n, hExists := s.hashToNonce[hash]
	c, cExists := s.hashToCall[hash]
	k, kExists := s.hashToKey[hash]

	if !hExists || !cExists || !kExists {
		return tx, errors.New("datum not found in txstore")
	}

	newNonce, newCall, newPublicKey := deepCopy(&n, c, &k)

	tx.Call = newCall
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
		c := s.hashToCall[h]

		newNonce, newCall, newPublicKey := deepCopy(&n, c, &k)

		txs = append(txs, UnconfirmedTx{Hash: h, Nonce: newNonce, Call: newCall, PublicKey: newPublicKey})
	}

	return txs
}

func (s *TxStore) InflightCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.nonceToHash)
}

type AccountStore struct {
	store map[*felt.Felt]*TxStore // map account address to txstore
	lock  sync.RWMutex
}

func NewAccountStore() *AccountStore {
	return &AccountStore{
		store: map[*felt.Felt]*TxStore{},
	}
}

// GetTxStore returns the TxStore for the provided account, creating it if it does not exist.
func (c *AccountStore) GetTxStore(accountAddress *felt.Felt) *TxStore {
	c.lock.Lock()
	defer c.lock.Unlock()
	store, ok := c.store[accountAddress]
	if !ok {
		store = NewTxStore()
		c.store[accountAddress] = store
	}
	return store
}

func (c *AccountStore) GetAllInflightCount() map[*felt.Felt]int {
	// use read lock for methods that read underlying data
	c.lock.RLock()
	defer c.lock.RUnlock()

	list := map[*felt.Felt]int{}

	for i := range c.store {
		list[i] = c.store[i].InflightCount()
	}

	return list
}

func (c *AccountStore) GetAllUnconfirmed() map[*felt.Felt][]string {
	// use read lock for methods that read underlying data
	c.lock.RLock()
	defer c.lock.RUnlock()

	list := map[*felt.Felt][]string{}

	for i := range c.store {
		list[i] = c.store[i].GetUnconfirmed()
	}
	return list
}
