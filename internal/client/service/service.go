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
	"time"
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

func (c *ClientService) UploadFile(ctx context.Context, filename string) error {
	const op = "client.service.UploadFile"

	// Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s: %v", op, ErrNotFound)
			return ErrNotFound
		}
		log.Printf("%s: failed to open file: %v", op, err)
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("%s: failed to close file: %v", op, err)
		}
	}()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, time.Second*60) // Увеличен таймаут
	defer cancel()

	// Создаем поток для загрузки файла
	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		return c.handleGRPCError(op, err)
	}

	// Отправляем имя файла
	if err = stream.Send(&pb.UploadFileRequest{Filename: filepath.Base(filename)}); err != nil {
		log.Printf("%s: failed to send filename: %v", op, err)
		return fmt.Errorf("failed to send filename: %v", err)
	}
	log.Printf("%s: sent filename: %s", op, filepath.Base(filename))

	// Передача данных
	buf := make([]byte, defaultBufSize)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("%s: failed to read file: %v", op, err)
			return fmt.Errorf("failed to read file: %v", err)
		}

		if n > 0 {
			if err := stream.Send(&pb.UploadFileRequest{Content: buf[:n]}); err != nil {
				log.Printf("%s: failed to send file chunk: %v", op, err)
				return fmt.Errorf("failed to send file chunk: %v", err)
			}
			log.Printf("%s: sent %d bytes", op, n) // Лог отправленных байт
		}
	}

	// Завершаем поток и получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("%s: failed to receive response: %v", op, err)
		return fmt.Errorf("failed to receive response: %v", err)
	}

	log.Printf("%s: server response: %s", op, resp.Message)
	return nil
}

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

	if err = os.MkdirAll(c.dataDir, os.ModePerm); err != nil {
		log.Printf("%s: failed to create directory: %v", op, err)
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Создаем временный файл
	tmpFilename := "downloaded_" + filename + ".tmp"
	tmpFilePath := filepath.Join(c.dataDir, tmpFilename)
	f, err := os.Create(tmpFilePath)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s: filepath:%s. Err: %v", op, fmt.Sprintf(c.dataDir+filename), ErrNotFound)
		return fmt.Errorf("%s: filepath:%s. Err: %v", op, fmt.Sprintf(c.dataDir+filename), ErrNotFound)
	}
	if err != nil {
		log.Printf("%s: failed to create temporary file: %v", op, err)
		return fmt.Errorf("failed to create temporary file: %v", err)
	}

	// Удаляем временный файл в случае ошибки
	var success bool
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("%s: failed to close temporary file: %v", op, err)
		}
		if !success {
			if err := os.Remove(tmpFilePath); err != nil {
				log.Printf("%s: failed to remove temporary file: %v", op, err)
			}
		}
	}()

	// Записываем данные во временный файл
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("%s: file download completed: %s", op, filename)
				break
			}
			return c.handleGRPCError(op, err)
		}

		if _, err = f.Write(resp.Content); err != nil {
			log.Printf("%s: failed to write to file: %v", op, err)
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}

	// Переименовываем временный файл в целевой
	targetFilename := filepath.Join(c.dataDir, "downloaded_"+filename)
	if err = os.Rename(tmpFilePath, targetFilename); err != nil {
		log.Printf("%s: failed to rename temporary file: %v", op, err)
		return fmt.Errorf("failed to rename temporary file: %v", err)
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
