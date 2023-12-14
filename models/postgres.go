package models

import (
	"fmt"
	"io/fs"

	"github.com/go-faster/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "baloo",
		Password: "junglebook",
		Database: "lenslocked",
		SSLMode:  "disable",
	}
}

// Open will open a SQL connection with the provided
// Postgres database. Callers of Open need to ensure
// the connection is eventually closed via the
// db.Close() method.
func Open(config PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", config.String())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return db, nil
}

func IsMigUpToDate(db *sqlx.DB, fs fs.FS, dir string) (bool, error) {
	goose.SetBaseFS(fs)
	err := goose.SetDialect("postgres")
	if err != nil {
		return false, fmt.Errorf("check migration: %w", err)
	}
	version, err := goose.GetDBVersion(db.DB)
	if err != nil {
		return false, err
	}
	_, err = goose.CollectMigrations(dir, version, goose.MaxVersion)
	if errors.Is(err, goose.ErrNoMigrationFiles) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("check migration: %w", err)
	}
	return false, nil
}

func MigrateFS(db *sqlx.DB, fs fs.FS, dir string) error {
	goose.SetBaseFS(fs)
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	err = goose.Up(db.DB, dir)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
