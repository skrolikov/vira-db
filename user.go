package db

import (
	"database/sql"
)

func CreateUser(db *sql.DB, username, password string) (string, error) {
	var id string
	err := db.QueryRow(`
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id
	`, username, password).Scan(&id)
	return id, err
}
