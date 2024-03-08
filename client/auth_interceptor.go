package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

// AuthInterceptor is a client interceptor for authentication
type AuthInterceptor struct {
	authClient  *AuthClient
	authMethods map[string]bool
	accessToken string
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(authClient *AuthClient, authMethods map[string]bool, refreshDuration time.Duration) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
	}
	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

// Unary returns a client interceptor to authenticate unary RPC
func (auth *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Printf("--> unary interceptor: %s", method)

		if auth.authMethods[method] {
			return invoker(auth.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (auth *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", auth.accessToken)
}

// Stream returns a client interceptor to authenticate stream RPC
func (auth *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %s", method)

		if auth.authMethods[method] {
			return streamer(auth.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (auth *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := auth.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			time.Sleep(wait)
			err := auth.refreshToken()
			if err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()

	return nil
}

func (auth *AuthInterceptor) refreshToken() error {
	accessToken, err := auth.authClient.Login()
	if err != nil {
		return err
	}

	auth.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}
