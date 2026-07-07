// Package api serves the three read/submit endpoints of the burn backend:
// a live credit preview by source address, the cabinet balance by Hearth
// address, and binding submission. The server is never authoritative: every
// number it returns is recomputable from the published artifacts.
package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/wavesplatform/gowaves/pkg/proto"

	"github.com/hearthchain/burning-page/internal/binding"
	"github.com/hearthchain/burning-page/internal/bindings"
	"github.com/hearthchain/burning-page/internal/chain/waves"
	"github.com/hearthchain/burning-page/internal/config"
	"github.com/hearthchain/burning-page/internal/credit"
	"github.com/hearthchain/burning-page/internal/hearthaddr"
	"github.com/hearthchain/burning-page/internal/journal"
	"github.com/hearthchain/burning-page/internal/layers"
	"github.com/hearthchain/burning-page/internal/snapshot"
)

const (
	maxPreviewConcurrency = 4 // preview is the only endpoint spending public-node quota
	maxBindingBodyBytes   = 4 << 10
	microPerCredit        = 1_000_000
)

// Node is the read surface the preview endpoint needs.
type Node interface {
	AllTransactions(ctx context.Context, addr string) ([]json.RawMessage, error)
}

// Server wires the endpoints to their dependencies.
type Server struct {
	node       Node
	journal    *journal.Journal
	registry   *bindings.Registry
	cfg        config.Config
	previewSem chan struct{}
}

// New builds a Server.
func New(node Node, j *journal.Journal, reg *bindings.Registry, cfg config.Config) *Server {
	return &Server{
		node:       node,
		journal:    j,
		registry:   reg,
		cfg:        cfg,
		previewSem: make(chan struct{}, maxPreviewConcurrency),
	}
}

//go:embed bind.html
var bindPage []byte

// Handler returns the HTTP handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/preview/waves/{address}", s.preview)
	mux.HandleFunc("GET /api/address/{hearth}", s.address)
	mux.HandleFunc("POST /api/bindings", s.postBinding)
	mux.HandleFunc("GET /bind", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(bindPage)
	})
	return mux
}

func (s *Server) preview(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("address")
	parsed, err := proto.NewAddressFromString(addr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_address", err.Error())
		return
	}
	if ok, vErr := parsed.Valid(proto.MainNetScheme); vErr != nil || !ok {
		writeError(w, http.StatusBadRequest, "invalid_address", "not a Waves mainnet address")
		return
	}
	select {
	case s.previewSem <- struct{}{}:
		defer func() { <-s.previewSem }()
	default:
		writeError(w, http.StatusTooManyRequests, "busy", "too many concurrent previews")
		return
	}
	txs, err := s.node.AllTransactions(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusBadGateway, "node_error", err.Error())
		return
	}
	deltas, status := waves.Deltas(txs, addr)
	if status.Kind != "ok" {
		writeError(w, http.StatusUnprocessableEntity, "unsupported_history", status.Reason+"; manual review")
		return
	}
	profile, _, err := layers.Build(deltas, nil)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "unsupported_history", err.Error())
		return
	}
	total, perLayer, err := credit.Compute(profile, s.journal)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "unsupported_history", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"address":          addr,
		"status":           "ok",
		"layers":           perLayer,
		"totalCreditMicro": total.String(),
		"totalCredit":      microToDecimal(total),
	})
}

func (s *Server) address(w http.ResponseWriter, r *http.Request) {
	hearth := r.PathValue("hearth")
	if err := hearthaddr.Validate(hearth, s.cfg.HearthSchemeByte()); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_address", err.Error())
		return
	}
	_, bundles, err := snapshot.Build(s.cfg.DataDir, s.journal, s.cfg.HearthSchemeByte())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "artifacts_error", err.Error())
		return
	}
	total := new(big.Int)
	burns := []any{}
	for _, b := range bundles {
		if b.Hearth != hearth {
			continue
		}
		burns = append(burns, b)
		const base = 10
		c, ok := new(big.Int).SetString(b.CreditMicro, base)
		if ok {
			total.Add(total, c)
		}
	}
	sources := s.registry.SourcesFor(hearth)
	if sources == nil {
		sources = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"hearthAddress":    hearth,
		"totalCreditMicro": total.String(),
		"totalCredit":      microToDecimal(total),
		"bindings":         sources,
		"burns":            burns,
	})
}

func (s *Server) postBinding(w http.ResponseWriter, r *http.Request) {
	var rec bindings.Record
	body := http.MaxBytesReader(w, r.Body, maxBindingBodyBytes)
	if err := json.NewDecoder(body).Decode(&rec); err != nil {
		writeError(w, http.StatusBadRequest, "malformed", err.Error())
		return
	}
	rec.Chain = "waves"
	if err := s.registry.Add(rec); err != nil {
		writeBindingError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"accepted": true})
}

func writeBindingError(w http.ResponseWriter, err error) {
	if errors.Is(err, binding.ErrBadSignature) || errors.Is(err, binding.ErrSourceMismatch) {
		writeError(w, http.StatusUnauthorized, "invalid_signature", err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_binding", err.Error())
}

// microToDecimal renders micro-HRTH as a decimal credit string, e.g.
// 49713174000 -> "49713.174000".
func microToDecimal(micro *big.Int) string {
	quo, rem := new(big.Int).QuoRem(micro, big.NewInt(microPerCredit), new(big.Int))
	return fmt.Sprintf("%s.%06d", quo.String(), rem.Int64())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
