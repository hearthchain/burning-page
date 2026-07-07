// Package hearthaddr validates Hearth addresses. The format is provisional
// until the chain fork fixes it: the Waves 26-byte layout (version 0x01,
// scheme byte, 20-byte key hash, 4-byte SecureHash checksum) under the Hearth
// scheme byte. Keeping it in one small package means the final format decision
// touches only this file.
package hearthaddr

import (
	"fmt"

	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

// New derives a Hearth address from a public key under the given scheme byte.
func New(scheme byte, pub crypto.PublicKey) (string, error) {
	addr, err := proto.NewAddressFromPublicKey(scheme, pub)
	if err != nil {
		return "", fmt.Errorf("hearthaddr: %w", err)
	}
	return addr.String(), nil
}

// Validate checks length, version, scheme byte and checksum of a Hearth address.
func Validate(addr string, scheme byte) error {
	parsed, err := proto.NewAddressFromString(addr)
	if err != nil {
		return fmt.Errorf("hearthaddr: %w", err)
	}
	valid, err := parsed.Valid(scheme)
	if err != nil {
		return fmt.Errorf("hearthaddr: %w", err)
	}
	if !valid {
		return fmt.Errorf("hearthaddr: invalid address for scheme %q", scheme)
	}
	return nil
}
