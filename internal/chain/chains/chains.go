// Package chains is the registry wiring chain slugs to their adapters and
// per-chain constants. Offline consumers (snapshot) use DeltasFor and
// BaseUnits without constructing node clients.
package chains

import (
	"encoding/json"
	"fmt"

	"github.com/hearthchain/burning-page/internal/chain"
	"github.com/hearthchain/burning-page/internal/chain/waves"
)

// DeltaFunc replays raw history rows into signed balance changes.
type DeltaFunc func(txs []json.RawMessage, addr string) ([]chain.Delta, chain.Status)

// wavesBaseUnits is wavelets per WAVES (8 decimals).
const wavesBaseUnits = 100_000_000

// DeltasFor returns the pure delta-replay rule of the named chain.
func DeltasFor(name string) (DeltaFunc, error) {
	switch name {
	case "waves":
		return waves.Deltas, nil
	default:
		return nil, fmt.Errorf("chains: unknown chain %q", name)
	}
}

// BaseUnits returns how many base units make one whole coin on the named
// chain (wavelets per WAVES, 10^4 base units per A).
func BaseUnits(name string) (uint64, error) {
	switch name {
	case "waves":
		return wavesBaseUnits, nil
	default:
		return 0, fmt.Errorf("chains: unknown chain %q", name)
	}
}
