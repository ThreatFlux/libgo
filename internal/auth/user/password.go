package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordConfig defines the parameters for Argon2 password hashing.
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig returns the default password hashing configuration.
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword creates a password hash using Argon2id.
func HashPassword(password string) (string, error) {
	return HashPasswordWithConfig(password, DefaultPasswordConfig())
}

// HashPasswordWithConfig creates a password hash using the specified configuration.
func HashPasswordWithConfig(password string, config *PasswordConfig) (string, error) {
	// Generate a random salt
	salt := make([]byte, config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	// Hash the password with Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Encode as base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format with parameters
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		config.Memory,
		config.Iterations,
		config.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword verifies a password against a hash.
func VerifyPassword(password, hash string) bool {
	// Extract the parameters, salt and key from the encoded hash
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false
	}

	// Check if it's an Argon2id hash
	if parts[1] != "argon2id" {
		return false
	}

	// Parse the version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}
	if version != argon2.Version {
		return false
	}

	// Parse the parameters
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	// Decode the salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	// Decode the hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Compute the hash of the password with the same parameters
	// Check for integer overflow before converting to uint32
	if len(decodedHash) > math.MaxUint32 {
		return false
	}
	keyLength := uint32(len(decodedHash))
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// Compare the hashes in constant time
	return subtle.ConstantTimeCompare(decodedHash, computedHash) == 1
}

// ExtractPasswordConfig extracts the configuration from a password hash.
func ExtractPasswordConfig(hash string) (*PasswordConfig, error) {
	// Extract the parameters from the encoded hash
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid hash format")
	}

	// Check if it's an Argon2id hash
	if parts[1] != "argon2id" {
		return nil, fmt.Errorf("not an Argon2id hash")
	}

	// Parse the parameters
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return nil, fmt.Errorf("parsing parameters: %w", err)
	}

	// Decode the salt to get its length
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, fmt.Errorf("decoding salt: %w", err)
	}

	// Decode the hash to get its length
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, fmt.Errorf("decoding hash: %w", err)
	}

	// Check for integer overflow before converting to uint32
	if len(salt) > math.MaxUint32 {
		return nil, fmt.Errorf("salt length exceeds maximum value")
	}
	if len(decodedHash) > math.MaxUint32 {
		return nil, fmt.Errorf("hash length exceeds maximum value")
	}

	return &PasswordConfig{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  uint32(len(salt)),
		KeyLength:   uint32(len(decodedHash)),
	}, nil
}

// NeedsRehash checks if a password hash should be rehashed with new parameters.
func NeedsRehash(hash string, config *PasswordConfig) (bool, error) {
	currentConfig, err := ExtractPasswordConfig(hash)
	if err != nil {
		return false, err
	}

	// Compare configurations
	return currentConfig.Memory != config.Memory ||
		currentConfig.Iterations != config.Iterations ||
		currentConfig.Parallelism != config.Parallelism ||
		currentConfig.KeyLength != config.KeyLength, nil
}
