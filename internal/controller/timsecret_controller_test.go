package controller

import (
	"testing"
)

func TestCalculateHash_Deterministic(t *testing.T) {
	data := map[string]string{
		"password": "secret123",
		"username": "admin",
		"database": "postgres",
		"api_key":  "abc123xyz",
	}

	// Calculate hash multiple times
	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		hashes[i] = calculateHash(data)
	}

	// All hashes should be identical
	firstHash := hashes[0]
	for i, hash := range hashes {
		if hash != firstHash {
			t.Errorf("Hash %d differs from first hash:\n  Expected: %s\n  Got: %s",
				i, firstHash, hash)
		}
	}
}

func TestCalculateHash_DifferentData(t *testing.T) {
	data1 := map[string]string{
		"password": "secret123",
		"username": "admin",
	}

	data2 := map[string]string{
		"password": "different",
		"username": "admin",
	}

	hash1 := calculateHash(data1)
	hash2 := calculateHash(data2)

	if hash1 == hash2 {
		t.Error("Different data produced the same hash")
	}
}

func TestCalculateHash_SameDataDifferentOrder(t *testing.T) {
	// Data with keys in different insertion order
	data1 := map[string]string{
		"a": "value1",
		"b": "value2",
		"c": "value3",
	}

	data2 := map[string]string{
		"c": "value3",
		"a": "value1",
		"b": "value2",
	}

	hash1 := calculateHash(data1)
	hash2 := calculateHash(data2)

	if hash1 != hash2 {
		t.Errorf("Same data with different order produced different hashes:\n  Hash1: %s\n  Hash2: %s",
			hash1, hash2)
	}
}

func TestCalculateHash_EmptyData(t *testing.T) {
	data := map[string]string{}
	hash := calculateHash(data)

	if hash == "" {
		t.Error("Empty data should still produce a hash")
	}

	// Empty data should always produce the same hash
	hash2 := calculateHash(data)
	if hash != hash2 {
		t.Error("Empty data produced different hashes")
	}
}
