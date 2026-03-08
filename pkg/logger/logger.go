package logger

import (
	"context"

	"go.uber.org/zap"
)

func New(ctx context.Context, env string) (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	switch env {
	case "dev":
		l, err = zap.NewDevelopment()
	case "prod":
		l, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	return l, nil
}
