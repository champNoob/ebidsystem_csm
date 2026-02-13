package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	orderService.StartMatchEventListener() //启动撮合事件监听器

	// 6. 初始化处理器（Handler）
	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)

	// 7. Router（Http服务，只接收 handler）
	r := route.SetupRouter(
		userHandler,
		orderHandler,
	)
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}

	engine.Stop()
	log.Println("Server exited properly")

}
