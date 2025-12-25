package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/service"
	"net/http"
	"strconv"

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
		req.OrderType,
		req.OrderSide,
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

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid order id"})
		return
	}

	userID := c.GetInt64("userID")
	role := c.GetString("role")

	err = h.service.CancelOrder(
		c.Request.Context(),
		orderID,
		userID,
		role,
	)

	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			c.JSON(404, gin.H{"error": err.Error()})
		case service.ErrOrderNotCancelable, service.ErrPermissionDenied:
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(200, gin.H{"message": "order canceled"})
}
