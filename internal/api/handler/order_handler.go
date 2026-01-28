package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/service"
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
		respondError(c, service.ErrInvalidInput)
		return
	}

	userID := c.GetInt64("userID")
	roleStr := c.GetString("role")
	role, err := model.ParseUserRole(roleStr)
	if err != nil {
		respondError(c, service.ErrInvalidUserRole)
		return
	}

	if err := h.service.CreateOrder(
		c.Request.Context(),
		userID,
		role,
		req.Symbol,
		req.OrderType,
		req.OrderSide,
		req.Price,
		req.Quantity,
	); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(201, gin.H{"message": "order created"})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	// 1. 从 JWT 中取 userID：
	userIDAny, exists := c.Get("userID")
	if !exists {
		respondError(c, service.ErrUserUnauthorized)
		return
	}
	userID := userIDAny.(int64)
	role := c.GetString("role")
	// 2. 解析 query 参数：
	var req request.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, service.ErrInvalidOrderQuery)
		return
	}
	// 3. 调用 service：
	orders, err := h.service.ListOrders(
		c.Request.Context(),
		userID,
		role,
		req.Status,
	)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(200, orders)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, service.ErrInvalidOrderID)
		return
	}

	userID := c.GetInt64("userID")
	role := c.GetString("role")
	// 调用服务层对应的函数：
	err = h.service.CancelOrder(
		c.Request.Context(),
		orderID,
		userID,
		role,
	)

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "order canceled"})
}
