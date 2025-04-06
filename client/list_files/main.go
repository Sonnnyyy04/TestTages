package main

import (
	"context"
	pb "github.com/Sonnnyyy04/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("что-то пошло не так: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		log.Panicf("что-то пошло не так: %v", err)
	}
	log.Printf("найденные %d файлы:", len(resp.Files))
	for _, file := range resp.Files {
		log.Printf("- %s (created: %s, modified: %s)",
			file.Filename,
			file.CreatedAt,
			file.UpdatedAt,
		)
	}
}
