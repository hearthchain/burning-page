// Package chain defines the chain-adapter port: the chain-agnostic types the
// watcher, credit engine and snapshot work with. One package per chain (waves
// first) implements the concrete detection and history extraction.
package chain

import (
	"encoding/json"
	"time"
)

// Window bounds a burn campaign in block heights, inclusive on both ends.
type Window struct {
	Start uint64 `json:"startHeight"`
	End   uint64 `json:"endHeight"`
}

// Burn is one detected burn: a transfer of the native coin to the published
// burn address, attributed to the sending address.
type Burn struct {
	TxID      string          `json:"txId"`
	Chain     string          `json:"chain"`
	Source    string          `json:"source"`
	Amount    uint64          `json:"amountBaseUnits"`
	Height    uint64          `json:"height"`
	Timestamp time.Time       `json:"timestamp"`
	Raw       json.RawMessage `json:"raw"`
}

// Delta is one signed native-coin balance change of an address.
type Delta struct {
	TxID      string    `json:"txId"`
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
}

// Status is the verdict of a delta reconstruction: Kind "ok" or "unsupported"
// (the history contains a transaction the adapter does not interpret; the
// address is blocked to manual review rather than risking a wrong credit).
type Status struct {
	Kind   string
	Reason string
}

// Verdict is the outcome of a double-source check: confirmed, mismatch (with
// the diverging field names) or pending_crosscheck while the second source has
// not yet buried the burn under enough confirmations.
type Verdict struct {
	Status     string   `json:"status"`
	Node       string   `json:"node,omitempty"`
	Mismatches []string `json:"mismatchFields,omitempty"`
}

// History is the reconstructed balance-delta history of one address together
// with the safety-invariant verdict. Status is "ok" only when the recomputed
// balance exactly matches the node-reported balance at ReferenceHeight.
type History struct {
	Address         string  `json:"address"`
	Deltas          []Delta `json:"deltas"`
	ReferenceHeight uint64  `json:"referenceHeight"`
	NodeBalance     uint64  `json:"nodeBalanceBaseUnits"`
	Recomputed      int64   `json:"recomputedBaseUnits"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason,omitempty"`
}
