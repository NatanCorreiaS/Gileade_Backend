package testutil

import (
	"context"
	"testing"
	"time"

	"gileade/gileade_backend/db"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestDB struct {
	DB        *gorm.DB
	Container *postgres.PostgresContainer
}

// StartPostgres sobe um Postgres de teste e abre a conexao Gorm.
func StartPostgres(t *testing.T) TestDB {
	t.Helper()

	ctx := context.Background()
	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("gileade_test"),
		postgres.WithUsername("gileade"),
		postgres.WithPassword("gileade"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("port: %v", err)
	}

	cfg := db.PostgresConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "gileade",
		Password: "gileade",
		DBName:   "gileade_test",
		SSLMode:  "disable",
		TimeZone: "UTC",
		LogLevel: logger.Silent,
	}

	var dbConn *gorm.DB
	var lastErr error
	for i := 0; i < 10; i++ {
		dbConn, err = db.OpenPostgres(cfg)
		if err == nil {
			lastErr = nil
			break
		}
		lastErr = err
		time.Sleep(300 * time.Millisecond)
	}
	if lastErr != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("OpenPostgres: %v", lastErr)
	}

	sqlDB, err := dbConn.DB()
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("DB: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = container.Terminate(ctx)
	})

	return TestDB{DB: dbConn, Container: container}
}
