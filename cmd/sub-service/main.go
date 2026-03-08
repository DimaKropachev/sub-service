package main

import (
	"context"
	"log"
	"os"

	"github.com/DimaKropachev/sub-service/internal/app"
	"github.com/DimaKropachev/sub-service/internal/config"
	"github.com/DimaKropachev/sub-service/pkg/logger"
)

// @title Subscriptions Service API
// @version 1.0
// @description API для управления подписками пользователей. Позволяет создавать, получать, обновлять и удалять подписки, а также считать их общую стоимость.
func main() {
	envPath := os.Getenv("ENV_PATH")

	cfg, err := config.ParseConfig(envPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	l, err := logger.New(context.Background(), cfg.Env)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	a := app.New(cfg, l)
	a.Start()
}
