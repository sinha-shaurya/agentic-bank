package service

import (
	"regexp"
	"testing"
)

var urlSafe = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func TestGenerateIsURLSafeAndUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for range 1000 {
		bearerToken := GenerateToken()
		if !urlSafe.MatchString(bearerToken) {
			t.Fatalf("token %q is not 43 url-safe characters", bearerToken)
		}
		if _, dup := seen[bearerToken]; dup {
			t.Fatalf("duplicate token %q", bearerToken)
		}
		seen[bearerToken] = struct{}{}
	}
}

func TestHashIsSHA256Hex(t *testing.T) {
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := Hash("abc"); got != want {
		t.Errorf("Hash(abc) = %s, want %s", got, want)
	}
	//nolint
	if Hash("abc") != Hash("abc") {
		t.Error("Hash is not deterministic")
	}
}
