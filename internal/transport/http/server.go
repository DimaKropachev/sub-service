package http

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/DimaKropachev/sub-service/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	Port int    `env:"HTTP_PORT" env-default:"8083"`
}

type Server struct {
	s    http.Server
}

func NewServer(cfg Config, h *Handler, m *Middleware) *Server {
	r := chi.NewRouter()

	r.Use(m.LoggingMiddleware, m.RequestIDMiddleware)

	r.Post("/subscriptions", h.AddNewSubscription)
	r.Get("/subscriptions/{id}", h.GetSubscriptionByID)
	r.Get("/subscriptions", h.GetListSubscriptions)
	r.Patch("/subscriptions/{id}", h.UpdateSubscription)
	r.Delete("/subscriptions/{id}", h.DeleteSubscriptionByID)
	r.Get("/subscriptions/cost", h.GetTotalCostSubscriptions)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return &Server{
		s: http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
			Handler: r,
		},
	}
}

func (s *Server) Run() error {
	return s.s.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
