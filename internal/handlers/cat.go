package handlers

import (
	"context"
	"errors"
	"net/http"

	"cp_lab1/internal/cache"
	"cp_lab1/internal/const_errors"
	"cp_lab1/internal/dto"
	"cp_lab1/internal/mapper"
	"cp_lab1/internal/models"
	"cp_lab1/internal/repositories"

	"github.com/gin-gonic/gin"
)

type CatHandler struct {
	opts        Options
	cache       cache.Cache
	repo        repositories.CatRepo
	shelterRepo repositories.ShelterRepo
}

func NewCatHandler(repo repositories.CatRepo, shelterRepo repositories.ShelterRepo, cache cache.Cache, opts Options) *CatHandler {
	return &CatHandler{
		opts:        opts,
		cache:       cache,
		repo:        repo,
		shelterRepo: shelterRepo,
	}
}

// @Summary Get all cats
// @Tags cats
// @Produce json
// @Success 200 {array} dto.CatResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats [get]
func (ch *CatHandler) GetAll(c *gin.Context) {
	cats, err := ch.repo.GetAll(c.Request.Context())
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.CatResponses(limitList(c, ch.opts, cats)))
}

// @Summary Get cat by ID
// @Tags cats
// @Produce json
// @Param id path int true "Cat ID"
// @Success 200 {object} dto.CatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats/{id} [get]
func (ch *CatHandler) GetByID(c *gin.Context) {
	id, valid := idParam(c)
	if !valid {
		return
	}

	key := cacheKey(id)
	var cat models.Cat

	if cacheGet(c, ch.cache, ch.opts, key, &cat) {
		c.JSON(http.StatusOK, mapper.CatResponse(cat))
		return
	}

	cat, err := ch.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	cacheSet(c, ch.cache, ch.opts, key, cat)

	c.JSON(http.StatusOK, mapper.CatResponse(cat))
}

// @Summary Create cat
// @Tags cats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param cat body dto.CatRequest true "Cat"
// @Success 201 {object} dto.CatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats [post]
func (ch *CatHandler) Create(c *gin.Context) {
	var request dto.CatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: bindErrorMessage(err, &request)})
		return
	}

	if err := ch.shelterHasCapacity(c.Request.Context(), request.ShelterID); err != nil {
		respondRepositoryError(c, err)
		return
	}

	cat, err := ch.repo.Create(c.Request.Context(), mapper.CatFromRequest(request))
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, mapper.CatResponse(cat))
}

// @Summary Update cat
// @Tags cats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Param cat body dto.CatRequest true "Cat"
// @Success 200 {object} dto.CatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats/{id} [put]
func (ch *CatHandler) Update(c *gin.Context) {
	id, valid := idParam(c)
	if !valid {
		return
	}

	var request dto.CatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: bindErrorMessage(err, &request)})
		return
	}

	existing, err := ch.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}
	if existing.ShelterID != request.ShelterID {
		if err := ch.shelterHasCapacity(c.Request.Context(), request.ShelterID); err != nil {
			respondRepositoryError(c, err)
			return
		}
	}

	cat := mapper.CatFromRequest(request)
	if err := ch.repo.Update(c.Request.Context(), id, cat); err != nil {
		respondRepositoryError(c, err)
		return
	}

	cat.ID = id
	key := cacheKey(id)
	ctx, cancel := cacheContext(c)
	if err := ch.cache.Invalidate(ctx, key); err != nil {
		logCacheError(c, "invalidate", key, err)
	}
	cancel()

	c.JSON(http.StatusOK, mapper.CatResponse(cat))
}

func (ch *CatHandler) shelterHasCapacity(ctx context.Context, shelterID uint64) error {
	shelter, err := ch.shelterRepo.GetByID(ctx, shelterID)
	if errors.Is(err, const_errors.NoEntityByID) {
		return const_errors.InvalidReference
	}
	if err != nil {
		return err
	}

	cats, err := ch.repo.GetCatsByShelterID(ctx, shelterID)
	if err != nil {
		return err
	}
	if uint64(len(cats)) >= shelter.Capacity {
		return const_errors.ShelterAtCapacity
	}

	return nil
}

// @Summary Delete cat
// @Tags cats
// @Security ApiKeyAuth
// @Param id path int true "Cat ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats/{id} [delete]
func (ch *CatHandler) Delete(c *gin.Context) {
	id, valid := idParam(c)
	if !valid {
		return
	}

	if err := ch.repo.Delete(c.Request.Context(), id); err != nil {
		respondRepositoryError(c, err)
		return
	}

	key := cacheKey(id)
	ctx, cancel := cacheContext(c)
	if err := ch.cache.Invalidate(ctx, key); err != nil {
		logCacheError(c, "invalidate", key, err)
	}
	cancel()

	c.Status(http.StatusNoContent)
}

// @Summary Delete all cats
// @Tags cats
// @Security ApiKeyAuth
// @Success 204
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cats [delete]
func (ch *CatHandler) DeleteAll(c *gin.Context) {
	if err := ch.repo.DeleteAll(c.Request.Context()); err != nil {
		respondRepositoryError(c, err)
		return
	}

	ctx, cancel := cacheContext(c)
	if err := ch.cache.InvalidateAll(ctx); err != nil {
		logCacheError(c, "invalidate all", "*", err)
	}
	cancel()

	c.Status(http.StatusNoContent)
}
