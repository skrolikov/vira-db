package db

import "errors"

var (
	ErrUserExists         = errors.New("пользователь уже существует")
	ErrNoUser             = errors.New("пользователь не найден")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrUserDisabled       = errors.New("пользователь отключен/заблокирован")
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrDuplicateUsername  = errors.New("пользователь уже существует")
	ErrDuplicateEmail     = errors.New("email уже зарегистрирован")
	ErrDBNotInitialized   = errors.New("база данных не инициализирована")
	ErrDBConnectionLost   = errors.New("потеряно соединение с базой данных")
)
