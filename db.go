package db

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	config "github.com/skrolikov/vira-config"
)

var (
	db   *sql.DB
	once sync.Once
)

// Init инициализирует соединение с БД один раз
func Init(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	var err error
	once.Do(func() {
		db, err = sql.Open("postgres", cfg.DBUrl)
		if err != nil {
			return
		}

		// Настройки пула можно расширить позже
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * time.Minute)

		if err = db.PingContext(ctx); err != nil {
			return
		}

		log.Println("✅ Подключено к БД!")
	})
	return db, err
}

// Get возвращает активное соединение
func Get() *sql.DB {
	if db == nil {
		log.Fatal("❌ DB не инициализирована. Вызовите Init() перед Get().")
	}
	return db
}
