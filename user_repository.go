package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// User — структура пользователя с дополнительными полями
type User struct {
	ID              string
	Username        string
	PasswordHash    string
	Email           string
	Role            string
	Confirmed       bool
	ConfirmToken    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     sql.NullTime
	PasswordChanged sql.NullTime
}

type userRepo struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр репозитория пользователей
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

// GetUserByID возвращает пользователя по ID
func (r *userRepo) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, username, password, email, role, confirmed, confirm_token, 
		       created_at, updated_at, last_login_at, password_changed
		FROM users 
		WHERE id = $1`

	user := &User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
		&user.Confirmed, &user.ConfirmToken, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.PasswordChanged,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserByUsername возвращает пользователя по имени пользователя
func (r *userRepo) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, password, email, role, confirmed, confirm_token,
		       created_at, updated_at, last_login_at, password_changed
		FROM users 
		WHERE username = $1`

	user := &User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
		&user.Confirmed, &user.ConfirmToken, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.PasswordChanged,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// GetUserByEmail возвращает пользователя по email
func (r *userRepo) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, password, email, role, confirmed, confirm_token,
		       created_at, updated_at, last_login_at, password_changed
		FROM users 
		WHERE email = $1`

	user := &User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
		&user.Confirmed, &user.ConfirmToken, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.PasswordChanged,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// ExistsByUsername проверяет существование пользователя с заданным именем
func (r *userRepo) ExistsByUsername(username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// ExistsByEmail проверяет существование пользователя с заданным email
func (r *userRepo) ExistsByEmail(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

// CreateUserExtended создает нового пользователя с расширенными полями
func (r *userRepo) CreateUserExtended(username, passwordHash, email, role string, confirmed bool, confirmToken string) (string, error) {
	var userID string
	query := `
		INSERT INTO users (username, password, email, role, confirmed, confirm_token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRow(query, username, passwordHash, email, role, confirmed, confirmToken).Scan(&userID)

	if err != nil {
		// Проверяем на нарушение уникальности
		if isUniqueConstraintError(err, "username") {
			return "", ErrDuplicateUsername
		}
		if isUniqueConstraintError(err, "email") {
			return "", ErrDuplicateEmail
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

// UpdateUser обновляет данные пользователя
func (r *userRepo) UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, role = $3, confirmed = $4, 
		    updated_at = NOW(), last_login_at = $5, password_changed = $6
		WHERE id = $7`

	_, err := r.db.Exec(query,
		user.Username, user.Email, user.Role, user.Confirmed,
		user.LastLoginAt, user.PasswordChanged, user.ID,
	)

	if err != nil {
		if isUniqueConstraintError(err, "username") {
			return ErrDuplicateUsername
		}
		if isUniqueConstraintError(err, "email") {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser удаляет пользователя по ID
func (r *userRepo) DeleteUser(id string) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ConfirmUser подтверждает пользователя по email и токену
func (r *userRepo) ConfirmUser(email, token string) error {
	query := `
		UPDATE users 
		SET confirmed = TRUE, confirm_token = ''
		WHERE email = $1 AND confirm_token = $2 AND NOT confirmed`

	result, err := r.db.Exec(query, email, token)
	if err != nil {
		return fmt.Errorf("failed to confirm user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdatePassword обновляет хэш пароля пользователя
func (r *userRepo) UpdatePassword(id, newHash string) error {
	query := `
		UPDATE users 
		SET password = $1, password_changed = NOW()
		WHERE id = $2`

	_, err := r.db.Exec(query, newHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// GetUsersByRole возвращает список пользователей с определенной ролью
func (r *userRepo) GetUsersByRole(role string, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, username, password, email, role, confirmed, confirm_token,
		       created_at, updated_at, last_login_at, password_changed
		FROM users 
		WHERE role = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, role, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by role: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Role,
			&user.Confirmed, &user.ConfirmToken, &user.CreatedAt, &user.UpdatedAt,
			&user.LastLoginAt, &user.PasswordChanged,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

// isUniqueConstraintError проверяет, является ли ошибка нарушением уникальности для указанного поля
func isUniqueConstraintError(err error, field string) bool {
	return err != nil && err.Error() == fmt.Sprintf("pq: duplicate key value violates unique constraint \"users_%s_key\"", field)
}
