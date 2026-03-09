package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/DimaKropachev/sub-service/internal/config"
	"github.com/DimaKropachev/sub-service/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func main() {
	var (
		migrationPath string
		command       string
		steps         int
		version       int
	)

	flag.StringVar(&migrationPath, "path", "./migrations", "path to migration directory")
	flag.StringVar(&command, "command", "up", "migration command: up, down, force, version")
	flag.IntVar(&steps, "step", 0, "number of migration steps")
	flag.IntVar(&version, "version", 0, "migration version (used with force)")
	flag.Parse()

	envPath := os.Getenv("ENV_PATH")
	cfg, err := config.ParseConfig(envPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
	}

	l, err := logger.New(context.Background(), cfg.Env)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.UserName,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSLMode,
	)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		connStr,
	)
	if err != nil {
		l.Fatal("failed to create instance migrate", zap.Error(err))
	}
	defer m.Close()

	switch command {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal("migrations up failed", zap.Error(err))
		}
		l.Info("migrations applied successfully!")
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			l.Fatal("down without -steps is not allowed (dangerous)")
		}
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal("migrations down failed", zap.Error(err))
		}
		l.Info("migrations rolled back successfully!")
	case "force":
		if version <= 0 {
			l.Fatal("force requires -version > 0")
		}
		if err = m.Force(version); err != nil {
			l.Fatal("force failed", zap.Error(err))
		}
		l.Info("database version forcibly set", zap.Int("version", version))
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			l.Info("no migrations applied yet")
			return
		}
		if err != nil {
			l.Fatal("failed to get version", zap.Error(err))
		}
		l.Info("version recieved", zap.Uint("Current_version", v), zap.Bool("Dirty_version", dirty))
		if dirty {
			l.Warn("database is in dirty state")
		}
	}
}
