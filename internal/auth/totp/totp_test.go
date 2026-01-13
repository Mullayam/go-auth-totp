package totp

import (
	"testing"
	"time"
)

type MockClock struct {
	Time time.Time
}

func (m MockClock) Now() time.Time {
	return m.Time
}

func TestRFC6238Vectors(t *testing.T) {
	secret := []byte("12345678901234567890")

	tests := []struct {
		Time     uint64
		Expected string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	}

	gen := NewGenerator()

	for _, tc := range tests {
		code, err := gen.GenerateCode(secret, tc.Time)
		if err != nil {
			t.Errorf("GenerateCode(%d) error: %v", tc.Time, err)
			continue
		}
		if code != tc.Expected {
			t.Errorf("GenerateCode(%d) = %s, want %s", tc.Time, code, tc.Expected)
		}
	}
}

func TestVerifyWindow(t *testing.T) {
	secret := []byte("12345678901234567890")

	// T = 1234567890
	mockTime := time.Unix(1234567890, 0)
	clock := MockClock{Time: mockTime}

	verifier := NewVerifier(clock, nil)
	verifier.Window = 1 // +/- 1 step

	// T (Current)
	gen := NewGenerator()
	codeNow, _ := gen.GenerateCode(secret, 1234567890)

	// T-1 (Previous 30s)
	codePrev, _ := gen.GenerateCode(secret, 1234567890-30)

	// T+1 (Next 30s)
	codeNext, _ := gen.GenerateCode(secret, 1234567890+30)

	// T-2 (Too old)
	codeOld, _ := gen.GenerateCode(secret, 1234567890-60)

	// Verify T (Should match)
	valid, err := verifier.Verify(secret, codeNow)
	if err != nil || !valid {
		t.Errorf("Verify(T) failed, got %v, %v", valid, err)
	}

	// Verify T-1 (Should match)
	valid, _ = verifier.Verify(secret, codePrev)
	if !valid {
		t.Errorf("Verify(T-1) failed")
	}

	// Verify T+1 (Should match)
	valid, _ = verifier.Verify(secret, codeNext)
	if !valid {
		t.Errorf("Verify(T+1) failed")
	}

	// Verify T-2 (Should fail)
	valid, _ = verifier.Verify(secret, codeOld)
	if valid {
		t.Errorf("Verify(T-2) passed, expected failure")
	}
}
