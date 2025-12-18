package main

import (
	"log"

	"ebidsystem_csm/internal/api/route"
	"ebidsystem_csm/internal/config"
	"ebidsystem_csm/internal/pkg/database"
)

func main() {
	//1.加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	//2.初始化数据库
	if err := database.InitMySQL(cfg.MySQL); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	if err := database.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("init redis failed: %v", err)
	}

	//3.启动 HTTP 服务
	r := route.SetupRouter()
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
