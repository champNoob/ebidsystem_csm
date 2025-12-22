package route

import (
	"ebidsystem_csm/internal/api/handler"
	"ebidsystem_csm/internal/middleware"
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

	// r.GET("/users/:id", userHandler.GetUser)
	r.POST("/users", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)

	// === 需要登录 ===
	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		auth.GET("/users/:id", userHandler.GetUser)
	}

	// === 需要 admin 权限 ===
	admin := r.Group("/admin")
	admin.Use(
		middleware.JWTAuthMiddleware(),
		middleware.RoleMiddleware("admin"),
	)
	{
		admin.GET("/users/:id", userHandler.GetUser)
	}

	return r
}
