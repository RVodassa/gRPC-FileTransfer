package main

import (
	"github.com/RVodassa/FileTransfer/internal/client/app"
	"github.com/RVodassa/FileTransfer/internal/client/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	cfgPath := os.Getenv("CLIENT_CONFIG_PATH")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Printf("Error loading client config: %v", err)
		return
	}

	newApp := app.New(cfg.Server.Address)

	rootCmd := &cobra.Command{
		Use:   "file_transfer_client",
		Short: "File Transfer CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Инициализация приложения
			if err = newApp.Initialize(); err != nil {
				log.Fatalf("did not connect: %v", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			newApp.Close()
		},
	}

	newApp.AddCommands(rootCmd)

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("command execution failed: %v", err)
	}
}
