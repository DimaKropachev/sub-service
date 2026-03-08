package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	UserName string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	DBName   string `env:"POSTGRES_DB" env-default:"postgres"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
}

type DataBase struct {
	DB *sql.DB
}

func New(cfg Config) (*DataBase, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DataBase{
		DB: db,
	}, nil
}

func (db *DataBase) Close() error {
	if err := db.DB.Close(); err != nil {
		return err
	}
	return nil
}
