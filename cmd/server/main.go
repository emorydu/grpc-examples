package main

import (
	"flag"
	"fmt"
	"github.com/emorydu/grpc-examples/pb"
	"github.com/emorydu/grpc-examples/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()

	log.Printf("start server on port %d", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")

	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
