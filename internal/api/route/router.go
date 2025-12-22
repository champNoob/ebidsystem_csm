package route

import (
	"ebidsystem_csm/internal/api/handler"
	"ebidsystem_csm/internal/middleware/auth"
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

	// === 注册路由 ===
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.POST("/users", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)

	// === 需要登录 ===
	api := r.Group("/api")
	api.Use(auth.JWTAuthMiddleware())

	api.GET("/me", userHandler.GetMe)

	// === 管理员接口 ===
	admin := r.Group("/api/admin")
	admin.Use(
		auth.JWTAuthMiddleware(),
		auth.RequireRole("admin"),
	)

	admin.GET("/users/:id", userHandler.GetUser)
	admin.POST("/users", userHandler.CreateUser)

	return r
}
