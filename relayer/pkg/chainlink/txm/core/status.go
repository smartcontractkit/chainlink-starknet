package core

// transaction states
type Status int

const (
	QUEUED Status = iota
	RETRY         // can be reached from ERRORED
	BROADCAST
	CONFIRMED // ending state for happy path
	ERRORED
	FATAL // ending state for failed txs (reverts, etc)
)

func (d Status) String() string {
	return [...]string{"QUEUED", "RETRY", "BROADCAST", "CONFIRMED", "FATAL"}[d]
}