package store

import (
	"errors"
	"sync"
	"time"

	"users-service/internal/model"
	"users-service/internal/security"
)

var ErrUserExists = errors.New("user already exists")

type UserStore struct {
	mu    sync.RWMutex
	users map[string]*model.User
	otps  map[string]security.OTPEntry
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*model.User),
		otps:  make(map[string]security.OTPEntry),
	}
}

func (s *UserStore) AddUser(user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Username == user.Username || u.Email == user.Email {
			return ErrUserExists
		}
	}

	s.users[user.Username] = user
	return nil
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (s *UserStore) SetOTP(username, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.otps[username] = security.OTPEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
}

func (s *UserStore) GetOTP(username string) (security.OTPEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.otps[username]
	return entry, ok
}

func (s *UserStore) DeleteOTP(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.otps, username)
}
