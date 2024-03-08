// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import "sync"

// UserStore is an interface to store users
type UserStore interface {
	// Save saves a user to the store
	Save(user *User) error
	// Find finds a user by username
	Find(username string) (*User, error)
}

// InMemoryUserStore stores users in memory
type InMemoryUserStore struct {
	sync.RWMutex

	users map[string]*User
}

// NewInMemoryUserStore returns a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// Save saves a user to the store
func (m *InMemoryUserStore) Save(user *User) error {
	m.Lock()
	defer m.Unlock()

	if m.users[user.Username] != nil {
		return ErrAlreadyExists
	}

	m.users[user.Username] = user.Clone()

	return nil
}

// Find finds a user by username
func (m *InMemoryUserStore) Find(username string) (*User, error) {
	m.RLock()
	defer m.RUnlock()

	user := m.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
