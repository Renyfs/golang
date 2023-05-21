package main

import (
	"client/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {

	conn, err := grpc.Dial(":8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatalf("build conn failed:%v\n", err)
	}

	client := service.NewProdServiceClient(conn)

	response, err := client.GetProductStock(context.Background(), &service.ProductRequest{ProdId: 1100})
	if err != nil {
		log.Fatalf("查询库存出错：%v\n", err)
	}
	fmt.Println("ProductStock = ", response.GetProdStock())
}
