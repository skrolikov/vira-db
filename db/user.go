package db

import (
	"database/sql"
	"errors"
)

// User — структура пользователя
type User struct {
	ID           string
	Username     string
	PasswordHash string
}

// CreateUser — создает пользователя
func CreateUser(db *sql.DB, username, passwordHash string) error {
	// Здесь пример вставки в базу
	_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, passwordHash)
	if err != nil {
		return err
	}
	return nil
}

// GetUserByUsername — возвращает пользователя по имени
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // нет такого пользователя
		}
		return nil, err
	}
	return user, nil
}
