package grpc

import (
	"context"
	"fmt"
	pb "github.com/Sonnnyyy04/protos/gen/go"
	"github.com/djherbis/times"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxUploadDownload  = 10
	maxListConnections = 100
	layoutForDate      = "2006-01-02"
	storagePath        = "./storage" //лучше вынести в config
)

type FileService struct {
	pb.UnimplementedFileServiceServer
	UploadDownloadSem chan struct{}
	ListFileSem       chan struct{}
	FileMutex         sync.RWMutex
}

func New() *FileService {
	return &FileService{
		UploadDownloadSem: make(chan struct{}, maxUploadDownload),
		ListFileSem:       make(chan struct{}, maxListConnections),
	}
}

func (f *FileService) UploadFile(stream pb.FileService_UploadFileServer) error {
	f.UploadDownloadSem <- struct{}{}
	defer func() {
		<-f.UploadDownloadSem
	}()
	firstChunk, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "не удалось получить первый чанк: %v", err)
	}
	filename := firstChunk.GetFilename()
	if filename == "" {
		return status.Error(codes.InvalidArgument, "имя файла не должно быть пустым")
	}
	filePath := filepath.Join(storagePath, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось создать файл: %v", err)
	}
	defer file.Close()
	log.Printf("Начата загрузка файла: %s", filename)
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Файл %s успешно загружен", filename)
			return stream.SendAndClose(&pb.UploadFileResponse{Message: "Файл загружен успешно"})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "ошибка при получении чанка: %v", err)
		}
		_, err = file.Write(chunk.GetData())
		if err != nil {
			return status.Errorf(codes.Internal, "не удалось записать чанк: %v", err)
		}
	}
}

func (f *FileService) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	f.ListFileSem <- struct{}{}
	defer func() {
		<-f.ListFileSem
	}()
	//ctx
	f.FileMutex.RLock()
	defer f.FileMutex.RUnlock()
	records, err := os.ReadDir(storagePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось прочитать директорию: %v", err)
	}
	var files []*pb.FileMetadata
	for _, record := range records {
		time, err := statTime(record.Name())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "не удалось получить время файла: %v", err)
		}
		files = append(files, &pb.FileMetadata{
			Filename:  record.Name(),
			CreatedAt: time.BirthTime().Format(layoutForDate),
			UpdatedAt: time.ModTime().Format(layoutForDate),
		})
	}
	log.Printf("Отправлен список из %d файлов", len(files))
	return &pb.ListFilesResponse{Files: files}, nil
}

func statTime(fileName string) (times.Timespec, error) {
	filePath := filepath.Join(storagePath, fileName)
	t, err := times.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения времени для %s: %w", filePath, err)
	}
	return t, err
}

func (f *FileService) DownloadFiles(req *pb.DownloadFilesRequest, stream pb.FileService_DownloadFilesServer) error {
	f.UploadDownloadSem <- struct{}{}
	defer func() {
		<-f.UploadDownloadSem
	}()
	filePath := filepath.Join(storagePath, req.GetFilename())
	file, err := os.Open(filePath)
	if err != nil {
		return status.Errorf(codes.NotFound, "файл не найден: %v", err)
	}
	defer file.Close()

	log.Printf("Начата передача файла: %s", req.GetFilename())
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "не удалось прочитать файл: %v", err)
		}
		if err := stream.Send(&pb.DownloadFilesResponse{Message: buf[:n]}); err != nil {
			return status.Errorf(codes.Internal, "не удалось отправить данные: %v", err)
		}
	}
	return nil
}
