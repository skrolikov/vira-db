package db

import (
	"context"
	"database/sql"
	"time"
)

type UserLoginRepository struct {
	db *sql.DB
}

func NewUserLoginRepository(db *sql.DB) *UserLoginRepository {
	return &UserLoginRepository{db: db}
}

func (r *UserLoginRepository) Save(ctx context.Context, userID, username, ip, userAgent string, loginTime time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_logins (user_id, username, ip, user_agent, login_time) VALUES ($1, $2, $3, $4, $5)`,
		userID, username, ip, userAgent, loginTime)
	return err
}
