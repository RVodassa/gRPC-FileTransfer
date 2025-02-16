package app

import (
	"github.com/RVodassa/FileTransfer/internal/client/service"
	pb "github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type App struct {
	address       string
	clientService *service.ClientService
	conn          *grpc.ClientConn
}

func New(address string) *App {
	return &App{address: address}
}

func (a *App) Initialize() error {
	var err error
	a.conn, err = grpc.NewClient(a.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client := pb.NewFileTransferClient(a.conn)
	a.clientService = service.New(client)
	return nil
}

func (a *App) Close() {
	if a.conn != nil {
		err := a.conn.Close()
		if err != nil {
			return
		}
	}
}

func (a *App) AddCommands(rootCmd *cobra.Command) {
	var uploadCmd = &cobra.Command{
		Use:   "upload [filename]",
		Short: "Upload a file to the server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			if err := a.clientService.UploadFile(filename); err != nil {
				log.Fatalf("failed to upload file: %v", err)
			}
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List files on the server",
		Run: func(cmd *cobra.Command, args []string) {
			a.clientService.ListFiles()
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [filename]",
		Short: "Download a file from the server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			a.clientService.GetFile(filename)
		},
	}

	rootCmd.AddCommand(uploadCmd, listCmd, getCmd)
}
