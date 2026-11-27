package repositories

import (
	"context"

	"cp_lab1/internal/models"
)

type CatRepo interface {
	GetAll(context.Context) ([]models.Cat, error)
	GetByID(context.Context, uint64) (models.Cat, error)
	GetCatsByShelterID(context.Context, uint64) ([]models.Cat, error)
	Create(context.Context, models.Cat) (models.Cat, error)
	Update(context.Context, uint64, models.Cat) error
	Delete(context.Context, uint64) error
	DeleteAll(context.Context) error
}
