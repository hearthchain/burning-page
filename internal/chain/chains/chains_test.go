package chains_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hearthchain/burning-page/internal/chain"
	"github.com/hearthchain/burning-page/internal/chain/chains"
)

func TestDeltasForReplaysWavesHistory(t *testing.T) {
	deltasFor, err := chains.DeltasFor("waves")
	require.NoError(t, err)

	txs := []json.RawMessage{json.RawMessage(
		`{"type":4,"id":"In1","sender":"3POther","recipient":"3PSenderAlice1111111111111111111111","assetId":null,"amount":200000000,"fee":100000,"feeAssetId":null,"timestamp":1753900000000,"height":4000001}`,
	)}
	deltas, status := deltasFor(txs, "3PSenderAlice1111111111111111111111")
	assert.Equal(t, chain.StatusOK, status.Kind)
	require.Len(t, deltas, 1)
	assert.Equal(t, int64(200000000), deltas[0].Amount)
}

func TestDeltasForRejectsUnknownChain(t *testing.T) {
	_, err := chains.DeltasFor("dogecoin")
	assert.ErrorContains(t, err, "dogecoin")
}

func TestBaseUnitsPerChain(t *testing.T) {
	units, err := chains.BaseUnits("waves")
	require.NoError(t, err)
	assert.Equal(t, uint64(100_000_000), units)

	_, err = chains.BaseUnits("dogecoin")
	assert.ErrorContains(t, err, "dogecoin")
}
