package storage

import (
	"errors"
	"sync"
)

var ErrUserNotFound = errors.New("user not found")

// User represents a user's TOTP state.
type User struct {
	ID              string
	EncryptedSecret string
	Enabled         bool
	RecoveryCodes   []string // Hashed recovery codes
}

// Repository defines the interface for user storage.
type Repository interface {
	GetUser(id string) (*User, error)
	SaveUser(user *User) error
}

// InMemoryRepository is a thread-safe in-memory implementation.
type InMemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		users: make(map[string]*User),
	}
}

func (r *InMemoryRepository) GetUser(id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	// Return a copy to avoid race conditions if caller modifies it
	userCopy := *u
	return &userCopy, nil
}

func (r *InMemoryRepository) SaveUser(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to store
	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}
