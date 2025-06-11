package db

import "errors"

var (
	// ErrUserExists - пользователь уже существует
	ErrUserExists = errors.New("пользователь уже существует")

	// ErrNoUser - пользователь не найден
	ErrNoUser = errors.New("пользователь не найден")

	// ErrInvalidCredentials - неверные учетные данные
	ErrInvalidCredentials = errors.New("неверные учетные данные")

	// ErrUserDisabled - пользователь отключен/заблокирован
	ErrUserDisabled = errors.New("пользователь отключен/заблокирован")
)
