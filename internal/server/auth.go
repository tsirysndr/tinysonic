package server

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// Check returns true when the given Subsonic-style query params authenticate.
// Either (t, s) token-style or p plaintext (with optional "enc:" hex prefix).
func authCheck(username, password, u, p, t, s string) bool {
	if u == "" || u != username {
		return false
	}
	if t != "" && s != "" {
		sum := md5.Sum([]byte(password + s))
		return strings.EqualFold(t, hex.EncodeToString(sum[:]))
	}
	if p != "" {
		decoded := p
		if rest, ok := strings.CutPrefix(p, "enc:"); ok {
			if b, err := hex.DecodeString(rest); err == nil {
				decoded = string(b)
			}
		}
		return decoded == password
	}
	return false
}
