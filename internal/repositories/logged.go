package repositories

import (
	"context"
	"log/slog"
	"time"

	"cp_lab1/internal/models"
)

func logCall(ctx context.Context, log *slog.Logger, repo, method string, start time.Time, err error, args ...any) {
	attrs := append([]any{"repo", repo, "method", method, "duration", time.Since(start)}, args...)
	if err != nil {
		log.WarnContext(ctx, "repository call failed", append(attrs, "error", err)...)
		return
	}
	log.InfoContext(ctx, "repository call", attrs...)
}

type LoggedCatRepo struct {
	next CatRepo
	log  *slog.Logger
}

var _ CatRepo = (*LoggedCatRepo)(nil)

func NewLoggedCatRepo(next CatRepo, log *slog.Logger) *LoggedCatRepo {
	return &LoggedCatRepo{next: next, log: log}
}

func (r *LoggedCatRepo) GetAll(ctx context.Context) (cats []models.Cat, err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "GetAll", start, err, "count", len(cats)) }(time.Now())
	return r.next.GetAll(ctx)
}

func (r *LoggedCatRepo) GetByID(ctx context.Context, id uint64) (cat models.Cat, err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "GetByID", start, err, "id", id) }(time.Now())
	return r.next.GetByID(ctx, id)
}

func (r *LoggedCatRepo) GetCatsByShelterID(ctx context.Context, shelterID uint64) (cats []models.Cat, err error) {
	defer func(start time.Time) {
		logCall(ctx, r.log, "cat", "GetCatsByShelterID", start, err, "shelterID", shelterID, "count", len(cats))
	}(time.Now())
	return r.next.GetCatsByShelterID(ctx, shelterID)
}

func (r *LoggedCatRepo) Create(ctx context.Context, cat models.Cat) (created models.Cat, err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "Create", start, err, "id", created.ID) }(time.Now())
	return r.next.Create(ctx, cat)
}

func (r *LoggedCatRepo) Update(ctx context.Context, id uint64, cat models.Cat) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "Update", start, err, "id", id) }(time.Now())
	return r.next.Update(ctx, id, cat)
}

func (r *LoggedCatRepo) Delete(ctx context.Context, id uint64) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "Delete", start, err, "id", id) }(time.Now())
	return r.next.Delete(ctx, id)
}

func (r *LoggedCatRepo) DeleteAll(ctx context.Context) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "cat", "DeleteAll", start, err) }(time.Now())
	return r.next.DeleteAll(ctx)
}

type LoggedShelterRepo struct {
	next ShelterRepo
	log  *slog.Logger
}

var _ ShelterRepo = (*LoggedShelterRepo)(nil)

func NewLoggedShelterRepo(next ShelterRepo, log *slog.Logger) *LoggedShelterRepo {
	return &LoggedShelterRepo{next: next, log: log}
}

func (r *LoggedShelterRepo) GetAll(ctx context.Context) (shelters []models.Shelter, err error) {
	defer func(start time.Time) {
		logCall(ctx, r.log, "shelter", "GetAll", start, err, "count", len(shelters))
	}(time.Now())
	return r.next.GetAll(ctx)
}

func (r *LoggedShelterRepo) GetByID(ctx context.Context, id uint64) (shelter models.Shelter, err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "shelter", "GetByID", start, err, "id", id) }(time.Now())
	return r.next.GetByID(ctx, id)
}

func (r *LoggedShelterRepo) Create(ctx context.Context, shelter models.Shelter) (created models.Shelter, err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "shelter", "Create", start, err, "id", created.ID) }(time.Now())
	return r.next.Create(ctx, shelter)
}

func (r *LoggedShelterRepo) Update(ctx context.Context, id uint64, shelter models.Shelter) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "shelter", "Update", start, err, "id", id) }(time.Now())
	return r.next.Update(ctx, id, shelter)
}

func (r *LoggedShelterRepo) Delete(ctx context.Context, id uint64) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "shelter", "Delete", start, err, "id", id) }(time.Now())
	return r.next.Delete(ctx, id)
}

func (r *LoggedShelterRepo) DeleteAll(ctx context.Context) (err error) {
	defer func(start time.Time) { logCall(ctx, r.log, "shelter", "DeleteAll", start, err) }(time.Now())
	return r.next.DeleteAll(ctx)
}
