package route

import (
	"ebidsystem_csm/internal/api/handler"
	"ebidsystem_csm/internal/middleware/auth"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// === CORS 配置 ===
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 前端 Vite 地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	// 用户侧：
	api.GET("/me", userHandler.GetMe)
	// 订单侧：
	api.POST(
		"/orders",
		auth.RequireRole("client", "seller", "trader"),
		orderHandler.CreateOrder,
	)
	api.GET(
		"/orders",
		auth.RequireRole("client", "seller", "trader", "admin"),
		orderHandler.ListOrders,
	)
	api.POST(
		"/orders/:id/cancel",
		auth.RequireRole("client", "seller", "trader", "admin"),
		orderHandler.CancelOrder,
	)

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
