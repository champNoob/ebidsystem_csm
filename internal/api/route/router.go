package route

import (
	"ebidsystem_csm/internal/api/handler"
	"ebidsystem_csm/internal/pkg/database"
	"ebidsystem_csm/internal/repository/mysql"
	"ebidsystem_csm/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// === 装配依赖 ===
	userRepo := mysql.NewUserRepo(database.MySQL)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// === 路由注册 ===
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/users/:id", userHandler.GetUser)
	r.POST("/users", userHandler.CreateUser)

	return r
}
