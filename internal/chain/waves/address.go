// Package waves implements the Waves mainnet chain adapter: burn address
// construction, node access, burn detection, balance deltas and cross-checks.
package waves

import (
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

const (
	addressBodySize = 22 // version byte + scheme byte + 20 bytes of "public key hash"
	addressSize     = 26 // body + 4-byte checksum
)

// Unspendable constructs a provably unspendable Waves address: the body is the
// given motto padded with a filler character, and only the trailing 4-byte
// checksum is computed honestly. Spending would require a key whose SecureHash
// equals the chosen text, so no keypair for the address can exist. Only the
// low-order checksum bytes differ from the padded template (4 bytes < 58^6),
// therefore the motto prefix survives re-encoding.
func Unspendable(scheme byte, motto string) (string, error) {
	for _, filler := range []byte("XxKkQqVvWwYyZz123456789") {
		for padLen := len(motto); padLen <= 40; padLen++ {
			addr, ok := unspendableCandidate(scheme, motto, filler, padLen)
			if ok {
				return addr, nil
			}
		}
	}
	return "", fmt.Errorf("waves: no unspendable address found for motto %q", motto)
}

func unspendableCandidate(scheme byte, motto string, filler byte, padLen int) (string, bool) {
	template := motto + strings.Repeat(string(filler), padLen-len(motto))
	raw, err := base58.Decode(template)
	if err != nil || len(raw) != addressSize {
		return "", false
	}
	body := raw[:addressBodySize]
	if body[0] != 0x01 || body[1] != scheme {
		return "", false
	}
	checksum, err := crypto.SecureHash(body)
	if err != nil {
		return "", false
	}
	addr := base58.Encode(append(body, checksum[:4]...))
	if !strings.HasPrefix(addr, motto) {
		return "", false
	}
	parsed, err := proto.NewAddressFromString(addr)
	if err != nil {
		return "", false
	}
	valid, err := parsed.Valid(scheme)
	if err != nil || !valid {
		return "", false
	}
	return addr, true
}
