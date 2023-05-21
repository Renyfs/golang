package main

import (
	"client/auth"
	"client/service"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
)

func main() {
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, _ := tls.LoadX509KeyPair("../TSL/client.pem", "../TSL/client.key")
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()
	ca, _ := ioutil.ReadFile("../TSL/server.crt")
	// 尝试解析所传入的 PEM 编码的证书。如果解析成功会将其加到 CertPool 中，便于后面的使用
	certPool.AppendCertsFromPEM(ca)
	// 构建基于 TLS 的 TransportCredentials 选项
	cred := credentials.NewTLS(&tls.Config{
		// 设置证书链，允许包含一个或多个
		Certificates: []tls.Certificate{cert},
		// 要求必须校验客户端的证书。可以根据实际情况选用以下参数
		ServerName: "*.renyf.com",
		RootCAs:    certPool,
	})
	token := &auth.Authentication{
		User:     "admin",
		Password: "admin",
	}
	conn, err := grpc.Dial(":8085", grpc.WithTransportCredentials(cred), grpc.WithPerRPCCredentials(token))
	defer conn.Close()
	if err != nil {
		log.Fatalf("build conn failed:%v\n", err)
	}

	client := service.NewProdServiceClient(conn)

	response, err := client.GetProductStock(context.Background(), &service.ProductRequest{ProdId: 110999000})
	if err != nil {
		log.Fatalf("查询库存出错：%v\n", err)
	}
	fmt.Println("ProductStock = ", response.GetProdStock())
}
