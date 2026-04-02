package main

import (
	"log"

	"seu-oj-backend/internal/bootstrap"
	"seu-oj-backend/internal/config"
	"seu-oj-backend/internal/database"
	"seu-oj-backend/internal/router"
)

func main() {
	logFile, err := bootstrap.InitLogging()
	if err != nil {
		log.Fatalf("init logging failed: %v", err)
	}
	defer logFile.Close()

	cfg := config.Load()

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis failed: %v", err)
	}

	engine := router.New(db, redisClient, cfg)
	if err := engine.Run(cfg.Server.Address()); err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}
