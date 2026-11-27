package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"cp_lab1/internal/cache"
	"cp_lab1/internal/const_errors"
	"cp_lab1/internal/dto"

	"github.com/gin-gonic/gin"
)

const cacheTimeout = time.Second

type Options struct {
	CacheTTL    time.Duration
	MaxListSize int
}

func cacheGet(c *gin.Context, cc cache.Cache, opts Options, key string, dest any) bool {
	if opts.CacheTTL <= 0 {
		c.Header("X-Cache", "DISABLED")
		return false
	}

	ctx, cancel := cacheContext(c)
	defer cancel()
	found, err := cc.Get(ctx, key, dest)
	if err != nil {
		logCacheError(c, "get", key, err)
		found = false
	}

	if found {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	return found
}

func cacheSet(c *gin.Context, cc cache.Cache, opts Options, key string, value any) {
	if opts.CacheTTL <= 0 {
		return
	}

	ctx, cancel := cacheContext(c)
	defer cancel()
	if err := cc.Set(ctx, key, value, opts.CacheTTL); err != nil {
		logCacheError(c, "set", key, err)
	}
}

func limitList[T any](c *gin.Context, opts Options, items []T) []T {
	c.Header("X-Total-Count", strconv.Itoa(len(items)))
	if opts.MaxListSize > 0 && len(items) > opts.MaxListSize {
		return items[:opts.MaxListSize]
	}
	return items
}

func cacheContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), cacheTimeout)
}

func cacheKey(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func logCacheError(c *gin.Context, operation, key string, err error) {
	log.Printf("%s %s: cache %s key %q: %v", c.Request.Method, c.Request.URL.Path, operation, key, err)
}

func idParam(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "id must be an unsigned integer"})
		return 0, false
	}
	return id, true
}

func respondRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, const_errors.NoEntityByID):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "resource not found"})
	case errors.Is(err, const_errors.InvalidReference):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "related resource not found"})
	case errors.Is(err, const_errors.EntityHasDependents):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "resource has dependent records"})
	case errors.Is(err, const_errors.ShelterAtCapacity):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "shelter is at full capacity"})
	case errors.Is(err, const_errors.CapacityBelowUsage):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "capacity is below current occupancy"})
	default:
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}
