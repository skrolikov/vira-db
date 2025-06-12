package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	config "github.com/skrolikov/vira-config"
)

var (
	instance *sql.DB
	once     sync.Once
	mu       sync.RWMutex
)

// DBStats представляет статистику подключения к БД
type DBStats struct {
	MaxOpenConnections int           `json:"max_open"`
	OpenConnections    int           `json:"open"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// Init инициализирует соединение с БД (singleton)
func Init(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	var initErr error

	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		conn, err := sql.Open("postgres", cfg.DBUrl)
		if err != nil {
			initErr = fmt.Errorf("failed to open DB connection: %w", err)
			return
		}

		// Настройки пула соединений
		conn.SetMaxOpenConns(cfg.DBMaxOpenConns)       // Максимум открытых соединений
		conn.SetMaxIdleConns(cfg.DBMaxIdleConns)       // Максимум бездействующих соединений
		conn.SetConnMaxLifetime(cfg.DBConnMaxLifetime) // Максимальное время жизни соединения
		conn.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime) // Максимальное время бездействия

		// Проверка соединения с таймаутом
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err = conn.PingContext(pingCtx); err != nil {
			_ = conn.Close()
			initErr = fmt.Errorf("failed to ping DB: %w", err)
			return
		}

		instance = conn
		log.Println("✅ Database connection established successfully")

	})

	return instance, initErr
}

// Get возвращает активное соединение с БД
func Get() (*sql.DB, error) {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return nil, ErrDBNotInitialized
	}
	return instance, nil
}

// Close безопасно закрывает соединение с БД
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return nil
	}

	if err := instance.Close(); err != nil {
		return fmt.Errorf("failed to close DB connection: %w", err)
	}

	instance = nil
	log.Println("🔌 Database connection closed")
	return nil
}

// Stats возвращает статистику по подключению к БД
func Stats() (*DBStats, error) {
	db, err := Get()
	if err != nil {
		return nil, err
	}

	stats := db.Stats()
	return &DBStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck проверяет состояние подключения
func HealthCheck(ctx context.Context) error {
	db, err := Get()
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("%w: %v", ErrDBConnectionLost, err)
	}
	return nil
}

// monitorConnection периодически проверяет соединение с БД
func monitorConnection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := HealthCheck(ctx); err != nil {
				log.Printf("⚠️ Database health check failed: %v", err)
				// Здесь можно добавить логику восстановления соединения
			}
		}
	}
}

// WithTransaction выполняет операции в транзакции
func WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	db, err := Get()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
