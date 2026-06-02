package expenses

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

type ExpenseHandler struct {
	service *ExpenseService
}

func NewExpenseHandler(service *ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		service: service,
	}
}

func (h *ExpenseHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.FindAll)
	router.POST("", h.Create)
	router.GET("/:id", h.FindByID)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.SoftDelete)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	var req CreateExpenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	expense, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Expense created successfully", expense)
}

func (h *ExpenseHandler) FindAll(c *gin.Context) {
	limit := parseIntQuery(c, "limit", 20)
	offset := parseIntQuery(c, "offset", 0)

	result, err := h.service.FindAll(c.Request.Context(), ListExpensesQuery{
		Limit:      limit,
		Offset:     offset,
		CategoryID: c.Query("category_id"),
		From:       c.Query("from"),
		To:         c.Query("to"),
		Search:     c.Query("search"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Expenses retrieved successfully", result)
}

func (h *ExpenseHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	expense, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		handleExpenseError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Expense retrieved successfully", expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	expense, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleExpenseError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Expense updated successfully", expense)
}

func (h *ExpenseHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.SoftDelete(c.Request.Context(), id); err != nil {
		handleExpenseError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Expense deleted successfully", nil)
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

func handleExpenseError(c *gin.Context, err error) {
	if errors.Is(err, ErrExpenseNotFound) {
		response.Error(c, http.StatusNotFound, "Expense not found")
		return
	}

	response.Error(c, http.StatusBadRequest, err.Error())
}
