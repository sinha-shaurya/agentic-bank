// Package token mints bearer tokens and derives the hash stored in their place.
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// TOKEN_BYTES is 256 bits of entropy, which is why a fast hash is safe here.
const TOKEN_BYTES = 32

// GenerateToken returns a new URL-safe bearer token. crypto/rand.Read never returns
// an error since Go 1.24; a broken entropy source crashes the process instead.
func GenerateToken() string {
	buf := make([]byte, TOKEN_BYTES)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

// Hash is the value persisted in users.token_sha256 and used for lookup.
func Hash(bearerToken string) string {
	sum := sha256.Sum256([]byte(bearerToken))
	return hex.EncodeToString(sum[:])
}
