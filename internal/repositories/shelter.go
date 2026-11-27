package repositories

import (
	"context"

	"cp_lab1/internal/models"
)

type ShelterRepo interface {
	GetAll(context.Context) ([]models.Shelter, error)
	GetByID(context.Context, uint64) (models.Shelter, error)
	Create(context.Context, models.Shelter) (models.Shelter, error)
	Update(context.Context, uint64, models.Shelter) error
	Delete(context.Context, uint64) error
	DeleteAll(context.Context) error
}
