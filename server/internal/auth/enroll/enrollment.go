package enroll

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"go-auth-totp/internal/auth/recovery"
	"go-auth-totp/internal/crypto"
	"net/url"
)

// Service handles new TOTP enrollments.
type Service struct {
	issuer      string
	crypto      crypto.CryptoService
	recoverySvc *recovery.Service
}

// NewService creates a new enrollment service.
func NewService(issuer string, cryptoService crypto.CryptoService) *Service {
	return &Service{
		issuer:      issuer,
		crypto:      cryptoService,
		recoverySvc: recovery.NewService(),
	}
}

// EnrollmentResponse contains the data needed for the client to set up TOTP.
type EnrollmentResponse struct {
	Secret        string   // The base32 encoded secret (for manual entry)
	EncryptedBlob string   // The encrypted secret (to be stored in DB)
	OTPAuthURL    string   // The URL for QR code generation
	RecoveryCodes []string // Plaintext recovery codes to show ONCE
	HashedCodes   []string // Hashed codes for storage
}

// Enroll initiates the TOTP enrollment for a user.
func (s *Service) Enroll(accountName string) (*EnrollmentResponse, error) {
	// 1. Generate 20-byte random secret
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// 2. Encode to Base32 (no padding) for standard compatibility
	secretBase32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	// 3. Encrypt the raw bytes for storage
	encryptedBlob, err := s.crypto.Encrypt(secretBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// 4. Generate Recovery Codes
	plainCodes, hashedCodes, err := s.recoverySvc.GenerateCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	// 5. Generate otpauth URL
	// Format: otpauth://totp/Issuer:Account?secret=SECRET&issuer=Issuer&algorithm=SHA1&digits=6&period=30
	v := url.Values{}
	v.Set("secret", secretBase32)
	v.Set("issuer", s.issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	v.Set("period", "30")

	// The label is "Issuer:Account"
	label := fmt.Sprintf("%s:%s", s.issuer, accountName)
	otpUrl := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     label,
		RawQuery: v.Encode(),
	}

	return &EnrollmentResponse{
		Secret:        secretBase32,
		EncryptedBlob: encryptedBlob,
		OTPAuthURL:    otpUrl.String(),
		RecoveryCodes: plainCodes,
		HashedCodes:   hashedCodes,
	}, nil
}
