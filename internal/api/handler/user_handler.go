package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/api/dto/response"
	"ebidsystem_csm/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

// GetMe 返回当前登录用户的信息
func (h *UserHandler) GetMe(c *gin.Context) {
	// 1. 从 JWT Middleware 写入的 context 中取 userID
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(500, gin.H{"error": "invalid user id type"})
		return
	}

	// 2. 调用 service 层
	user, err := h.service.GetUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 3. 返回用户信息
	c.JSON(200, response.UserMeResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ctx := c.Request.Context()
	if err := h.service.CreateUser(
		c.Request.Context(),
		service.CreateUserInput{
			Username: req.Username,
			Password: req.Password,
			Role:     req.Role,
		},
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.service.Login(
		c.Request.Context(),
		service.LoginInput{
			Username: req.Username,
			Password: req.Password,
		},
	)

	if err != nil {
		switch err {
		case service.ErrUserNotFound, service.ErrInvalidPassword:
			c.JSON(401, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(500, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(200, gin.H{"token": token})
}
