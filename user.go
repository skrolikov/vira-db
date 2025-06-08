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
func CreateUser(db *sql.DB, username, passwordHash string) (string, error) {
	var userID string
	err := db.QueryRow(`
		INSERT INTO users (username, password) 
		VALUES ($1, $2) 
		RETURNING id
	`, username, passwordHash).Scan(&userID)

	if err != nil {
		return "", err
	}
	return userID, nil
}

// GetUserByUsername — возвращает пользователя по имени
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // нет такого пользователя
		}
		return nil, err
	}
	return user, nil
}
