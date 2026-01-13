package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// Generator handles the creation of TOTP codes.
type Generator struct {
	Digits int
	Period uint64
}

// NewGenerator returns a generator configured for Google Authenticator compatibility.
// defaults: 6 digits, 30 second period.
func NewGenerator() *Generator {
	return &Generator{
		Digits: 6,
		Period: 30,
	}
}

// GenerateCode creates a TOTP code for the given secret and time.
// secret is expected to be a raw byte slice (not base32 encoded).
func (g *Generator) GenerateCode(secret []byte, timestamp uint64) (string, error) {
	counter := timestamp / g.Period
	return g.generateHOTP(secret, counter)
}

// GenerateCodeFromBase32 is a helper to generate from a base32 string secret.
func (g *Generator) GenerateCodeFromBase32(secretBase32 string, timestamp uint64) (string, error) {
	// Add padding if missing, though standard is no padding
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secretBase32))
	if err != nil {
		// Try with standard padding just in case
		secretBytes, err = base32.StdEncoding.DecodeString(strings.ToUpper(secretBase32))
		if err != nil {
			return "", fmt.Errorf("invalid base32 secret: %v", err)
		}
	}
	return g.GenerateCode(secretBytes, timestamp)
}

// generateHOTP generates an HOTP token (RFC 4226)
func (g *Generator) generateHOTP(secret []byte, counter uint64) (string, error) {
	// 1. HMAC-SHA1(K, C)
	h := hmac.New(sha1.New, secret)
	if err := binary.Write(h, binary.BigEndian, counter); err != nil {
		return "", err
	}
	sum := h.Sum(nil)

	// 2. Dynamic Truncation
	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (int(sum[offset]&0x7f) << 24) |
		(int(sum[offset+1]&0xff) << 16) |
		(int(sum[offset+2]&0xff) << 8) |
		(int(sum[offset+3] & 0xff))

	// 3. Modulo 10^Digits
	otp := binaryCode % int(math.Pow10(g.Digits))

	// 4. Checksum / Padding
	format := fmt.Sprintf("%%0%dd", g.Digits)
	return fmt.Sprintf(format, otp), nil
}
