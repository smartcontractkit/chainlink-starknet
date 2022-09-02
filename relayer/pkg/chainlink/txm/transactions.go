package txm

import (
	"fmt"
	"sync"

	"github.com/dontpanicdao/caigo/types"
	"github.com/google/uuid"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/core"
)

type transaction struct {
	tx         types.Transaction
	id         string
	status     core.Status
	err        string
	retryCount uint
}

var _ core.Tx[types.Transaction] = transaction{}

func (tx transaction) Sender() string {
	return tx.tx.SenderAddress
}

func (tx transaction) ID() string {
	return tx.id
}

func (tx transaction) Tx() types.Transaction {
	return tx.tx
}

func (tx transaction) Hash() string {
	return tx.tx.TransactionHash
}

func (tx transaction) Status() core.Status {
	return tx.status
}

func (tx transaction) Err() string {
	return tx.err
}

// struct for managing transaction states
type txStatuses struct {
	txs  map[string]transaction
	lock sync.RWMutex
}

var _ core.TxStatuses[types.Transaction] = &txStatuses{}

func (t *txStatuses) Get(state core.Status) (out []core.Tx[types.Transaction]) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for k := range t.txs {
		if t.txs[k].status == state {
			out = append(out, t.txs[k])
		}
	}
	return out
}

func (t *txStatuses) Exists(id string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	_, exists := t.txs[id]
	return exists
}

func (t *txStatuses) Queued(tx types.Transaction) (core.Tx[types.Transaction], error) {
	id := uuid.New().String()

	// validate it exists
	if t.Exists(id) {
		return transaction{}, fmt.Errorf("transaction with id: %s already exists", id)
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.txs[id] = transaction{
		tx:     tx,
		id:     id,
		status: core.QUEUED,
	}
	return t.txs[id], nil
}

func (t *txStatuses) Retry(id string) (core.Tx[types.Transaction], error) {
	if !t.Exists(id) {
		return transaction{}, fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != core.ERRORED {
		return transaction{}, fmt.Errorf("tx must have errored before retry")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.tx.Nonce = ""  // clear if retrying
	tx.tx.MaxFee = "" // clear if retrying
	tx.status = core.RETRY
	tx.retryCount += 1
	t.txs[id] = tx
	return t.txs[id], nil
}

func (t *txStatuses) Broadcast(id, txhash string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != core.QUEUED && t.txs[id].status != core.RETRY {
		return fmt.Errorf("tx must be queued before broadcast")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.tx.TransactionHash = txhash
	tx.status = core.BROADCAST
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Confirmed(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != core.BROADCAST {
		return fmt.Errorf("tx must be broadcast before confirmed")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.status = core.CONFIRMED
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Errored(id, err string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.status = core.ERRORED
	tx.err = err
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Fatal(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != core.ERRORED {
		return fmt.Errorf("tx must have errored before fatal")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.status = core.FATAL
	t.txs[id] = tx
	return nil
}