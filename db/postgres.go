package db

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
	LogLevel logger.LogLevel
}

// NewPostgresConfigFromEnv monta a config do Postgres a partir do ambiente.
func NewPostgresConfigFromEnv() (PostgresConfig, error) {
	host, err := requiredEnv("DB_HOST")
	if err != nil {
		return PostgresConfig{}, err
	}
	user, err := requiredEnv("DB_USER")
	if err != nil {
		return PostgresConfig{}, err
	}
	password, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return PostgresConfig{}, err
	}
	dbName, err := requiredEnv("DB_NAME")
	if err != nil {
		return PostgresConfig{}, err
	}
	portStr, err := requiredEnv("DB_PORT")
	if err != nil {
		return PostgresConfig{}, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return PostgresConfig{}, fmt.Errorf("DB_PORT inválida: %w", err)
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timeZone := os.Getenv("DB_TIMEZONE")
	if timeZone == "" {
		timeZone = "UTC"
	}

	return PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
		TimeZone: timeZone,
		LogLevel: logger.Silent,
	}, nil
}

// OpenPostgres abre uma conexao Gorm com Postgres.
func OpenPostgres(cfg PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
		cfg.TimeZone,
	)

	gcfg := &gorm.Config{}
	if cfg.LogLevel != 0 {
		gcfg.Logger = logger.Default.LogMode(cfg.LogLevel)
	}

	db, err := gorm.Open(postgres.Open(dsn), gcfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Defaults conservadores; podem ser ajustados via código chamador.
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// requiredEnv valida a existencia de uma variavel de ambiente obrigatoria.
func requiredEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", errors.New("variável de ambiente obrigatória não definida: " + key)
	}
	return v, nil
}
