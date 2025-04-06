package main

import (
	"context"
	"fmt"
	pb "github.com/Sonnnyyy04/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := uploadFile(fmt.Sprintf("copy_%d.png", i))
			if err != nil {
				log.Printf("Ошибка загрузки #%d: %v\n", i, err)
			} else {
				log.Printf("Успешно загружен файл #%d\n", i)
			}
		}()
	}
	wg.Wait()
}

func uploadFile(remoteName string) error {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("не удалось подключиться: %w", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	filePath := "./gopher.png"
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка при открытии файла: %w", err)
	}
	defer file.Close()

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		return fmt.Errorf("ошибка при вызове UploadFile: %w", err)
	}

	// Отправка первого чанка
	firstReq := &pb.UploadFileRequest{
		Data:     nil,
		Filename: remoteName,
	}
	if err := stream.Send(firstReq); err != nil {
		return fmt.Errorf("ошибка при отправке первого чанка: %w", err)
	}

	// Отправка остальной чанков
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка при чтении файла: %w", err)
		}
		req := &pb.UploadFileRequest{
			Data: buf[:n],
		}
		if err := stream.Send(req); err != nil {
			return fmt.Errorf("ошибка при отправке чанка: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("ошибка при получении ответа: %w", err)
	}
	log.Println("Ответ сервера:", resp.GetMessage())
	return nil
}
