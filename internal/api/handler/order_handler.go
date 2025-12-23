package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req request.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	if err := h.service.CreateOrder(
		c.Request.Context(),
		userID,
		req.Symbol,
		req.Side,
		req.Price,
		req.Quantity,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "order created"})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.GetInt64("userID")
	role := c.GetString("role")

	orders, err := h.service.ListOrders(
		c.Request.Context(),
		userID,
		role,
	)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, orders)
}
