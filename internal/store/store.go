// Package store reads and writes the append-only JSONL artifacts that are
// both the backend's working state and the published reproducibility bundle.
// There is no database: artifacts plus in-memory indexes are the design.
package store

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	dirPerm      = 0o750
	filePerm     = 0o600
	maxLineBytes = 64 << 20 // a transfers page line never comes close; generous safety bound
)

// AppendJSONL marshals v and appends it as one line, creating parents as needed.
func AppendJSONL(path string, v any) error {
	if mkErr := os.MkdirAll(filepath.Dir(path), dirPerm); mkErr != nil {
		return fmt.Errorf("store: %w", mkErr)
	}
	line, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm) //nolint:gosec // artifact paths come from our own config
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer f.Close() //nolint:errcheck // double close after the checked one below is harmless
	if _, wErr := f.Write(append(line, '\n')); wErr != nil {
		return fmt.Errorf("store: %w", wErr)
	}
	return f.Close()
}

// ReadJSONL reads every line of a JSONL file into T. A missing file is an
// empty history, not an error: the watcher starts from nothing.
func ReadJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path) //nolint:gosec // artifact paths come from our own config
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only file

	var out []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineBytes)
	for sc.Scan() {
		var v T
		if uErr := json.Unmarshal(sc.Bytes(), &v); uErr != nil {
			return nil, fmt.Errorf("store: %s: %w", path, uErr)
		}
		out = append(out, v)
	}
	if scErr := sc.Err(); scErr != nil {
		return nil, fmt.Errorf("store: %s: %w", path, scErr)
	}
	return out, nil
}

// Sha256File returns the hex sha256 of a file, for artifact checksums.
func Sha256File(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // artifact paths come from our own config
	if err != nil {
		return "", fmt.Errorf("store: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only file
	h := sha256.New()
	if _, cpErr := io.Copy(h, f); cpErr != nil {
		return "", fmt.Errorf("store: %w", cpErr)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
