package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestBcryptHasherRoundTrip(t *testing.T) {
	h := NewBcryptHasher(0)
	hash, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "" || hash == "password123" {
		t.Error("hash should not be empty or equal to plaintext")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("hash should start with $2 (bcrypt marker), got %q", hash[:min(2, len(hash))])
	}
	if err := h.Verify("password123", hash); err != nil {
		t.Errorf("Verify with correct password: %v", err)
	}
}

func TestBcryptHasherRejectsWrongPassword(t *testing.T) {
	h := NewBcryptHasher(0)
	hash, _ := h.Hash("password123")
	err := h.Verify("WRONG", hash)
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("wrong password: got %v, want ErrInvalidPassword", err)
	}
}

func TestBcryptHasherCostZeroUsesDefault(t *testing.T) {
	h1 := NewBcryptHasher(0)
	h2 := NewBcryptHasher(-5)
	hash1, _ := h1.Hash("x")
	hash2, _ := h2.Hash("x")
	if hash1 == hash2 {
		t.Error("two random salts should produce different hashes")
	}
	if err := h1.Verify("x", hash1); err != nil {
		t.Error("verify should succeed")
	}
	if err := h2.Verify("x", hash2); err != nil {
		t.Error("verify should succeed")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
