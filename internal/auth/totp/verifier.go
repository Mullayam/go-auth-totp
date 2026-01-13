package totp

import (
	"crypto/subtle"
	"go-auth-totp/internal/config"
	"go-auth-totp/pkg/timeutil"
)

// Verifier handles the validation of TOTP codes.
type Verifier struct {
	generator *Generator
	clock     timeutil.Clock
	// Window represents the number of steps to check before and after the current time.
	// A window of 1 means checking T-1, T, T+1.
	Window uint64
}

// NewVerifier creates a secure verifier with default settings.
func NewVerifier(clock timeutil.Clock, cfg *config.Config) *Verifier {
	if clock == nil {
		clock = timeutil.RealClock{}
	}
	return &Verifier{
		generator: NewGenerator(),
		clock:     clock,
		Window:    uint64(cfg.WindowSize), // Allow +/- 30/60 seconds drift (offset user clock)
	}
}

// Verify checks if the provided code is valid for the given secret at the current time.
// It checks the current time step and +/- Window steps.
// Returns true if valid, false otherwise.
// This function MUST be constant time where possible (the comparison is).
func (v *Verifier) Verify(secret []byte, inputCode string) (bool, error) {
	if len(inputCode) != v.generator.Digits {
		return false, nil // Invalid length, fail fast (length is not sensitive)
	}

	currentTime := uint64(v.clock.Now().Unix())
	currentStep := currentTime / v.generator.Period

	// Check window: [current - window, current + window]
	// We iterate through all checks to ensure roughly constant work (though generation time might vary slightly)
	start := currentStep - v.Window
	end := currentStep + v.Window
	matched := 0

	// log.Printf("DEBUG: Verifying Input: %s at Time: %d (Step: %d)", inputCode, currentTime, currentStep)

	for step := start; step <= end; step++ {
		// Calculate what the time would be for this step (approx, just need step for generation)
		validCode, err := v.generator.generateHOTP(secret, step)
		if err != nil {
			return false, err
		}

		// log.Printf("DEBUG: Step %d -> Expected: %s", step, validCode)

		// subtle.ConstantTimeCompare returns 1 if equal, 0 otherwise
		if subtle.ConstantTimeCompare([]byte(validCode), []byte(inputCode)) == 1 {
			matched = 1
		}
	}

	return matched == 1, nil
}
