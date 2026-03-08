package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaKropachev/sub-service/internal/config"
	"github.com/DimaKropachev/sub-service/internal/repository"
	"github.com/DimaKropachev/sub-service/internal/service"
	"github.com/DimaKropachev/sub-service/internal/transport/http"
	"github.com/DimaKropachev/sub-service/pkg/postgres"
	"go.uber.org/zap"
)

type App struct {
	cfg *config.Config
	log *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Start() {
	db, err := postgres.New(a.cfg.DB)
	if err != nil {
		a.log.Fatal("failed to connect database", zap.Error(err))
	}
	a.log.Info("database connect successfully")

	repo := repository.New(db.DB)
	a.log.Info("repository initialized")

	service := service.New(repo, a.log)
	a.log.Info("service initialized")

	handler := http.NewHandler(service, a.log)
	middleware := http.NewMiddleware(a.cfg.HTTP, a.log)
	server := http.NewServer(a.cfg.HTTP, handler, middleware)
	a.log.Info("http server initialized")

	serverCtx, cancelServCtx := context.WithCancel(context.Background())
	defer cancelServCtx()

	go func() {
		a.log.Info("http server starting...", zap.String("addr", fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)))
		if err = server.Run(); err != nil {
			a.log.Error("failed to start http server", zap.Error(err))
			cancelServCtx()
		}
	}()

	// Graceful Shutdown

	graceSh := make(chan os.Signal, 1)
	signal.Notify(graceSh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sign := <-graceSh:
		a.log.Info("Shutdown signal received, starting graceful shutdown...", zap.String("signal", sign.String()))
	case <-serverCtx.Done():
		a.log.Info("critical server error, initialized graceful shutdown...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Stop(shutdownCtx); err != nil {
		a.log.Error("error shutting down http server", zap.Error(err))
	} else {
		a.log.Info("http server gracefully stopped")
	}

	if err := db.Close(); err != nil {
		a.log.Warn("failed to close database connection", zap.Error(err))
	} else {
		a.log.Info("database connection closed successfully")
	}

	a.log.Info("application stopped gracefully")
}
