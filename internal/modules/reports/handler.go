package reports

import (
	"net/http"

	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service *ReportService
}

func NewReportHandler(service *ReportService) *ReportHandler {
	return &ReportHandler{
		service: service,
	}
}

func (h *ReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/daily", h.Daily)
	router.GET("/monthly", h.Monthly)
	router.GET("/category", h.Category)
}

func (h *ReportHandler) Daily(c *gin.Context) {
	items, err := h.service.Daily(c.Request.Context(), DailyReportQuery{
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Daily report retrieved successfully", items)
}

func (h *ReportHandler) Monthly(c *gin.Context) {
	items, err := h.service.Monthly(c.Request.Context(), MonthlyReportQuery{
		Year: c.Query("year"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Monthly report retrieved successfully", items)
}

func (h *ReportHandler) Category(c *gin.Context) {
	items, err := h.service.Category(c.Request.Context(), CategoryReportQuery{
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Category report retrieved successfully", items)
}
