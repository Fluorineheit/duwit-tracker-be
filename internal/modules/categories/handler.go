package categories

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service *CategoryService
}

func NewCategoryHandler(service *CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

func (h *CategoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.FindAll)
	router.POST("", h.Create)
	router.GET("/:id", h.FindByID)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.SoftDelete)
}

func (h *CategoryHandler) FindAll(c *gin.Context) {
	result, err := h.service.FindAll(c.Request.Context(), ListCategoriesQuery{
		Type:   c.Query("type"),
		Limit:  parseIntQuery(c, "limit", 20),
		Cursor: c.Query("cursor"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Categories retrieved successfully", result)
}

func (h *CategoryHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	category, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		handleCategoryError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category retrieved successfully", category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleCategoryError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Category created successfully", category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleCategoryError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category updated successfully", category)
}

func (h *CategoryHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.SoftDelete(c.Request.Context(), id); err != nil {
		handleCategoryError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category deleted successfully", nil)
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func handleCategoryError(c *gin.Context, err error) {
	if errors.Is(err, ErrCategoryNotFound) {
		response.Error(c, http.StatusNotFound, "Category not found")
		return
	}

	if errors.Is(err, ErrCategoryAlreadyExists) {
		response.Error(c, http.StatusConflict, "Category already exists")
		return
	}

	response.Error(c, http.StatusBadRequest, err.Error())
}