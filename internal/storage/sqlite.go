package storage

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.initSchema(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		encrypted_secret TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS recovery_codes (
		user_id TEXT,
		code_hash TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) GetUser(id string) (*User, error) {
	var user User
	var enabled bool // driver handles BOOLEAN as bool

	err := r.db.QueryRow("SELECT id, encrypted_secret, enabled FROM users WHERE id = ?", id).Scan(&user.ID, &user.EncryptedSecret, &enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user.Enabled = enabled

	// Load recovery codes
	rows, err := r.db.Query("SELECT code_hash FROM recovery_codes WHERE user_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		user.RecoveryCodes = append(user.RecoveryCodes, hash)
	}

	return &user, nil
}

func (r *SQLiteRepository) SaveUser(user *User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Upsert User
	_, err = tx.Exec(`
		INSERT INTO users (id, encrypted_secret, enabled) VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET encrypted_secret = ?, enabled = ?
	`, user.ID, user.EncryptedSecret, user.Enabled, user.EncryptedSecret, user.Enabled)
	if err != nil {
		return err
	}

	// 2. Replace Recovery Codes (Full replace strategy for simplicity)
	_, err = tx.Exec("DELETE FROM recovery_codes WHERE user_id = ?", user.ID)
	if err != nil {
		return err
	}

	// Bulk insert could be better, but loop is fine for 8 codes
	stmt, err := tx.Prepare("INSERT INTO recovery_codes (user_id, code_hash) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, hash := range user.RecoveryCodes {
		if _, err := stmt.Exec(user.ID, hash); err != nil {
			return err
		}
	}

	return tx.Commit()
}
