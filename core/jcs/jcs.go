package jcs

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gowebpki/jcs"
)

// CanonicalizeJSON returns the RFC 8785 (JCS) canonical form of JSON input.
func CanonicalizeJSON(input []byte) ([]byte, error) {
	return jcs.Transform(input)
}

// DigestJCS canonicalizes JSON (RFC 8785) and returns a sha256 hex digest.
func DigestJCS(input []byte) (string, error) {
	canonical, err := CanonicalizeJSON(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
