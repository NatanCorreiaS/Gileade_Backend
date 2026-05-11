package db_test

import (
	"context"
	"testing"

	"gileade/gileade_backend/db"
	"gileade/gileade_backend/internal/testutil"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm/logger"
)

// TestOpenPostgres valida abertura de conexao com Postgres.
func TestOpenPostgres(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	container := tdb.Container
	ctx := context.Background()

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
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

	dbConn, err := db.OpenPostgres(cfg)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	if dbConn == nil {
		t.Fatalf("dbConn nil")
	}
}

// Garante que a imagem do módulo existe no go.mod/go.sum.
var _ *postgres.PostgresContainer
