package scaffold

import (
	"bytes"
	"context"

	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/arcplus/go-lib/pb"
	"github.com/arcplus/go-lib/scaffold/internal"
)

// ServerErrorConvertor convert *Error to gRPC error
func ServerErrorConvertor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	tid := "x-tracer-id"
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
	buf.Write(pb.MustMarshal(req.(pb.Message)))

	// recover
	defer func() {
		if r := recover(); r != nil {
			logger.Skip(1).Errorf("panic recover grpc: %s\nerr: %v\nstack:\n%s", buf.String(), r, log.TakeStacktrace())
			// if panic, set custom error to 'err', in order that client and sense it.
			err = status.Errorf(codes.Internal, "panic: %v", r)
		}
	}()

	resp, err = handler(ctx, req)

	var code int32
	if err != nil {
		buf.WriteString("\nerr: ")
		buf.WriteString(errs.StackTrace(err))

		// convert normal error to gRPC error
		if _, ok := status.FromError(err); !ok {
			e := errs.ToError(err)

			code = int32(e.Code())

			s := &spb.Status{
				Code:    code,
				Message: e.Message(),
			}

			if alert := e.Alert(); alert != "" {
				s.Details = pb.MarshalAny(&internal.ErrorInfo{
					Alert: alert,
				})
			}

			err = status.ErrorProto(s)
		}
	} else {
		buf.WriteString("\nresp: ")
		buf.Write(pb.MustMarshal(resp.(pb.Message)))
	}

	if err != nil && code < int32(errs.CodeBadRequest) {
		logger.Error(buf.String())
	} else if logger.DebugEnabled() {
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
