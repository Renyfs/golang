package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"server/service"
)

func main() {
	cred, err := credentials.NewServerTLSFromFile("../TSL/test.pem", "../TSL/test.key")
	if err != nil {
		fmt.Println("证书生成失败")
	}
	grpcServer := grpc.NewServer(grpc.Creds(cred))

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
