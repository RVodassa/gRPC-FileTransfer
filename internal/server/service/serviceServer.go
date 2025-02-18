package service

import (
	"context"
	"errors"
	"github.com/RVodassa/FileTransfer/pkg/file"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/RVodassa/FileTransfer/internal/server/config"
	"github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultBufSize = 1024 * 1024 // 1 MB буфер для чтения/записи файлов
)

var ErrNotFound = errors.New("file not found")
var ErrLimitRequest = errors.New("too many requests")
var ErrFilesNotFound = errors.New("file not found")

type FileServiceServer struct {
	file_transfer.UnimplementedFileTransferServer
	dataDir               string
	fileUploadSemaphore   *semaphore.Weighted
	fileDownloadSemaphore *semaphore.Weighted
	listFilesSemaphore    *semaphore.Weighted
	mu                    sync.Mutex
}

// NewServiceServer возвращает новый инстанс сервиса
func NewServiceServer(cfg *config.ServerConfig) *FileServiceServer {
	return &FileServiceServer{
		dataDir:               cfg.ServerDataDir,
		fileUploadSemaphore:   semaphore.NewWeighted(int64(cfg.Server.Limits.UploadRequests)),
		fileDownloadSemaphore: semaphore.NewWeighted(int64(cfg.Server.Limits.DownloadRequests)),
		listFilesSemaphore:    semaphore.NewWeighted(int64(cfg.Server.Limits.ListRequests)),
	}
}

// UploadFile загружает файл клиента на сервер
func (s *FileServiceServer) UploadFile(stream file_transfer.FileTransfer_UploadFileServer) error {
	const op = "server.service.UploadFile"

	// Ограничивает кол-во одновременных запросов
	if err := s.fileUploadSemaphore.Acquire(stream.Context(), 1); err != nil {
		return status.Error(codes.ResourceExhausted, ErrLimitRequest.Error())
	}
	defer s.fileUploadSemaphore.Release(1)

	// обработка данных
	var filename string
	var f *file.File

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("%s: filename:%s. Upload completed", op, filename)
				return stream.SendAndClose(&file_transfer.UploadFileResponse{Message: "File uploaded successfully!"})
			}
			log.Printf("%s: filename:%s. failed to receive data: %v", op, filename, err)
			return status.Errorf(codes.Internal, "filename:%s. failed to receive data: %v", filename, err)
		}

		//  создает файл в первом цикле for
		if filename == "" {
			s.mu.Lock()
			defer s.mu.Unlock()

			f = file.NewFile()
			filename = filepath.Base(req.Filename)

			err = f.SetFile(filename, s.dataDir)
			if err != nil {
				log.Printf("%s: failed to set file: %v", op, err)
				return status.Errorf(codes.Internal, "failed to set file: %v", err)
			}
		}

		// записывает данные в файл
		if len(req.Content) > 0 {
			log.Printf("%s: received %d bytes for file: %s", op, len(req.Content), filename)
			if err = f.Write(req.Content); err != nil {
				log.Printf("%s: failed to write data: %v", op, err)
				return status.Errorf(codes.Internal, "failed to write data: %v", err)
			}
		} else {
			log.Printf("%s: received empty content for file: %s", op, filename)
		}
	}
}

// ListFiles возвращает клиенту информацию о файлах
func (s *FileServiceServer) ListFiles(ctx context.Context, req *file_transfer.Empty) (*file_transfer.ListFilesResponse, error) {
	const op = "server.service.ListFiles"

	// Ограничивает кол-во одновременных запросов
	if err := s.listFilesSemaphore.Acquire(ctx, 1); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, ErrLimitRequest.Error())
	}
	defer s.listFilesSemaphore.Release(1)

	// читает директорию с файлами
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, ErrFilesNotFound.Error())
		}
		log.Printf("%s: failed to read directory: %v", op, err)
		return nil, status.Errorf(codes.Internal, "failed to read directory: %v", err)
	}

	// Читает информацию о файлах
	var fileInfos []*file_transfer.FileInfo
	var fileStat os.FileInfo
	for _, f := range files {
		if !f.IsDir() {
			filePath := filepath.Join(s.dataDir, f.Name())
			fileStat, err = os.Stat(filePath)
			if err != nil {
				log.Printf("%s: failed to get file info: %v", op, err)
				continue // Пропускаем файл, если не удалось получить информацию
			}

			fileInfos = append(fileInfos, &file_transfer.FileInfo{
				Name:             f.Name(),
				CreationTime:     fileStat.ModTime().Format("2006-01-02 15:04:05"),
				ModificationTime: fileStat.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return &file_transfer.ListFilesResponse{Files: fileInfos}, nil
}

// GetFile отправляет файл клиенту
func (s *FileServiceServer) GetFile(req *file_transfer.GetFileRequest, stream file_transfer.FileTransfer_GetFileServer) error {
	const op = "server.service.GetFile"

	// Ограничиваем кол-во одновременных скачиваний
	if err := s.fileDownloadSemaphore.Acquire(stream.Context(), 1); err != nil {
		return status.Errorf(codes.ResourceExhausted, ErrLimitRequest.Error())
	}
	defer s.fileDownloadSemaphore.Release(1)

	filePath := filepath.Join(s.dataDir, req.Filename)

	// Проверяет, существует ли директория с файлами
	if _, err := os.Stat(filepath.Dir(filePath)); err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s: directory does not exist: %v", op, err)
			return status.Error(codes.NotFound, ErrFilesNotFound.Error())
		}
		log.Printf("%s: failed to stat directory: %v", op, err)
		return status.Errorf(codes.Internal, "failed to stat directory: %v", err)
	}

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%s: file not found: %v", op, err)
		return status.Error(codes.NotFound, ErrNotFound.Error())
	} else if err != nil {
		log.Printf("%s: failed to stat file: %v", op, err)
		return status.Errorf(codes.Internal, "failed to stat file: %v", err)
	}

	// Отправляет файл клиенту частями
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("%s: failed to open file: %v", op, err)
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("%s: failed to close file: %v", op, closeErr)
		}
	}()

	buf := make([]byte, defaultBufSize)
	var n int
	for {
		n, err = f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("%s: failed to read file: %v", op, err)
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}
		if err = stream.Send(&file_transfer.GetFileResponse{Content: buf[:n]}); err != nil {
			log.Printf("%s: failed to send file chunk: %v", op, err)
			return status.Errorf(codes.Internal, "failed to send file chunk: %v", err)
		}
	}
	return nil
}
