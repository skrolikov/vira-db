package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(dbURL string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("БД не отвечает:", err)
	}

	log.Println("Подключено к БД!")
	return db
}
