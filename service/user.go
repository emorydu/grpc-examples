// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// User contains user's information
type User struct {
	Username       string
	HashedPassword string
	Role           string
}

// NewUser returns a new user
func NewUser(username string, password string, role string) (*User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	u := &User{
		Username:       username,
		HashedPassword: string(hashed),
		Role:           role,
	}

	return u, nil
}

// IsCorrectPassword checks if the provided password is correct or not
func (u *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

// Clone returns a clone of this user
func (u *User) Clone() *User {
	return &User{
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
		Role:           u.Role,
	}
}
