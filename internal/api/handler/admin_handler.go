package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/apperror"
	"ebidsystem_csm/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{service: s}
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	res, err := h.service.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUserRoleStats(c *gin.Context) {
	res, err := h.service.GetUserRoleStats(c.Request.Context())
	if err != nil {
		respondError(c, apperror.ErrInternal)
		return
	}
	c.JSON(200, res)
}

func (h *AdminHandler) GetUserRanking(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 10)
	res, err := h.service.GetUserRanking(c.Request.Context(), limit)
	if err != nil {
		respondError(c, apperror.ErrInternal)
		return
	}
	c.JSON(200, res)
}

func (h *AdminHandler) ListUserOrders(c *gin.Context) {
	// userID := c.Param("id")
	// res, err := h.service.ListUserOrders(c.Request.Context(), userID)
	// if err != nil {
	// 	respondError(c, apperror.ErrInternal)
	// 	return
	// }
	// c.JSON(200, res)
}

func (h *AdminHandler) GetSymbolStats(c *gin.Context) {
	stats, err := h.service.GetSymbolStats(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}

func (h *AdminHandler) GetOrderStatusStats(c *gin.Context) {
	res, err := h.service.GetOrderStatusStats(c.Request.Context())
	if err != nil {
		respondError(c, apperror.ErrInternal)
		return
	}
	c.JSON(200, res)
}

func (h *AdminHandler) GetRecentTrades(c *gin.Context) {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	res, err := h.service.GetRecentTrades(c.Request.Context(), limit)
	if err != nil {
		respondError(c, apperror.ErrInternal)
		return
	}
	c.JSON(200, res)
}

func (h *AdminHandler) GetTradeTimeline(c *gin.Context) {
	res, err := h.service.GetTradeTimeline(c.Request.Context())
	if err != nil {
		respondError(c, apperror.ErrInternal)
		return
	}
	c.JSON(200, res)
}

func (h *AdminHandler) AdminListOrders(c *gin.Context) {
	var req request.AdminListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, apperror.ErrInvalidOrderQuery)
		return
	}

	res, err := h.service.GetAdminOrders(
		c.Request.Context(),
		service.AdminListOrdersInput{
			UserID:   req.UserID,
			Symbol:   req.Symbol,
			Status:   req.Status,
			Side:     req.Side,
			Type:     req.Type,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(200, res)
}
