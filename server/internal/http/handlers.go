package http

import (
	"encoding/json"
	"go-auth-totp/internal/auth/enroll"
	"go-auth-totp/internal/auth/ratelimit"
	"go-auth-totp/internal/auth/recovery"
	"go-auth-totp/internal/auth/totp"
	"go-auth-totp/internal/crypto"
	"go-auth-totp/internal/storage"
	"log"
	"net/http"
)

type Handlers struct {
	Repo        storage.Repository
	Crypto      crypto.CryptoService
	EnrollSvc   *enroll.Service
	RecoverySvc *recovery.Service
	Verifier    *totp.Verifier
	Limiter     ratelimit.Limiter
}

type EnrollRequest struct {
	UserID string `json:"user_id"`
}

type VerifyRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

func (h *Handlers) EncodeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handlers) ErrorJSON(w http.ResponseWriter, status int, msg string) {
	h.EncodeJSON(w, status, map[string]string{"error": msg})
}

// EnrollHandler initiates the enrollment process.
func (h *Handlers) EnrollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req EnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 1. Generate Secret & QR
	log.Printf("Enrolling user: %s", req.UserID)
	resp, err := h.EnrollSvc.Enroll(req.UserID)
	if err != nil {
		log.Printf("Enrollment failed for %s: %v", req.UserID, err)
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate secret")
		return
	}

	// 2. Storage: Save user with DISABLED state
	user := &storage.User{
		ID:              req.UserID,
		EncryptedSecret: resp.EncryptedBlob,
		RecoveryCodes:   resp.HashedCodes,
		Enabled:         false, // IMPORTANT: Not enabled until verified
	}
	if err := h.Repo.SaveUser(user); err != nil {
		log.Printf("SaveUser failed for %s: %v", req.UserID, err)
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to save user")
		return
	}
	log.Printf("User %s saved to DB", req.UserID)

	// 3. Return Secret & QR URL
	// In production, might render the QR code as PNG data URI here.
	h.EncodeJSON(w, http.StatusOK, resp)
}

// VerifyHandler confirms the first code and enables TOTP.
func (h *Handlers) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	log.Printf("Verifying user: %s with code: %s", req.UserID, req.Code)

	// Rate Limit
	if !h.Limiter.Allow(req.UserID) {
		h.ErrorJSON(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}

	// 1. Load User
	user, err := h.Repo.GetUser(req.UserID)
	if err != nil {
		log.Printf("GetUser failed for %s: %v", req.UserID, err)
		h.ErrorJSON(w, http.StatusNotFound, "User not found")
		return
	}

	if user.Enabled {
		h.ErrorJSON(w, http.StatusConflict, "TOTP already enabled")
		return
	}

	// 2. Decrypt Secret
	secretBytes, err := h.Crypto.Decrypt(user.EncryptedSecret)
	if err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to decrypt secret")
		return
	}

	// 3. Verify Code
	valid, err := h.Verifier.Verify(secretBytes, req.Code)
	if err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Verification error")
		return
	}

	if !valid {
		h.ErrorJSON(w, http.StatusUnauthorized, "Invalid code")
		return
	}

	// 4. Enable TOTP
	user.Enabled = true
	if err := h.Repo.SaveUser(user); err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	h.EncodeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

// ValidateHandler checks a code for an enrolled user (Login flow).
func (h *Handlers) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Rate Limit
	if !h.Limiter.Allow(req.UserID) {
		h.ErrorJSON(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}

	// 1. Load User
	user, err := h.Repo.GetUser(req.UserID)
	if err != nil {
		h.ErrorJSON(w, http.StatusNotFound, "User not found")
		return
	}

	if !user.Enabled {
		h.ErrorJSON(w, http.StatusPreconditionFailed, "TOTP not enabled")
		return
	}

	// 2. Decrypt Secret
	secretBytes, err := h.Crypto.Decrypt(user.EncryptedSecret)
	if err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to decrypt secret")
		return
	}

	// 3. Verify Code
	valid, err := h.Verifier.Verify(secretBytes, req.Code)
	if err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Verification error")
		return
	}

	if !valid {
		h.ErrorJSON(w, http.StatusUnauthorized, "Invalid code")
		return
	}

	h.EncodeJSON(w, http.StatusOK, map[string]string{"status": "valid"})
}

// RecoverHandler allows login using a recovery code.
func (h *Handlers) RecoverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Rate Limit
	if !h.Limiter.Allow(req.UserID) {
		h.ErrorJSON(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}

	// 1. Load User
	user, err := h.Repo.GetUser(req.UserID)
	if err != nil {
		h.ErrorJSON(w, http.StatusNotFound, "User not found")
		return
	}

	if !user.Enabled {
		h.ErrorJSON(w, http.StatusPreconditionFailed, "TOTP not enabled")
		return
	}

	// 2. Validate Recovery Code
	remainingCodes, ok := h.RecoverySvc.ValidateAndConsume(req.Code, user.RecoveryCodes)
	if !ok {
		h.ErrorJSON(w, http.StatusUnauthorized, "Invalid recovery code")
		return
	}

	// 3. Update User (Remove used code)
	user.RecoveryCodes = remainingCodes
	if err := h.Repo.SaveUser(user); err != nil {
		h.ErrorJSON(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	h.EncodeJSON(w, http.StatusOK, map[string]string{"status": "recovered", "msg": "Recovery code accepted"})
}
