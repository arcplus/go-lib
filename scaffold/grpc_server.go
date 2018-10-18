package scaffold

import (
	"bytes"
	"context"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/arcplus/go-lib/scaffold/internal"
	"github.com/arcplus/go-lib/tool"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServerErrorConvertor convert *Error to gRPC error
func ServerErrorConvertor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	tid := "x-mock-id"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if t := md.Get("x-request-id"); len(t) != 0 && t[0] != "" {
			tid = t[0]
		}
	}

	ctx = context.WithValue(ctx, "x-request-id", tid)

	logger := log.Trace(tid)

	// TODO using pool
	buf := &bytes.Buffer{}

	buf.WriteString("method: ")
	buf.WriteString(info.FullMethod)

	buf.WriteString("\nreq: ")
	buf.WriteString(tool.MarshalToString(req))

	// recover
	defer func() {
		if r := recover(); r != nil {
			logger.Skip(1).Errorf("recover grpc: %s\nerr: %v\nstack:\n%s", buf.String(), r, log.TakeStacktrace())
			// if panic, set custom error to 'err', in order that client and sense it.
			err = status.Errorf(codes.Internal, "panic: %v", r)
		}
	}()

	resp, err = handler(ctx, req)
	var code int32
	if err != nil {
		if _, ok := status.FromError(err); !ok {
			e := errs.ToError(err)

			code = int32(e.Code())

			s := &spb.Status{
				Code:    code,
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

	buf.WriteString("\nresp: ")
	buf.WriteString(tool.MarshalToString(resp))

	buf.WriteString("\nerr: ")
	buf.WriteString(tool.MarshalToString(err))

	if err != nil && code < 1400 {
		logger.Error(buf.String())
	} else {
		logger.Debug(buf.String())
	}

	return resp, err
}

// GRPCServeOpts is helper  UnaryInterceptorChain with Recovery and WrapError
var GRPCServerOpts = grpc.UnaryInterceptor(ServerErrorConvertor)

// NewGRPCServer is helper func to create *grpc.Server
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(opt, GRPCServerOpts)...)
}
