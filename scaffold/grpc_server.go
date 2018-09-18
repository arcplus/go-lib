package scaffold

import (
	"context"
	"runtime"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/arcplus/go-lib/scaffold/internal"
	"github.com/arcplus/go-lib/tool"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServerErrorConvertor convert *Error to gRPC error
func ServerErrorConvertor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		log.Skip(1).Errorf("method: %s\nerr: %s\nreq:%s", info.FullMethod, errs.StackTrace(err), tool.MarshalToString(req))
		if _, ok := status.FromError(err); !ok {
			e := errs.ToError(err)

			s := &spb.Status{
				Code:    int32(e.Code()),
				Message: e.Message(),
			}

			if alert := e.Alert(); alert != "" {
				errInfo := &internal.ErrorInfo{
					Alert: alert,
				}
				a, _ := ptypes.MarshalAny(errInfo)
				s.Details = []*any.Any{a}
			}

			err = status.ErrorProto(s)
		}
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
			log.Skip(1).Errorf("recover grpc invoke: %s\nerr: %v\nstack:\n%s", info.FullMethod, r, string(stack))
			// if panic, set custom error to 'err', in order that client and sense it.
			err = status.Errorf(codes.Internal, "panic error: %v", r)
		}
	}()

	return handler(ctx, req)
}

// GRPCServeOpts is helper  UnaryInterceptorChain with Recovery and WrapError
var GRPCServerOpts = grpc.UnaryInterceptor(UnaryInterceptorChain(Recovery, ServerErrorConvertor))

// NewGRPCServer is helper func to create *grpc.Server
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(opt, GRPCServerOpts)...)
}
