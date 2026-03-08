package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/DimaKropachev/sub-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
		log.Fatalf("failed to create instance migrate: %v\n", err)
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
			log.Fatalf("migrations up failed: %v\n", err)
		}
		log.Println("migrations applied successfully!")
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			log.Fatalln("down without -steps is not allowed (dangerous)")
		}
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrtions down failed: %v\n", err)
		}
		log.Println("migrations rolled back successfully!")
	case "force":
		if version <= 0 {
			log.Fatalln("force requires -version > 0")
		}
		if err = m.Force(version); err != nil {
			log.Fatalf("force failed: %v\n", err)
		}
		log.Printf("database version forcibly set to %d\n", version)
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Println("no migrations applied yet")
			return
		}
		if err != nil {
			log.Fatalf("failed to get version: %v\n", err)
		}

		log.Printf("Current version: %d, Dirty: %v\n", v, dirty)
		if dirty {
			log.Println("WARNING: database is in dirty state")
		}
	}
}
