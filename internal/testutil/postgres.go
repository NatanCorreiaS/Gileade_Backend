package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresTestDB struct {
	DB        *gorm.DB
	Container *postgres.PostgresContainer
}

// StartPostgres sobe um container Postgres para testes.
func StartPostgres(t *testing.T) PostgresTestDB {
	t.Helper()

	ctx := context.Background()
	container, err := postgres.Run(
		ctx,
		"postgres:18.3",
		postgres.WithDatabase("gileade_test"),
		postgres.WithUsername("gileade"),
		postgres.WithPassword("gileade"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("postgres connection string: %v", err)
	}

	db, err := gorm.Open(
		pgdriver.Open(connStr),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}

	return PostgresTestDB{DB: db, Container: container}
}
