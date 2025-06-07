package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Connect открывает соединение с базой
func Connect(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Подключено к БД!")
	return db, nil
}
