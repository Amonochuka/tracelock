package access

import (
	"testing"
	"time"
)

func TestGenerateHash(t *testing.T) {
	timestamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	hash1 := GenerateHash(1, 1, "enter", timestamp, "", "fingerprint")
	hash2 := GenerateHash(1, 1, "enter", timestamp, "", "fingerprint")

	if hash1 != hash2 {
		t.Errorf("expected same inputs to produce same hash, got %s and %s", hash1, hash2)
	}
}

func TestGenerateHashDifferentInputs(t *testing.T) {
	timestamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	hash1 := GenerateHash(1, 1, "enter", timestamp, "", "fingerprint")
	hash2 := GenerateHash(2, 1, "enter", timestamp, "", "fingerprint")

	if hash1 == hash2 {
		t.Error("expected different user IDs to produce different hashes")
	}
}