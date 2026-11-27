package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"
)

const defaultPath = "config.json"

type Config struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	GinMode    string `json:"ginMode"`
	APIPrefix  string `json:"apiPrefix"`
	APIVersion string `json:"apiVersion"`

	APIKey          string `json:"apiKey"`
	ReadOnly        bool   `json:"readOnly"`
	MaxListSize     int    `json:"maxListSize"`
	CacheTTLSeconds int    `json:"cacheTTLSeconds"`

	LogLevel string `json:"logLevel"`

	PostgresDSN string `json:"-"`
	RedisDSN    string `json:"-"`
}

func (c *Config) BasePath() string {
	return c.APIPrefix + "/" + c.APIVersion
}

func (c *Config) CacheTTL() time.Duration {
	return time.Duration(c.CacheTTLSeconds) * time.Second
}

func (c *Config) SlogLevel() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		return slog.LevelInfo
	}
	return level
}

func defaults() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            "8080",
		GinMode:         "debug",
		APIPrefix:       "/api",
		APIVersion:      "v1",
		APIKey:          "demo",
		MaxListSize:     100,
		CacheTTLSeconds: 300,
		LogLevel:        "info",
	}
}

func Load() (*Config, error) {
	cfg := defaults()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist) && os.Getenv("CONFIG_PATH") == "":
	case err != nil:
		return nil, fmt.Errorf("read config %q: %w", path, err)
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %q: %w", path, err)
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config %q: %w", path, err)
	}

	cfg.PostgresDSN = (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")),
		Host:     "postgres:5432",
		Path:     os.Getenv("POSTGRES_DB"),
		RawQuery: "sslmode=disable",
	}).String()
	cfg.RedisDSN = (&url.URL{
		Scheme: "redis",
		User:   url.UserPassword(os.Getenv("REDIS_USER"), os.Getenv("REDIS_USER_PASSWORD")),
		Host:   "redis:6379",
		Path:   os.Getenv("REDIS_DB"),
	}).String()

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.APIKey == "" {
		return errors.New("apiKey must not be empty")
	}
	if c.MaxListSize < 1 {
		return errors.New("maxListSize must be greater than zero")
	}
	if c.CacheTTLSeconds < 0 {
		return errors.New("cacheTTLSeconds must not be negative")
	}
	return nil
}
