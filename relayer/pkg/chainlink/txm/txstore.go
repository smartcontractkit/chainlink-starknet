package txm

import (
	"fmt"
	"math/big"
	"sync"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"golang.org/x/exp/maps"
)

// TxStore tracks broadcast & unconfirmed txs
type TxStore struct {
	lock         sync.RWMutex
	nonceToHash  map[*big.Int]string // map nonce to txhash
	hashToNonce  map[string]*big.Int // map hash to nonce
	currentNonce *big.Int            // minimum nonce
}

func NewTxStore(current *big.Int) *TxStore {
	return &TxStore{
		nonceToHash:  map[*big.Int]string{},
		hashToNonce:  map[string]*big.Int{},
		currentNonce: current,
	}
}

func (s *TxStore) Save(nonce *big.Int, hash string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.currentNonce.Cmp(nonce) == 1 {
		return fmt.Errorf("nonce too low: %s < %s (lowest)", nonce, s.currentNonce)
	}
	if h, exists := s.nonceToHash[nonce]; exists {
		return fmt.Errorf("nonce used: tried to use nonce (%s) for tx (%s), already used by (%s)", nonce, hash, h)
	}

	// store hash
	s.nonceToHash[nonce] = hash
	s.hashToNonce[hash] = nonce

	// find next unused nonce
	_, exists := s.nonceToHash[s.currentNonce]
	for exists {
		s.currentNonce.Add(s.currentNonce, big.NewInt(1))
		_, exists = s.nonceToHash[s.currentNonce]
	}
	return nil
}

func (s *TxStore) Confirm(hash string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if nonce, exists := s.hashToNonce[hash]; exists {
		delete(s.hashToNonce, hash)
		delete(s.nonceToHash, nonce)
	}
	return fmt.Errorf("tx hash does not exist: %s", hash)
}

func (s *TxStore) GetUnconfirmed() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return maps.Values(s.nonceToHash)
}

func (s *TxStore) InflightCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.nonceToHash)
}

type ChainTxStore map[caigotypes.Hash]*TxStore

func (c ChainTxStore) Save(from caigotypes.Hash, nonce *big.Int, hash string) error {
	if err := c.validate(from); err != nil {
		// if does not exist, create a new store for the address
		c[from] = NewTxStore(nonce)
	}
	return c[from].Save(nonce, hash)
}

func (c ChainTxStore) Confirm(from caigotypes.Hash, hash string) error {
	if err := c.validate(from); err != nil {
		return err
	}
	return c[from].Confirm(hash)
}

func (c ChainTxStore) InflightCount(from caigotypes.Hash) (int, error) {
	if err := c.validate(from); err != nil {
		return 0, err
	}
	return c[from].InflightCount(), nil
}

func (c ChainTxStore) GetUnconfirmed() map[caigotypes.Hash][]string {
	list := map[caigotypes.Hash][]string{}

	for i := range c {
		list[i] = c[i].GetUnconfirmed()
	}
	return list
}

func (c ChainTxStore) validate(from caigotypes.Hash) error {
	if _, exists := c[from]; !exists {
		return fmt.Errorf("from address does not exist on ChainTxStore: %s", from)
	}
	return nil
}