package app

import (
	"fmt"
	"log/slog"

	"cp_lab1/internal/cache"
	"cp_lab1/internal/config"
	h "cp_lab1/internal/handlers"
	"cp_lab1/internal/models"
	r "cp_lab1/internal/repositories"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type App struct {
	CatHandler     *h.CatHandler
	ShelterHandler *h.ShelterHandler
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{TranslateError: true, Logger: gormlogger.Discard})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := db.AutoMigrate(&models.Shelter{}, &models.Cat{}); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	opt, err := redis.ParseURL(cfg.RedisDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	rc := redis.NewClient(opt)

	catRepo := r.NewLoggedCatRepo(r.NewCatRepoPsql(db), logger)
	shelterRepo := r.NewLoggedShelterRepo(r.NewShelterRepoPsql(db), logger)
	opts := h.Options{CacheTTL: cfg.CacheTTL(), MaxListSize: cfg.MaxListSize}

	catCache := cache.NewRedis(rc, "cat")
	shelterCache := cache.NewRedis(rc, "shelter")

	catHandler := h.NewCatHandler(catRepo, shelterRepo, catCache, opts)
	shelterHandler := h.NewShelterHandler(shelterRepo, catRepo, shelterCache, opts)

	return &App{
		CatHandler:     catHandler,
		ShelterHandler: shelterHandler,
	}, nil
}
