package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"net"
	"server/service"
)

// Token 认证
func main() {
	// 从证书相关文件中读取和解析信息，得到证书公钥、密钥对
	cert, err := tls.LoadX509KeyPair("../TSL/test.pem", "../TSL/test.key")
	if err != nil {
		log.Fatal("证书读取错误", err)
	}
	// 创建一个新的、空的 CertPool
	certPool := x509.NewCertPool()

	ca, err := ioutil.ReadFile("../TSL/server.crt")
	if err != nil {
		log.Fatal("ca证书读取错误", err)
	}
	// 尝试解析所传入的 PEM 编码的证书。如果解析成功会将其加到 CertPool 中，便于后面的使用
	certPool.AppendCertsFromPEM(ca)
	// 构建基于 TLS 的 TransportCredentials 选项
	cred := credentials.NewTLS(&tls.Config{
		// 设置证书链，允许包含一个或多个
		Certificates: []tls.Certificate{cert},
		// 要求必须校验客户端的证书。可以根据实际情况选用以下参数
		ClientAuth: tls.RequireAndVerifyClientCert,
		// 设置根证书的集合，校验方式使用 ClientAuth 中设定的模式
		ClientCAs: certPool,
	})
	//通过拦截器实现认证
	var authInterceptor grpc.UnaryServerInterceptor
	authInterceptor = func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = Auth(ctx)
		if err != nil {
			return
		}
		return handler(ctx, err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(cred), grpc.UnaryInterceptor(authInterceptor))

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

func Auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing credentials")
	}
	var user string
	var password string
	if val, ok := md["user"]; ok {
		user = val[0]
	}
	if val, ok := md["password"]; ok {
		password = val[0]
	}

	if user != "admin" || password != "admin" {
		return status.Errorf(codes.Unauthenticated, "token不合法")
	}
	return nil
}
