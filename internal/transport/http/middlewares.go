package http

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ctxKey string

const RequestIDKey ctxKey = "x-request-id"

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(statusCode int) {
	sw.status = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

type Middleware struct {
	cfg Config
	log *zap.Logger
}

func NewMiddleware(cfg Config, log *zap.Logger) *Middleware {
	return &Middleware{
		cfg: cfg,
		log: log,
	}
}

func (m *Middleware) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := &statusWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(sw, r)

		m.log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", sw.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
