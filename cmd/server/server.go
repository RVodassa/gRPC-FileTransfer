package main

import (
	"github.com/RVodassa/FileTransfer/internal/server/app"
	"github.com/RVodassa/FileTransfer/internal/server/config"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	cfgPath := os.Getenv("SERVER_CONFIG_PATH")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	app.Run(cfg)
}
