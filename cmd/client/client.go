package main

import (
	"github.com/RVodassa/FileTransfer/internal/client/app"
	"github.com/RVodassa/FileTransfer/internal/client/config"
	"github.com/spf13/cobra"
	"log"
)

// ClientConfigPath путь до файла конфиг.
const ClientConfigPath = "./configs/client_config.yaml"

func main() {

	cfg, err := config.LoadConfig(ClientConfigPath)
	if err != nil {
		log.Printf("Error loading client config: %v", err)
		return
	}

	// cobra CLI и инициализация App

	newApp := app.New(cfg)

	rootCmd := &cobra.Command{
		Use:   "file_transfer_client",
		Short: "File Transfer CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err = newApp.Initialize(); err != nil {
				log.Fatalf("did not connect: %v", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			newApp.Close()
		},
	}

	newApp.AddCommands(rootCmd) // App

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("command execution failed: %v", err)
	}
}
