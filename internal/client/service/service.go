package service

import (
	"context"
	"fmt"
	pb "github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const clientDir = "./data/client"

type ClientService struct {
	client pb.FileTransferClient
}

func New(client pb.FileTransferClient) *ClientService {
	return &ClientService{
		client: client,
	}
}

func (c *ClientService) UploadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {

		}
	}(file)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		log.Fatalf("could not upload file: %v", err)
	}

	if err = stream.Send(&pb.UploadFileRequest{Filename: filepath.Base(filename)}); err != nil {
		log.Fatalf("failed to send filename: %v", err)
	}

	buf := make([]byte, 32*1024)
	var n int
	for {
		n, err = file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("failed to read file: %v", err)
		}

		if n > 0 {
			if err = stream.Send(&pb.UploadFileRequest{Content: buf[:n]}); err != nil {
				log.Fatalf("failed to send file: %v", err)
			}
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive response: %v", err)
	}

	fmt.Printf("Server response: %s\n", resp.Message)
	return nil
}

func (c *ClientService) ListFiles() {
	resp, err := c.client.ListFiles(context.Background(), &pb.Empty{})
	if err != nil {
		log.Printf("could not list files: %v", err)
	}

	fmt.Println("Files in the upload directory:")
	for _, fileInfo := range resp.Files {
		fmt.Printf("Name: %s, Creation Time: %s, Modification Time: %s\n",
			fileInfo.Name, fileInfo.CreationTime, fileInfo.ModificationTime)
	}
}

func (c *ClientService) GetFile(filename string) {
	req := &pb.GetFileRequest{Filename: filename}
	stream, err := c.client.GetFile(context.Background(), req)
	if err != nil {
		log.Printf("could not get file: %v", err)
	}

	filename = filepath.Base("downloaded_" + filename)
	filePath := filepath.Join(clientDir, filename)

	err = os.MkdirAll(clientDir, os.ModePerm)
	if err != nil {
		log.Printf("failed to create directory: %v", err)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("could not create file: %v", err)
	}
	defer func(outFile *os.File) {
		err = outFile.Close()
		if err != nil {

		}
	}(outFile)

	for {
		resp, errResp := stream.Recv()
		if errResp != nil {
			if err == io.EOF {
				break
			}
			log.Printf("failed to receive file content: %v", err)
		}

		// Записываем содержимое в файл
		if _, err = outFile.Write(resp.Content); err != nil {
			log.Printf("failed to write to file: %v", err)
		}
	}

	fmt.Printf("File %s downloaded successfully.\n", filename)
}
