package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	config "github.com/skrolikov/vira-config"
	logger "github.com/skrolikov/vira-logger"
)

var (
	instance *sql.DB
	once     sync.Once
	mu       sync.RWMutex
	logg     *logger.Logger // Кастомный логгер
)

// SetLogger задаёт логгер для пакета db
func SetLogger(l *logger.Logger) {
	logg = l
}

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
			initErr = fmt.Errorf("не удалось открыть соединение с БД: %w", err)
			if logg != nil {
				logg.Error("не удалось открыть соединение с БД: %v", err)
			}
			return
		}

		conn.SetMaxOpenConns(cfg.DBMaxOpenConns)
		conn.SetMaxIdleConns(cfg.DBMaxIdleConns)
		conn.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
		conn.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err = conn.PingContext(pingCtx); err != nil {
			_ = conn.Close()
			initErr = fmt.Errorf("не удалось пропинговать БД: %w", err)
			if logg != nil {
				logg.Error("не удалось пропинговать БД: %v", err)
			}
			return
		}

		instance = conn

		if logg != nil {
			logg.Info("✅ Соединение с базой данных установлено успешно")
		}

		// Запуск мониторинга в отдельной горутине
		go monitorConnection(ctx, 30*time.Second)
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
		if logg != nil {
			logg.Error("не удалось закрыть соединение с БД: %v", err)
		}
		return fmt.Errorf("не удалось закрыть соединение с БД: %w", err)
	}

	instance = nil

	if logg != nil {
		logg.Info("🔌 Соединение с базой данных закрыто")
	}

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

// monitorConnection периодически проверяет соединение с БД и логгирует ошибки
func monitorConnection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if logg != nil {
				logg.Info("🔕 Мониторинг соединения с БД остановлен")
			}
			return
		case <-ticker.C:
			if err := HealthCheck(ctx); err != nil {
				if logg != nil {
					logg.Warn("⚠️ Ошибка проверки здоровья БД: %v", err)
				}
				// Здесь можно добавить логику восстановления соединения или алерты
			} else {
				if logg != nil {
					logg.Debug("✔️ Проверка здоровья БД успешна")
				}
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
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // повторно вызываем панику после отката
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			err = fmt.Errorf("ошибка транзакции: %v, ошибка отката: %w", err, rbErr)
			return err
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	return nil
}
