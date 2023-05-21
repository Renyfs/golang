package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"server/service"
)

func main() {
	grpcServer := grpc.NewServer()

	service.RegisterProdServiceServer(grpcServer, service.ProductService{})
	listener, err := net.Listen("tcp", ":8085")
	if err != nil {
		log.Fatalf("启动监听出错：%v\n", err)
	}
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("启动服务出错：%v\n", err)
	}

}
