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
	adminHandler *handler.AdminHandler,
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
		auth.RequireRole("client", "seller", "trader", "admin"), //# -admin
		orderHandler.CancelOrder,
	)

	// === 管理员接口 ===
	admin := r.Group("/api/admin")
	admin.Use(
		auth.JWTAuthMiddleware(),
		auth.RequireRole("admin"),
	)

	// 用户管理：
	adminUsers := admin.Group("/users")
	{
		adminUsers.GET("/:id", userHandler.GetUser)
		adminUsers.POST("", userHandler.CreateUser)
		// 后续实现：
		// adminUsers.GET("", userHandler.ListUsers)
		// adminUsers.PUT("/:id/role", userHandler.UpdateUserRole)
		// adminUsers.POST("/:id/disable", userHandler.DisableUser)
		// adminUsers.GET("/:id/orders", orderHandler.AdminListUserOrders)
	}
	// 订单管理：
	/*
		adminOrders := admin.Group("/orders")
		{
			// 后续实现：
			// adminOrders.GET("", orderHandler.AdminListOrders)
		}
	*/
	// 交易管理：
	adminTrades := admin.Group("/trades")
	{
		// 后续实现：
		// adminTrades.GET("", adminHandler.GetRecentTrades)
		adminTrades.GET("/recent", adminHandler.GetRecentTrades)
	}
	// 看板：
	dashboard := admin.Group("/dashboard")
	{
		dashboard.GET("", adminHandler.GetDashboard)
		dashboard.GET("/symbols", adminHandler.GetSymbolStats)
		dashboard.GET("/order-status", adminHandler.GetOrderStatusStats)
		dashboard.GET("/user-roles", adminHandler.GetUserRoleStats)
		dashboard.GET("/user-ranking", adminHandler.GetUserRanking)
		dashboard.GET("/trades/timeline", adminHandler.GetTradeTimeline)
	}

	return r
}
