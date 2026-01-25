package handler

import (
	"ebidsystem_csm/internal/api/dto/request"
	"ebidsystem_csm/internal/api/dto/response"
	"ebidsystem_csm/internal/service"
	"log"
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
		respondError(c, service.ErrInvalidUserID)
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(200, user)
}

// GetMe 返回当前登录用户的信息
func (h *UserHandler) GetMe(c *gin.Context) {
	// 1. 从 JWT Middleware 写入的 context 中取 userID
	userIDAny, exists := c.Get("userID")
	if !exists {
		respondError(c, service.ErrUserUnauthorized)
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		log.Printf("Invalid userID type: %T", userIDAny)
		respondError(c, service.ErrInternal)
		return
	}

	// 2. 调用 service 层
	user, err := h.service.GetUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		respondError(c, err)
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
		respondError(c, service.ErrInvalidInput)
		return
	}

	if err := h.service.CreateUser(
		c.Request.Context(),
		service.CreateUserInput{
			Username: req.Username,
			Password: req.Password,
			Role:     req.Role,
		},
	); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, service.ErrInvalidInput)
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
		respondError(c, err)
		return
	}

	c.JSON(200, gin.H{"token": token})
}
