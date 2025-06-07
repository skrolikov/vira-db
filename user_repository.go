package db

import (
	"database/sql"
	"errors"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(username, passwordHash string) (string, error) {
	var userID string
	err := r.db.QueryRow(`
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`, username, passwordHash).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *userRepo) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow("SELECT id, username, password_hash FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
