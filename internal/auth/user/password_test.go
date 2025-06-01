package user

import (
	"strings"
	"testing"
)

const (
	testPasswordValue = "test-password"
)

func TestHashPassword(t *testing.T) {
	password := testPasswordValue

	// Test with default config
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Hash should be in the format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("Hash doesn't have expected format, got: %s", hash)
	}

	// Test with custom config
	customConfig := &PasswordConfig{
		Memory:      32 * 1024,
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  8,
		KeyLength:   16,
	}

	customHash, err := HashPasswordWithConfig(password, customConfig)
	if err != nil {
		t.Fatalf("HashPasswordWithConfig failed: %v", err)
	}

	// Ensure the custom config was used
	if !strings.Contains(customHash, "m=32768,t=2,p=1") {
		t.Errorf("Custom hash doesn't have expected parameters, got: %s", customHash)
	}

	// Ensure different salt is used each time
	anotherHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	if hash == anotherHash {
		t.Error("Hashing the same password twice should produce different hashes")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := testPasswordValue
	wrongPassword := "wrong-password"

	// Generate a hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify with correct password
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword should return true for the correct password")
	}

	// Verify with wrong password
	if VerifyPassword(wrongPassword, hash) {
		t.Error("VerifyPassword should return false for the wrong password")
	}

	// Verify with invalid hash format
	invalidHash := "$invalidformat$"
	if VerifyPassword(password, invalidHash) {
		t.Error("VerifyPassword should return false for an invalid hash format")
	}

	// Verify with hash that has wrong algorithm
	wrongAlgHash := "$argon2i$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if VerifyPassword(password, wrongAlgHash) {
		t.Error("VerifyPassword should return false for hash with wrong algorithm")
	}
}

func TestExtractPasswordConfig(t *testing.T) {
	// Create a hash with known config
	knownConfig := &PasswordConfig{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	password := testPasswordValue
	hash, err := HashPasswordWithConfig(password, knownConfig)
	if err != nil {
		t.Fatalf("HashPasswordWithConfig failed: %v", err)
	}

	// Extract config from the hash
	extractedConfig, err := ExtractPasswordConfig(hash)
	if err != nil {
		t.Fatalf("ExtractPasswordConfig failed: %v", err)
	}

	// Verify the extracted config matches the original
	if extractedConfig.Memory != knownConfig.Memory {
		t.Errorf("Extracted memory value doesn't match: got %d, want %d",
			extractedConfig.Memory, knownConfig.Memory)
	}

	if extractedConfig.Iterations != knownConfig.Iterations {
		t.Errorf("Extracted iterations value doesn't match: got %d, want %d",
			extractedConfig.Iterations, knownConfig.Iterations)
	}

	if extractedConfig.Parallelism != knownConfig.Parallelism {
		t.Errorf("Extracted parallelism value doesn't match: got %d, want %d",
			extractedConfig.Parallelism, knownConfig.Parallelism)
	}

	if extractedConfig.SaltLength != knownConfig.SaltLength {
		t.Errorf("Extracted salt length value doesn't match: got %d, want %d",
			extractedConfig.SaltLength, knownConfig.SaltLength)
	}

	if extractedConfig.KeyLength != knownConfig.KeyLength {
		t.Errorf("Extracted key length value doesn't match: got %d, want %d",
			extractedConfig.KeyLength, knownConfig.KeyLength)
	}

	// Test with invalid hash format
	_, err = ExtractPasswordConfig("invalid-hash")
	if err == nil {
		t.Error("ExtractPasswordConfig should fail with an invalid hash format")
	}

	// Test with wrong algorithm
	wrongAlgHash := "$argon2i$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	_, err = ExtractPasswordConfig(wrongAlgHash)
	if err == nil {
		t.Error("ExtractPasswordConfig should fail with a hash using wrong algorithm")
	}
}

func TestNeedsRehash(t *testing.T) {
	password := testPasswordValue

	// Create a hash with default config
	defaultConfig := DefaultPasswordConfig()
	hash, err := HashPasswordWithConfig(password, defaultConfig)
	if err != nil {
		t.Fatalf("HashPasswordWithConfig failed: %v", err)
	}

	// Test with the same config - should not need rehash
	needsRehash, err := NeedsRehash(hash, defaultConfig)
	if err != nil {
		t.Fatalf("NeedsRehash failed: %v", err)
	}
	if needsRehash {
		t.Error("NeedsRehash should return false for the same config")
	}

	// Test with a different config - should need rehash
	newConfig := &PasswordConfig{
		Memory:      128 * 1024, // Different memory value
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	needsRehash, err = NeedsRehash(hash, newConfig)
	if err != nil {
		t.Fatalf("NeedsRehash failed: %v", err)
	}
	if !needsRehash {
		t.Error("NeedsRehash should return true for different config")
	}

	// Test with invalid hash format
	_, err = NeedsRehash("invalid-hash", defaultConfig)
	if err == nil {
		t.Error("NeedsRehash should fail with an invalid hash format")
	}
}
