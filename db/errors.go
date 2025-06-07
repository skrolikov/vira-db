package db

import "errors"

var (
	ErrUserExists = errors.New("user already exists")
	ErrNoUser     = errors.New("user not found")
)
