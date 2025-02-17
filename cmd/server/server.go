package main

import (
	"github.com/RVodassa/FileTransfer/internal/server/app"
	"github.com/RVodassa/FileTransfer/internal/server/config"
	"log"
)

const ServerConfigPath = "./configs/server_config.yaml"

func main() {

	cfg, err := config.LoadConfig(ServerConfigPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	app.Run(cfg)
}
