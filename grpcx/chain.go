package grpcx

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryServer build the multi interceptors into one interceptor chain.
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildServerUnaryInterceptor(interceptors[i], info, chain)
		}
		return chain(ctx, req)
	}
}

// build is the interceptor chain helper
func buildServerUnaryInterceptor(c grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return c(ctx, req, info, handler)
	}
}

// ChainStreamServer build the multi interceptors into one interceptor chain.
func ChainStreamServer(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildServerStreamInterceptor(interceptors[i], info, chain)
		}
		return chain(srv, ss)
	}
}

func buildServerStreamInterceptor(c grpc.StreamServerInterceptor, info *grpc.StreamServerInfo, handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		return c(srv, stream, info, handler)
	}
}

// WithUnaryServerChain is a grpc.Server config option that accepts multiple unary interceptors.
func WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(ChainUnaryServer(interceptors...))
}

// WithStreamServerChain is a grpc.Server config option that accepts multiple stream interceptors.
func WithStreamServerChain(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption {
	return grpc.StreamInterceptor(ChainStreamServer(interceptors...))
}
