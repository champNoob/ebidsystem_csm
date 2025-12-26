package main

import (
	"log"

	"ebidsystem_csm/internal/api/handler"
	"ebidsystem_csm/internal/api/route"
	"ebidsystem_csm/internal/config"
	"ebidsystem_csm/internal/matching"
	db "ebidsystem_csm/internal/pkg/database"
	"ebidsystem_csm/internal/repository/mysql"
	"ebidsystem_csm/internal/service"
)

func main() {
	// 1. 加载配置：
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 2. 初始化数据库：
	if err := db.InitMySQL(cfg.MySQL); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}
	if err := db.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("init redis failed: %v", err)
	}

	// 3. 初始化撮合引擎（matching）
	engine := matching.NewEngine()
	engine.Start()

	// 4. 初始化仓储层（repository）
	userRepo := mysql.NewUserRepo(db.MySQL)
	orderRepo := mysql.NewOrderRepo(db.MySQL)

	// 5. 初始化服务层（service）
	orderService := service.NewOrderService(orderRepo, engine)
	userService := service.NewUserService(userRepo)

	// 6. 初始化处理器（Handler）
	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)

	// 7. Router（Http服务，只接收 handler）
	r := route.SetupRouter(
		userHandler,
		orderHandler,
	)
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
