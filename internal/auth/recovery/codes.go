package recovery

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

const (
	CodeLength = 10
	CodeCount  = 8
	// Charset for recovery codes (alphanumeric, easy to read)
	Charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// GenerateCodes creates a set of recovery codes and their hashes.
// Returns plainCodes (for user) and hashedCodes (for storage).
func (s *Service) GenerateCodes() ([]string, []string, error) {
	plainCodes := make([]string, CodeCount)
	hashedCodes := make([]string, CodeCount)

	for i := 0; i < CodeCount; i++ {
		code, err := generateRandomString(CodeLength)
		if err != nil {
			return nil, nil, err
		}
		plainCodes[i] = code
		hashedCodes[i] = hash(code)
	}

	return plainCodes, hashedCodes, nil
}

// ValidateAndConsume checks if the provided code validates against any of the stored hashes.
// If valid, returns the new list of hashed codes (removing the used one) and true.
// If invalid, returns the original list and false.
func (s *Service) ValidateAndConsume(inputCode string, storedHashes []string) ([]string, bool) {
	inputHash := hash(inputCode)

	// Check against all stored hashes
	// This is O(N) but N is small (8-10).
	// We do NOT need constant time because these are one-time codes and high entropy.
	// But strictly speaking, we should be careful.
	// For recovery codes, timing attacks are less practical due to lockout/rate limiting,
	// preventing brute force of 10 chars from charset 32 (~50 bits entropy).

	for i, h := range storedHashes {
		if h == inputHash {
			// Found match! Remove it.
			// Fast delete from slice
			newHashes := append(storedHashes[:i], storedHashes[i+1:]...)
			return newHashes, true
		}
	}

	return storedHashes, false
}

func hash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(Charset))))
		if err != nil {
			return "", err
		}
		b[i] = Charset[num.Int64()]
	}
	return string(b), nil
}
