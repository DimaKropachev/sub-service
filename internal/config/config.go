package config

import (
	"github.com/DimaKropachev/sub-service/internal/transport/http"
	"github.com/DimaKropachev/sub-service/pkg/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string `env:"ENVIROMENT" env-default:"dev"`
	DB   postgres.Config 
	HTTP http.Config
}

func ParseConfig(path string) (*Config, error) {
	cfg := &Config{}

	if path != "" {
		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			return nil, err
		}
	} else {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
