package scaffold

import (
	"context"
	"runtime"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/arcplus/go-lib/tool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryInterceptorChain build the multi interceptors into one interceptor chain.
func UnaryInterceptorChain(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain, info)
		}
		return chain(ctx, req)
	}
}

// build is the interceptor chain helper
func build(c grpc.UnaryServerInterceptor, n grpc.UnaryHandler, info *grpc.UnaryServerInfo) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return c(ctx, req, info, n)
	}
}

// WrapError wrap *Error to gRPC error
func WrapError(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		err = errs.ToGRPC(err)
		// TODO req may be changed by handler.
		log.Errorf("method:%s, err:%s, req:%s", info.FullMethod, err.Error(), tool.MarshalToString(req))
	}
	return resp, err
}

const (
	MAXSTACKSIZE = 4096
)

// Recovery interceptor to handle grpc panic
func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// recovery func
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, MAXSTACKSIZE)
			stack = stack[:runtime.Stack(stack, false)]
			log.Errorf("recover grpc invoke: %s, err=%v, stack:\n%s\n", info.FullMethod, r, string(stack))
			// if panic, set custom error to 'err', in order that client and sense it.
			err = status.Errorf(codes.Internal, "panic error: %v", r)
		}
	}()

	return handler(ctx, req)
}

// GRPCServeOpts is helper  UnaryInterceptorChain with Recovery and WrapError
var GRPCServeOpts = grpc.UnaryInterceptor(UnaryInterceptorChain(Recovery, WrapError))

// NewGRPCServer is helper func to create *grpc.Server
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(opt, GRPCServeOpts)...)
}
