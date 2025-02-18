package service

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/RVodassa/FileTransfer/pkg/protos/gen/file_transfer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
)

type ClientService struct {
	client  pb.FileTransferClient
	dataDir string
}

func New(client pb.FileTransferClient, dataDir string) *ClientService {
	return &ClientService{
		client:  client,
		dataDir: dataDir,
	}
}

var ErrNotFound = errors.New("file not found")
var ErrInternalServer = errors.New("internal server error")

const defaultBufSize = 1024 * 1024 // 1 MB буфер для чтения/записи файлов

// UploadFile загружает файл на сервер
func (c *ClientService) UploadFile(ctx context.Context, filePath string) error {
	const op = "client.service.UploadFile"

	if filePath == "" {
		log.Printf("%s: file path is empty", op)
		return fmt.Errorf("%s: filename is required", op)
	}
	// Поиск файла
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s: filePath:%s. Err: %v", op, filePath, ErrNotFound)
			return fmt.Errorf("%s: filePath:%s. Err: %w", op, filePath, ErrNotFound)
		}
		log.Printf("%s: filePath:%s. Err: %v", op, filePath, err)
		return fmt.Errorf("%s: filePath:%s. Err: %w", op, filePath, err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("%s: filePath:%s. Err: %v", op, filePath, err)
		}
	}()

	// Поток для загрузки файла
	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		return c.handleGRPCError(op, err)
	}

	// Отправляет имя файла
	filename := filepath.Base(filePath)
	if err = stream.Send(&pb.UploadFileRequest{Filename: filepath.Base(filename)}); err != nil {
		log.Printf("%s: filename:%s. Err: %v", op, filename, err)
		return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
	}

	// Передача данных
	var n int
	buf := make([]byte, defaultBufSize)
	for {
		n, err = file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("%s: filename:%s. Err: %v", op, filename, err)
			return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
		}

		if n > 0 {
			if err = stream.Send(&pb.UploadFileRequest{Content: buf[:n]}); err != nil {
				log.Printf("%s: filename:%s. Err: %v", op, filename, err)
				return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
			}
			log.Printf("%s: filename:%s. Sent %d bytes", op, filename, n) // Лог отправленных байт
		}
	}

	// Завершение потока и ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("%s: filename:%s. Err: %v", op, filename, err)
		return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
	}

	log.Printf("%s: filename:%s. %s", op, filename, resp.Message)
	return nil
}

// ListFiles вернет список доступных на сервере файлов
func (c *ClientService) ListFiles(ctx context.Context) error {
	const op = "client.service.ListFiles"

	resp, err := c.client.ListFiles(ctx, &pb.Empty{})
	if err != nil {
		return c.handleGRPCError(op, err)
	}

	fmt.Println("Files in the upload directory:")
	for _, fileInfo := range resp.Files {
		fmt.Printf("Name: %s, Creation Time: %s, Modification Time: %s\n",
			fileInfo.Name, fileInfo.CreationTime, fileInfo.ModificationTime)
	}
	return nil
}

// GetFile скачивает файл с сервера
func (c *ClientService) GetFile(ctx context.Context, filename string) error {
	const op = "client.service.GetFile"

	if filename == "" {
		log.Printf("%s: filename is required", op)
		return fmt.Errorf("filename is required")
	}

	req := &pb.GetFileRequest{Filename: filename}
	stream, err := c.client.GetFile(ctx, req)
	if err != nil {
		return c.handleGRPCError(op, err)
	}

	// Директория для хранения файлов клиента
	if err = os.MkdirAll(c.dataDir, os.ModePerm); err != nil {
		log.Printf("%s: filename:%s. Err: %v", op, filename, err)
		return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
	}

	// Создаем временный файл
	tmpFilename := "downloaded_" + filename + ".tmp"
	tmpFilePath := filepath.Join(c.dataDir, tmpFilename)

	f, err := os.Create(tmpFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("%s: filepath:%s. Err: %v", op, fmt.Sprintf(c.dataDir+filename), ErrNotFound)
			return fmt.Errorf("%s: filepath:%s. Err: %w", op, fmt.Sprintf(c.dataDir+filename), ErrNotFound)
		}
		log.Printf("%s: filename:%s. Err: %v", op, filename, err)
		return fmt.Errorf("%s: filename:%s. Err: %w", op, filename, err)
	}

	// Удаляем временный файл в случае ошибки
	var success bool
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("%s: filename:%s. Err: %v", op, filename, err)
			return
		}
		if !success {
			if err = os.Remove(tmpFilePath); err != nil {
				log.Printf("%s: filename:%s. Err: %v", op, filename, err)
			}
		}
	}()

	// Записываем данные во временный файл
	var resp *pb.GetFileResponse
	for {
		resp, err = stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("%s: filename:%s. Download completed", op, filename)
				break
			}
			return c.handleGRPCError(op, err)
		}

		if _, err = f.Write(resp.Content); err != nil {
			log.Printf("%s: filename:%s. Err: %v", op, filename, err)
			return fmt.Errorf("%s: filename:%s. Err: %v", op, filename, err)
		}
	}

	// Переименовываем временный файл в целевой
	targetFilename := filepath.Join(c.dataDir, "downloaded_"+filename)
	if err = os.Rename(tmpFilePath, targetFilename); err != nil {
		log.Printf("%s: filename:%s. Err: %v", op, filename, err)
		return fmt.Errorf("%s: filename:%s. Err: %v", op, filename, err)
	}

	// Устанавливаем флаг успешного завершения
	success = true

	log.Printf("%s: file %s downloaded successfully", op, filename)
	return nil
}

func (c *ClientService) handleGRPCError(op string, err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		log.Printf("%s: unexpected error: %v", op, err)
		return fmt.Errorf("unexpected error: %v", err)
	}
	errorDesc := st.Message()
	switch st.Code() {
	case codes.NotFound:
		log.Printf("%s: %v", op, errorDesc)
		return fmt.Errorf("%v", ErrNotFound)
	case codes.ResourceExhausted:
		log.Printf("%s: %v", op, errorDesc)
		return fmt.Errorf("%v", errorDesc)
	default:
		log.Printf("%s: operation failed: %v", op, errorDesc)
		return fmt.Errorf("%s: operation failed: %v. Err: %v", op, ErrInternalServer, errorDesc)
	}
}
