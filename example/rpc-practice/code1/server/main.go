package main

import (
	"net"
	"net/http"
	"net/rpc"
)

//  基于 HTTP 的 RPC

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
	rpc.HandleHTTP()
	listener, _ := net.Listen("tcp", ":8085")

	http.Serve(listener, nil)
}
