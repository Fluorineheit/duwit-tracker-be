package server

import (
	"net/http"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/Fluorineheit/duwit-tracker-be/internal/modules/categories"
	"github.com/Fluorineheit/duwit-tracker-be/internal/modules/expenses"
	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(cfg config.Config, db *pgxpool.Pool) *gin.Engine {
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

		api.GET("/health/database", func(c *gin.Context) {
			if err := db.Ping(c.Request.Context()); err != nil {
				response.Error(c, http.StatusInternalServerError, "Database connection failed")
				return
			}
			response.Success(c, http.StatusOK, "Database connection is healthy", gin.H{
				"status": "ok",
			})
		})

		categoryRepo := categories.NewCategoryRepository(db)
		categoryService := categories.NewCategoryService(categoryRepo, cfg)
		categoryHandler := categories.NewCategoryHandler(categoryService)
		categoryHandler.RegisterRoutes(api.Group("/categories"))

		expenseRepo := expenses.NewExpenseRepository(db)
		expenseService := expenses.NewExpenseService(expenseRepo, cfg)
		expenseHandler := expenses.NewExpenseHandler(expenseService)
		expenseHandler.RegisterRoutes(api.Group("/expenses"))
	}

	return router
}
