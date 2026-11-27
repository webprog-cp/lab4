package router

import (
	"net/http"

	"cp_lab1/internal/app"
	"cp_lab1/internal/config"
	"cp_lab1/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(cfg *config.Config, app *app.App) (*gin.Engine, error) {
	gin.SetMode(cfg.GinMode)
	binding.EnableDecoderDisallowUnknownFields = true

	ge := gin.New()
	ge.Use(gin.Logger(), gin.Recovery())

	if err := ge.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	registerRoutes(ge, cfg, app)

	return ge, nil
}

func writeGuard(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ReadOnly {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "API is in read-only mode"})
			return
		}
		if c.GetHeader("X-API-Key") != cfg.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid API key"})
			return
		}
		c.Next()
	}
}

func registerCats(ge *gin.RouterGroup, ch *handlers.CatHandler, guard gin.HandlerFunc) {
	cats := ge.Group("/cats")
	cats.GET("", ch.GetAll)
	cats.GET("/:id", ch.GetByID)

	protected := cats.Group("", guard)
	protected.POST("", ch.Create)
	protected.PUT("/:id", ch.Update)
	protected.DELETE("", ch.DeleteAll)
	protected.DELETE("/:id", ch.Delete)
}

func registerShelters(ge *gin.RouterGroup, sh *handlers.ShelterHandler, guard gin.HandlerFunc) {
	shelters := ge.Group("/shelters")
	shelters.GET("", sh.GetAll)
	shelters.GET("/:id", sh.GetByID)
	shelters.GET("/:id/cats", sh.GetCats)

	protected := shelters.Group("", guard)
	protected.POST("", sh.Create)
	protected.PUT("/:id", sh.Update)
	protected.DELETE("", sh.DeleteAll)
	protected.DELETE("/:id", sh.Delete)
}

func registerRoutes(ge *gin.Engine, cfg *config.Config, app *app.App) {
	api := ge.Group(cfg.BasePath())
	guard := writeGuard(cfg)
	registerCats(api, app.CatHandler, guard)
	registerShelters(api, app.ShelterHandler, guard)
	ge.StaticFile("/openapi.yaml", "docs/openapi.yaml")
	ge.StaticFile("/", "web/index.html")
	ge.Static("/assets", "web/assets")
	ge.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))
}
