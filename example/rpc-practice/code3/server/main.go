package main

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

//  基于 TCP 的 RPC

type Args struct {
	X, Y int
}

// serviceA
type ServiceA struct{}

// 为 serviceA 注册方法
func (s *ServiceA) Add(args *Args, reply *int) error {
	*reply = args.X + args.Y
	return nil
}

func main() {
	service := new(ServiceA)
	rpc.Register(service)

	listener, _ := net.Listen("tcp", ":8085")

	for {
		conn, _ := listener.Accept()
		rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
