package app

import (
	"github.com/RVodassa/FileTransfer/internal/server/config"
	"github.com/RVodassa/FileTransfer/internal/server/service"
	pb "github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Run(cfg *config.ServerConfig) {

	lis, err := net.Listen("tcp", cfg.Server.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	serviceServer := service.NewServiceServer(cfg)
	pb.RegisterFileTransferServer(s, serviceServer)

	log.Printf("Server is running on port %s", cfg.Server.Address)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
