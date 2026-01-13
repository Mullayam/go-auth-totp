package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	MasterKey  []byte
	DBPath     string
	Port       string
	WindowSize uint64
}

func Load() (*Config, error) {
	// Load .env file if it exists, ignore error if missing (e.g. prod env vars)
	_ = godotenv.Load()
	windowSize, _ := strconv.ParseUint(getEnv("WINDOW_SIZE", "1"), 10, 64)

	cfg := &Config{
		AppName:    getEnv("TOTP_APP_NAME", "EnjoysAuthTOTP"),
		DBPath:     getEnv("DB_PATH", "totp.db"),
		Port:       getEnv("PORT", "8080"),
		WindowSize: windowSize,
	}

	masterKeyHex := os.Getenv("TOTP_MASTER_KEY")
	if masterKeyHex != "" {
		key, err := hex.DecodeString(masterKeyHex)
		if err != nil {
			return nil, fmt.Errorf("invalid TOTP_MASTER_KEY hex: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("TOTP_MASTER_KEY must be 32 bytes, got %d", len(key))
		}
		cfg.MasterKey = key
	} else {
		log.Println("WARNING: TOTP_MASTER_KEY not set. Generating random key for this session (NOT PERSISTENT).")
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate random key: %w", err)
		}
		log.Printf("Generated Master Key (Hex): %x", key)
		cfg.MasterKey = key
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
