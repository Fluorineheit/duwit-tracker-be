package server

import (
	"net/http"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Welcome to CatatDuit API", gin.H{
			"app": cfg.AppName,
			"env": cfg.AppEnv,
		})
	})

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			response.Success(c, http.StatusOK, "API is healthy", gin.H{
				"status": "ok",
			})
		})
	}

	return router
}