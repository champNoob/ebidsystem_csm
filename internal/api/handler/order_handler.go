package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/model"
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
		c.JSON(http.StatusBadRequest, gin.H{ //400
			"error": service.ErrInvalidInput.Error(), //服务层错误封装，统一映射
		})
		return
	}

	userID := c.GetInt64("userID")
	roleStr := c.GetString("role")
	role, err := model.ParseUserRole(roleStr)
	if err != nil {
		c.JSON(403, gin.H{"error": "invalid user role"})
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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "order created"})
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	// 1. 从 JWT 中取 userID：
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDAny.(int64)
	role := c.GetString("role")
	// 2. 解析 query 参数：
	var req request.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid query"})
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
	// 调用服务层对应的函数：
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
		case service.ErrOrderNotCancellable, service.ErrPermissionDenied:
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(200, gin.H{"message": "order canceled"})
}
