package main

import (
	internalgrpc "TestTages/internal/grpc"
	"fmt"
	pb "github.com/Sonnnyyy04/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	fileService := internalgrpc.New()
	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Panicf("проблена со слушателем: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, fileService)
	reflection.Register(grpcServer)
	fmt.Println("gRPC Server запущен на порту 50051")
	if err := grpcServer.Serve(l); err != nil {
		log.Fatalf(" проблема с gRPC server: %v", err)
	}
}
