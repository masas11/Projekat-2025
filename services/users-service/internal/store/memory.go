package store

import (
	"errors"
	"sync"

	"users-service/internal/model"
)

var (
	ErrUserExists = errors.New("user with given username or email already exists")
)

type UserStore struct {
	mu    sync.RWMutex
	users map[string]*model.User // key = username
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*model.User),
	}
}

func (s *UserStore) AddUser(user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username]; exists {
		return ErrUserExists
	}

	for _, u := range s.users {
		if u.Email == user.Email {
			return ErrUserExists
		}
	}

	s.users[user.Username] = user
	return nil
}
