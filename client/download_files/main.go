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
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("проблема с подключением: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	filename := "copy_1.png"
	req := pb.DownloadFilesRequest{
		Filename: filename,
	}
	stream, err := client.DownloadFiles(context.Background(), &req)
	if err != nil {
		log.Panicf("ошибка при запросе потока: %v", err)
	}
	err = os.MkdirAll("downloads", os.ModePerm)
	if err != nil {
		log.Panicf("не удалось создать папку: %v", err)
	}
	path := fmt.Sprintf("downloads/%s", "downloaded_"+filename)
	outFile, err := os.Create(path)
	if err != nil {
		log.Panicf("не удалось создать файл: %v", err)
	}
	defer outFile.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("Загрузка завершена успешно")
			break
		}
		if err != nil {
			log.Panicf("ошибка при получении данных: %v", err)
		}
		_, err = outFile.Write(resp.GetMessage())
		if err != nil {
			log.Fatalf("ошибка при записи в файл: %v", err)
		}
	}

}
