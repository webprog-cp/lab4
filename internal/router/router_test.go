package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "cp_lab1/internal/app"
	"cp_lab1/internal/config"
	"cp_lab1/internal/handlers"
	"cp_lab1/internal/models"
	"cp_lab1/internal/repositories"

	"github.com/gin-gonic/gin"
)

type mockCache struct{}

func (mockCache) Set(context.Context, string, any, time.Duration) error { return nil }
func (mockCache) Get(context.Context, string, any) (bool, error)        { return false, nil }
func (mockCache) Invalidate(context.Context, string) error              { return nil }
func (mockCache) InvalidateAll(context.Context) error                   { return nil }

type panicCatRepo struct {
	repositories.CatRepo
}

func (panicCatRepo) GetAll(context.Context) ([]models.Cat, error) {
	panic("test panic")
}

func TestRecoveryMiddlewareCoversRegisteredRoutes(t *testing.T) {
	catRepo := panicCatRepo{}
	shelterRepo := repositories.NewShelterRepoMemory()
	a := &application.App{
		CatHandler:     handlers.NewCatHandler(catRepo, shelterRepo, mockCache{}, testOpts),
		ShelterHandler: handlers.NewShelterHandler(shelterRepo, catRepo, mockCache{}, testOpts),
	}
	r, err := New(testConfig(), a)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/cats", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

var testOpts = handlers.Options{CacheTTL: time.Minute, MaxListSize: 100}

const apiKey = "test-key"

func testConfig() *config.Config {
	return &config.Config{GinMode: gin.TestMode, APIPrefix: "/api", APIVersion: "v1", APIKey: apiKey}
}

func TestWriteGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig()
	r := gin.New()
	r.POST("/protected", writeGuard(cfg), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	send := func(key string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/protected", nil)
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := send(""); code != http.StatusUnauthorized {
		t.Fatalf("status without key = %d, want %d", code, http.StatusUnauthorized)
	}
	if code := send(apiKey); code != http.StatusNoContent {
		t.Fatalf("status with key = %d, want %d", code, http.StatusNoContent)
	}

	cfg.ReadOnly = true
	if code := send(apiKey); code != http.StatusForbidden {
		t.Fatalf("status in read-only mode = %d, want %d", code, http.StatusForbidden)
	}
}

func TestMaxListSizeLimitsResponse(t *testing.T) {
	catRepo := repositories.NewCatRepoMemory()
	shelterRepo := repositories.NewShelterRepoMemory()
	for i := 0; i < 3; i++ {
		if _, err := shelterRepo.Create(context.Background(), models.Shelter{Name: "S", City: "Kyiv", Address: "A", Capacity: 1}); err != nil {
			t.Fatal(err)
		}
	}
	opts := handlers.Options{MaxListSize: 2}
	a := &application.App{
		CatHandler:     handlers.NewCatHandler(catRepo, shelterRepo, mockCache{}, opts),
		ShelterHandler: handlers.NewShelterHandler(shelterRepo, catRepo, mockCache{}, opts),
	}
	r, err := New(testConfig(), a)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/shelters", nil))
	if got := strings.Count(w.Body.String(), `"id"`); got != 2 {
		t.Fatalf("returned %d shelters, want 2; body = %s", got, w.Body.String())
	}
	if got := w.Header().Get("X-Total-Count"); got != "3" {
		t.Fatalf("X-Total-Count = %q, want 3", got)
	}
}
func TestRequestDTOValidation(t *testing.T) {
	catRepo := repositories.NewCatRepoMemory()
	shelterRepo := repositories.NewShelterRepoMemory()
	if _, err := shelterRepo.Create(context.Background(), models.Shelter{Name: "Test", City: "Kyiv", Address: "Test", Capacity: 1}); err != nil {
		t.Fatal(err)
	}
	a := &application.App{
		CatHandler:     handlers.NewCatHandler(catRepo, shelterRepo, mockCache{}, testOpts),
		ShelterHandler: handlers.NewShelterHandler(shelterRepo, catRepo, mockCache{}, testOpts),
	}
	r, err := New(testConfig(), a)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		body string
		want int
		msg  string
	}{
		{
			name: "cat accepts zero months",
			path: "/api/v1/cats",
			body: `{"name":"Kitten","ageMonths":0,"gender":"female","adopted":false,"shelterID":1}`,
			want: http.StatusCreated,
		},
		{
			name: "cat rejects client ID",
			path: "/api/v1/cats",
			body: `{"id":42,"name":"Kitten","ageMonths":0,"gender":"female","adopted":false,"shelterID":1}`,
			want: http.StatusBadRequest,
			msg:  `unknown field \"id\"`,
		},
		{
			name: "cat rejects excessive age",
			path: "/api/v1/cats",
			body: `{"name":"Old cat","ageMonths":361,"gender":"male","adopted":false,"shelterID":1}`,
			want: http.StatusBadRequest,
			msg:  "ageMonths must be at most 360",
		},
		{
			name: "cat rejects invalid gender",
			path: "/api/v1/cats",
			body: `{"name":"Cat","ageMonths":12,"gender":"unknown","adopted":false,"shelterID":1}`,
			want: http.StatusBadRequest,
			msg:  "gender must be one of: male, female",
		},
		{
			name: "shelter rejects zero capacity",
			path: "/api/v1/shelters",
			body: `{"name":"Shelter","city":"Kyiv","address":"Main Street","capacity":0}`,
			want: http.StatusBadRequest,
			msg:  "capacity is required and must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", apiKey)
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tt.want, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tt.msg) {
				t.Fatalf("body = %s, want it to contain %q", w.Body.String(), tt.msg)
			}
		})
	}
}
