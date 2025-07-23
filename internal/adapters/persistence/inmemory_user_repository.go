package persistence

import (
	"context"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"nbox/internal/domain"
	"sync"
)

type userSchema map[string]struct {
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Status   string   `json:"status"`
}

type Users map[string]domain.User

type inMemoryUserRepository struct {
	users Users
	mu    sync.RWMutex
}

func NewInMemoryUserRepository(jsonCredentials []byte) (domain.UserRepository, error) {
	var schema userSchema
	if err := json.Unmarshal(jsonCredentials, &schema); err != nil {
		return nil, err
	}

	store := &inMemoryUserRepository{users: make(Users, len(schema))}

	store.mu.Lock()
	defer store.mu.Unlock()

	for username, data := range schema {
		store.users[username] = domain.User{
			Username: username,
			Password: data.Password,
			Roles:    data.Roles,
			Status:   data.Status,
		}
	}

	return store, nil
}

func (r *inMemoryUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[username]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

func (r *inMemoryUserRepository) ValidatePassword(ctx context.Context, username string, password string) (*domain.User, error) {
	user, err := r.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	return user, nil
}
