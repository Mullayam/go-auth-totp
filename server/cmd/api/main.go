package main

import (
	"go-auth-totp/internal/auth/enroll"
	"go-auth-totp/internal/auth/ratelimit"
	"go-auth-totp/internal/auth/recovery"
	"go-auth-totp/internal/auth/totp"
	"go-auth-totp/internal/config"
	"go-auth-totp/internal/crypto"
	internalHttp "go-auth-totp/internal/http"
	"go-auth-totp/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cryptoSvc, err := crypto.NewAESGCMEncryption(cfg.MasterKey)
	if err != nil {
		log.Fatalf("Failed to init crypto: %v", err)
	}

	// 2. Setup Services
	repo, err := storage.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}

	enrollSvc := enroll.NewService(cfg.AppName, cryptoSvc)
	recoverySvc := recovery.NewService()
	// Pass nil to use RealClock
	verifier := totp.NewVerifier(nil, cfg)
	// Limit: 3 attempts, refill 1 every 30s
	limiter := ratelimit.NewInMemoryLimiter(30*time.Second, 3)

	// 3. Setup Handlers
	h := &internalHttp.Handlers{
		Repo:        repo,
		Crypto:      cryptoSvc,
		EnrollSvc:   enrollSvc,
		RecoverySvc: recoverySvc,
		Verifier:    verifier,
		Limiter:     limiter,
	}

	r := mux.NewRouter()
	r.HandleFunc("/enroll", h.EnrollHandler).Methods("POST")
	r.HandleFunc("/verify", h.VerifyHandler).Methods("POST")
	r.HandleFunc("/validate", h.ValidateHandler).Methods("POST")
	r.HandleFunc("/recover", h.RecoverHandler).Methods("POST")

	// 4. Start Server
	log.Printf("Server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
