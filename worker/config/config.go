package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"
)

var defaultCfg = pq.Config{
	Host:           "localhost",
	Port:           5432,
	User:           "postgres",
	Password:       "postgres",
	Database:       "postgres",
	ConnectTimeout: 5 * time.Second,
	SSLMode:        pq.SSLModeDisable,
}

func GetConfig() *pq.Config {
	slog.Debug("env", os.Environ())

	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		port = 0
	}

	cfg := pq.Config{
		Host:     os.Getenv("PG_HOST"),
		Port:     uint16(port),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASS"),
		Database: os.Getenv("PG_DB"),
		SSLMode:  pq.SSLModeDisable,
	}

	if err := validateConfig(&cfg); err != nil {
		slog.Warn("Failed to validate config", "err", err)
		return &defaultCfg
	}

	return &cfg
}

func validateConfig(cfg *pq.Config) error {
	if cfg.Host == "" ||
		cfg.Port == 0 ||
		cfg.User == "" ||
		cfg.Database == "" ||
		cfg.Password == "" {
		return errors.New("some fields are missing in ENV")
	}

	return nil
}
