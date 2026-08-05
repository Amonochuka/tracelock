package access

import "testing"

func TestVerifyHashChain_ValidChain(t *testing.T) {
	links := []chainLink{
		{Hash: "hash1", PreviousHash: ""},
		{Hash: "hash2", PreviousHash: "hash1"},
		{Hash: "hash3", PreviousHash: "hash2"},
	}

	valid, count := verifyHashChain(links)

	if !valid {
		t.Error("expected chain to be valid")
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestVerifyHashChain_BrokenChain(t *testing.T) {
	links := []chainLink{
		{Hash: "hash1", PreviousHash: ""},
		{Hash: "hash2", PreviousHash: "hash1"},
		// tampered: this should reference "hash2", not "wrong_hash"
		{Hash: "hash3", PreviousHash: "wrong_hash"},
		{Hash: "hash4", PreviousHash: "hash3"},
	}

	valid, count := verifyHashChain(links)

	if valid {
		t.Error("expected chain to be detected as broken")
	}
	// should stop counting at the break — links 1 and 2 were valid, link 3 broke it
	if count != 2 {
		t.Errorf("expected count to stop at 2 (before the break), got %d", count)
	}
}

func TestVerifyHashChain_EmptyChain(t *testing.T) {
	var links []chainLink

	valid, count := verifyHashChain(links)

	if !valid {
		t.Error("expected an empty chain to be considered valid (nothing to break)")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestVerifyHashChain_SingleEvent(t *testing.T) {
	links := []chainLink{
		{Hash: "hash1", PreviousHash: ""}, // first event in a zone has no previous hash
	}

	valid, count := verifyHashChain(links)

	if !valid {
		t.Error("expected a single first event with empty PreviousHash to be valid")
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestVerifyHashChain_BrokenAtFirstLink(t *testing.T) {
	links := []chainLink{
		// first event should have PreviousHash == "", this one doesn't
		{Hash: "hash1", PreviousHash: "should_be_empty"},
		{Hash: "hash2", PreviousHash: "hash1"},
	}

	valid, count := verifyHashChain(links)

	if valid {
		t.Error("expected chain to be invalid when the first link's PreviousHash isn't empty")
	}
	if count != 0 {
		t.Errorf("expected count 0 since the break happens immediately, got %d", count)
	}
}
