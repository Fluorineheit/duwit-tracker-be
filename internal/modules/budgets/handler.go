package budgets

import (
	"errors"
	"net/http"

	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

type BudgetHandler struct {
	service *BudgetService
}

func NewBudgetHandler(service *BudgetService) *BudgetHandler {
	return &BudgetHandler{
		service: service,
	}
}

func (h *BudgetHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.FindAll)
	router.POST("", h.Create)
	router.DELETE("/:id", h.SoftDelete)
}

func (h *BudgetHandler) FindAll(c *gin.Context) {
	result, err := h.service.FindAll(c.Request.Context(), ListBudgetsQuery{
		Month: c.Query("month"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Budgets retrieved successfully", result)
}

func (h *BudgetHandler) Create(c *gin.Context) {
	var req CreateBudgetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	budget, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleBudgetError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Budget created successfully", budget)
}

func (h *BudgetHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.SoftDelete(c.Request.Context(), id); err != nil {
		handleBudgetError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Budget deleted successfully", nil)
}

func handleBudgetError(c *gin.Context, err error) {
	if errors.Is(err, ErrBudgetNotFound) {
		response.Error(c, http.StatusNotFound, "Budget not found")
		return
	}

	if errors.Is(err, ErrBudgetAlreadyExists) {
		response.Error(c, http.StatusConflict, "Budget already exists")
		return
	}

	response.Error(c, http.StatusBadRequest, err.Error())
}
