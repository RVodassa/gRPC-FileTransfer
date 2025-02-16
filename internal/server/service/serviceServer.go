package service

import (
	"context"
	"github.com/RVodassa/FileTransfer/internal/server/config"
	"github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
)

const dataDir = "./data/server" // Папка для сохранения файлов

type FileServiceServer struct {
	file_transfer.UnimplementedFileTransferServer
	fileUploadSemaphore   *semaphore.Weighted
	fileDownloadSemaphore *semaphore.Weighted
	listFilesSemaphore    *semaphore.Weighted
}

func NewServiceServer(cfg *config.ServerConfig) *FileServiceServer {
	return &FileServiceServer{
		fileUploadSemaphore:   semaphore.NewWeighted(int64(cfg.Server.Limits.UploadRequests)),
		fileDownloadSemaphore: semaphore.NewWeighted(int64(cfg.Server.Limits.DownloadRequests)),
		listFilesSemaphore:    semaphore.NewWeighted(int64(cfg.Server.Limits.ListRequests)),
	}
}

func (s *FileServiceServer) UploadFile(stream file_transfer.FileTransfer_UploadFileServer) error {
	log.Println("UploadFile called")
	file := NewFile()
	var filename string
	var fileCreated bool

	defer func() {
		if fileCreated {
			if err := file.Close(); err != nil {
				log.Printf("failed to close file: %v", err)
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("File upload completed successfully.")
			break
		}
		if err != nil {
			log.Printf("failed to receive data: %v", err)
			return err
		}

		if filename == "" {
			if req.Filename == "" {
				log.Printf("filename is required")
				return status.Error(codes.InvalidArgument, "filename is required")
			}
			filename = filepath.Base(req.Filename)
			if err = file.SetFile(filename, dataDir); err != nil {
				log.Printf("failed to create file: %v", err)
				return err
			}
			fileCreated = true
			log.Printf("File created: %s", filename)
		}

		if len(req.Content) > 0 {
			if err = file.Write(req.Content); err != nil {
				log.Printf("failed to write data: %v", err)
				return err
			}
		} else {
			log.Println("Received empty content, skipping...")
		}
	}

	return stream.SendAndClose(&file_transfer.UploadFileResponse{Message: "File uploaded successfully!"})
}

func (s *FileServiceServer) ListFiles(ctx context.Context, req *file_transfer.Empty) (*file_transfer.ListFilesResponse, error) {
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	var fileInfos []*file_transfer.FileInfo
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dataDir, file.Name())
			fileStat, err := os.Stat(filePath)
			if err != nil {
				log.Printf("error: %v", err)
				return nil, err
			}

			fileInfos = append(fileInfos, &file_transfer.FileInfo{
				Name:             file.Name(),
				CreationTime:     fileStat.ModTime().Format("2006-01-02 15:04:05"),
				ModificationTime: fileStat.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return &file_transfer.ListFilesResponse{Files: fileInfos}, nil
}

func (s *FileServiceServer) GetFile(req *file_transfer.GetFileRequest, stream file_transfer.FileTransfer_GetFileServer) error {
	filePath := filepath.Join(dataDir, req.Filename)

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("error: %v", err)
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {

		}
	}(file)

	buf := make([]byte, 1024) // Размер буфера
	for {
		n, errRead := file.Read(buf)
		if errRead != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Отправляем часть файла
		if err = stream.Send(&file_transfer.GetFileResponse{Content: buf[:n]}); err != nil {
			return err
		}
	}

	return nil
}
