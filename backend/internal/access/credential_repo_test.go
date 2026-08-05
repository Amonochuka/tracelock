package access

import (
	"strings"
	"testing"
)

func TestNormalizeCredentialHash_RawCredentialIsHashed(t *testing.T) {
	raw := "example biometric template"
	hashed := normalizeCredentialHash(raw)

	if hashed == raw {
		t.Fatal("expected raw biometric credential to be hashed")
	}

	if len(hashed) != 64 {
		t.Fatalf("expected a 64-character hex hash, got %d", len(hashed))
	}
}

func TestNormalizeCredentialHash_PreservesHexCredential(t *testing.T) {
	hex := strings.Repeat("a", 64)
	normalized := normalizeCredentialHash(hex)

	if normalized != hex {
		t.Fatalf("expected normalized credential to preserve 64-char hex string, got %q", normalized)
	}
}

func TestNormalizeCredentialHash_LowercasesHexCredential(t *testing.T) {
	upperHex := strings.ToUpper(strings.Repeat("a", 64))
	normalized := normalizeCredentialHash(upperHex)

	if normalized != strings.ToLower(upperHex) {
		t.Fatalf("expected normalized credential to lowercase hex string, got %q", normalized)
	}
}
