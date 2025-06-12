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
	Email        string
	Role         string
	Confirmed    bool
	ConfirmToken string
}

type userRepo struct {
	db *sql.DB
}

func (r *userRepo) GetUserByID(id string) (*User, error) {
	query := "SELECT id, username, password FROM users WHERE id = $1"
	user := &User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // пользователь не найден
		}
		return nil, err // другая ошибка
	}
	return user, nil
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) ExistsByUsername(username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *userRepo) ExistsByEmail(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *userRepo) CreateUserExtended(username, passwordHash, email, role string, confirmed bool, confirmToken string) (string, error) {
	var userID string
	err := r.db.QueryRow(`
		INSERT INTO users (username, password, email, role, confirmed, confirm_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, username, passwordHash, email, role, confirmed, confirmToken).Scan(&userID)

	if err != nil {
		// Можно тут обработать уникальность username/email, если настроены уникальные индексы
		return "", err
	}
	return userID, nil
}

func (r *userRepo) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, username, password, email, role, confirmed, confirm_token
		FROM users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role, &user.Confirmed, &user.ConfirmToken)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
