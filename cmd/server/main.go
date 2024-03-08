// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/emorydu/grpc-examples/pb"
	"github.com/emorydu/grpc-examples/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 10 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()

	log.Printf("start server on port %d", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

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

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/proto.LaptopService/"

	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin", "secret", "admin")
	if err != nil {
		return err
	}

	return createUser(userStore, "user", "secret", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}
