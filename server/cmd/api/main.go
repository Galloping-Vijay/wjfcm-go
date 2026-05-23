package main

import (
	"flag"
	"log"

	"wjfcm-go/internal/config"
	"wjfcm-go/internal/database"
	"wjfcm-go/internal/router"
)

func main() {
	envFile := flag.String("f", "", "env file path, default .env")
	flag.Parse()

	cfg := config.Load()
	if *envFile != "" {
		cfg = config.Load(*envFile)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Printf("connect database: %v", err)
		log.Printf("database is unavailable, only installer routes may work")
	}

	r := router.New(cfg, db)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
