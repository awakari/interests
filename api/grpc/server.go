package grpc

import (
	"fmt"
	"github.com/awakari/subscriptions/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

func Serve(svc service.Service, port uint16) (err error) {
	c := NewServiceController(svc)
	c = NewAuthMiddleware(c)
	srv := grpc.NewServer()
	RegisterServiceServer(srv, c)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		err = srv.Serve(conn)
	}
	return
}
