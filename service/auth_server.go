// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"github.com/emorydu/grpc-examples/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer is the server for authentication
type AuthServer struct {
	userStore  UserStore
	jwtManager *JWTManager

	pb.UnimplementedAuthServiceServer
}

// NewAuthServer returns a new auth server
func NewAuthServer(userStore UserStore, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

// Login is a unary RPC to login user
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.userStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {

		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.LoginResponse{AccessToken: token}

	return res, nil
}
