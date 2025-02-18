package app

import (
	"context"
	"github.com/RVodassa/FileTransfer/internal/client/config"
	"github.com/RVodassa/FileTransfer/internal/client/service"
	pb "github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type App struct {
	cfg           *config.Config
	clientService *service.ClientService
	conn          *grpc.ClientConn
}

func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

// Initialize установка значений для полей conn и clientService в App
func (a *App) Initialize() error {
	const op = "app.Initialize"
	var err error
	a.conn, err = grpc.NewClient(a.cfg.Server.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("%s: error creating grpc client. Error: %v", op, err)
		return err
	}
	client := pb.NewFileTransferClient(a.conn)
	a.clientService = service.New(client, a.cfg.ClientDataDir)
	return nil
}

// Close закрывает conn в App
func (a *App) Close() {
	const op = "app.Close"
	if a.conn != nil {
		err := a.conn.Close()
		if err != nil {
			log.Printf("%s. fail close conn. Error: %v", op, err)
			return
		}
	}
}

// AddCommands настройка команд cobra CLI
func (a *App) AddCommands(rootCmd *cobra.Command) {

	var uploadCmd = &cobra.Command{
		Use:   "upload [filename]",
		Short: "Upload a file to the server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			_ = a.clientService.UploadFile(context.Background(), filename)
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List files on the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = a.clientService.ListFiles(context.Background())
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [filename]",
		Short: "Download a file from the server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			_ = a.clientService.GetFile(context.Background(), filename)
		},
	}

	rootCmd.AddCommand(uploadCmd, listCmd, getCmd)
}
