package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrLoginNotFound = errors.New("login record not found")
)

// UserLogin представляет запись о входе пользователя
type UserLogin struct {
	ID         string
	UserID     string
	Username   string
	IP         string
	UserAgent  string
	LoginTime  time.Time
	LogoutTime sql.NullTime
	SessionID  string
	Success    bool
	FailReason sql.NullString
}

// UserLoginRepository определяет интерфейс для работы с историей входов
type UserLoginRepository interface {
	Save(ctx context.Context, userID, username, ip, userAgent, sessionID string, loginTime time.Time, success bool, failReason string) error
	UpdateLogoutTime(ctx context.Context, sessionID string, logoutTime time.Time) error
	GetBySessionID(ctx context.Context, sessionID string) (*UserLogin, error)
	GetLastUserLogins(ctx context.Context, userID string, limit int) ([]*UserLogin, error)
	GetFailedLogins(ctx context.Context, username string, since time.Time) (int, error)
	CleanupOldRecords(ctx context.Context, before time.Time) (int64, error)
}

type UserLoginRepositoryImpl struct {
	db *sql.DB
}

// NewUserLoginRepository создает новый репозиторий для работы с историей входов
func NewUserLoginRepository(db *sql.DB) *UserLoginRepositoryImpl {
	return &UserLoginRepositoryImpl{db: db}
}

// Save сохраняет информацию о входе пользователя
func (r *UserLoginRepositoryImpl) Save(
	ctx context.Context,
	userID, username, ip, userAgent, sessionID string,
	loginTime time.Time,
	success bool,
	failReason string,
) error {
	var reason sql.NullString
	if failReason != "" {
		reason = sql.NullString{String: failReason, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_logins (
			user_id, username, ip, user_agent, login_time, 
			session_id, success, fail_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, username, ip, userAgent, loginTime,
		sessionID, success, reason,
	)

	if err != nil {
		return fmt.Errorf("failed to save user login: %w", err)
	}
	return nil
}

// UpdateLogoutTime обновляет время выхода пользователя
func (r *UserLoginRepositoryImpl) UpdateLogoutTime(ctx context.Context, sessionID string, logoutTime time.Time) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE user_logins SET logout_time = $1 WHERE session_id = $2`,
		logoutTime, sessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to update logout time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLoginNotFound
	}

	return nil
}

// GetBySessionID возвращает запись о входе по идентификатору сессии
func (r *UserLoginRepositoryImpl) GetBySessionID(ctx context.Context, sessionID string) (*UserLogin, error) {
	login := &UserLogin{}
	err := r.db.QueryRowContext(ctx,
		`SELECT 
			id, user_id, username, ip, user_agent, 
			login_time, logout_time, session_id, success, fail_reason
		FROM user_logins 
		WHERE session_id = $1`,
		sessionID,
	).Scan(
		&login.ID, &login.UserID, &login.Username, &login.IP, &login.UserAgent,
		&login.LoginTime, &login.LogoutTime, &login.SessionID, &login.Success, &login.FailReason,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLoginNotFound
		}
		return nil, fmt.Errorf("failed to get login by session ID: %w", err)
	}

	return login, nil
}

// GetLastUserLogins возвращает последние записи о входах пользователя
func (r *UserLoginRepositoryImpl) GetLastUserLogins(ctx context.Context, userID string, limit int) ([]*UserLogin, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
			id, user_id, username, ip, user_agent, 
			login_time, logout_time, session_id, success, fail_reason
		FROM user_logins 
		WHERE user_id = $1
		ORDER BY login_time DESC
		LIMIT $2`,
		userID, limit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query user logins: %w", err)
	}
	defer rows.Close()

	var logins []*UserLogin
	for rows.Next() {
		login := &UserLogin{}
		err := rows.Scan(
			&login.ID, &login.UserID, &login.Username, &login.IP, &login.UserAgent,
			&login.LoginTime, &login.LogoutTime, &login.SessionID, &login.Success, &login.FailReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan login record: %w", err)
		}
		logins = append(logins, login)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return logins, nil
}

// GetFailedLogins возвращает количество неудачных попыток входа для пользователя
func (r *UserLoginRepositoryImpl) GetFailedLogins(ctx context.Context, username string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) 
		FROM user_logins 
		WHERE username = $1 AND success = false AND login_time > $2`,
		username, since,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count failed logins: %w", err)
	}

	return count, nil
}

// CleanupOldRecords удаляет старые записи о входах
func (r *UserLoginRepositoryImpl) CleanupOldRecords(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_logins WHERE login_time < $1`,
		before,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old login records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
