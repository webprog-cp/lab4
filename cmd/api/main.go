package main

import (
	"log"
	"log/slog"
	"os"

	"cp_lab1/internal/app"
	"cp_lab1/internal/config"
	"cp_lab1/internal/router"

	_ "cp_lab1/docs"
)

// @title Cats shelter management API
// @version 1.0
// @description REST API service for cats shelter management.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.SlogLevel()}))
	logger.Info("config loaded", "readOnly", cfg.ReadOnly, "maxListSize", cfg.MaxListSize, "cacheTTL", cfg.CacheTTL())

	a, err := app.New(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	r, err := router.New(cfg, a)
	if err != nil {
		log.Fatal(err)
	}

	host := cfg.Host + ":" + cfg.Port

	if err := r.Run(host); err != nil {
		log.Fatal(err)
	}
}
