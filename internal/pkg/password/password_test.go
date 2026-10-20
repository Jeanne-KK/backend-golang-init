package password

import (
	"strings"
	"testing"

	"backend/internal/config"
)

func testArgon2Config() config.Argon2Config {
	return config.Argon2Config{
		MemoryKiB:  1024,
		Iterations: 1,
		Threads:    1,
		SaltLength: 16,
		KeyLength:  32,
	}
}

func TestHashVerifyRoundTrip(t *testing.T) {
	h := NewHasher(testArgon2Config())
	encoded, err := h.Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, phcPrefix) {
		t.Fatalf("expected PHC prefix, got %q", encoded)
	}
	ok, err := h.Verify("correct-horse-battery-staple", encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected verify true")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	h := NewHasher(testArgon2Config())
	encoded, err := h.Hash("password-one")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := h.Verify("password-two", encoded)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected verify false")
	}
}

func TestVerifyInvalidEncoded(t *testing.T) {
	h := NewHasher(testArgon2Config())
	ok, err := h.Verify("password", "not-a-phc-string")
	if err == nil {
		t.Fatal("expected error for invalid encoded")
	}
	if ok {
		t.Fatal("expected verify false")
	}
}
