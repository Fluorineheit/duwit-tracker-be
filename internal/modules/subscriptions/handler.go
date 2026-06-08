package subscriptions

import (
	"errors"
	"net/http"

	"github.com/Fluorineheit/duwit-tracker-be/internal/response"
	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	service *SubscriptionService
}

func NewSubscriptionHandler(service *SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.FindAll)
	router.POST("", h.Create)
	router.PUT("/:id", h.Update)
	router.DELETE("/:id", h.Deactivate)
}

func (h *SubscriptionHandler) FindAll(c *gin.Context) {
	result, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Subscriptions retrieved successfully", result)
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	subscription, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Subscription created successfully", subscription)
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	subscription, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSubscriptionError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription updated successfully", subscription)
}

func (h *SubscriptionHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Deactivate(c.Request.Context(), id); err != nil {
		handleSubscriptionError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Subscription deactivated successfully", nil)
}

func handleSubscriptionError(c *gin.Context, err error) {
	if errors.Is(err, ErrSubscriptionNotFound) {
		response.Error(c, http.StatusNotFound, "Subscription not found")
		return
	}

	response.Error(c, http.StatusBadRequest, err.Error())
}
