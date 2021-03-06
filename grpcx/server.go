package grpcx

import (
	"bytes"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/arcplus/go-lib/pb"
)

var (
	MaxRecvMsgSize = grpc.MaxRecvMsgSize
	MaxSendMsgSize = grpc.MaxSendMsgSize
)

// NewServer is helper func to create *grpc.Server
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append([]grpc.ServerOption{WithUnaryServerChain(ServerErrorConvertor)}, opts...)
	return grpc.NewServer(opts...)
}

// ServerErrorConvertor convert *Error to gRPC error
func ServerErrorConvertor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	tid := "x-request-id"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if t := md.Get("x-request-id"); len(t) != 0 {
			tid = t[0]
		}
	}

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
			logger.Skip(1).Errorf("grpc panic recover: %s\nerr: %v\nstack:\n%s", buf.String(), r, log.TakeStacktrace())
			// if panic, set custom error to 'err', in order that client and sense it.
			err = status.Errorf(codes.Internal, "panic: %v", r)
		}
	}()

	resp, err = handler(context.WithValue(ctx, "x-request-id", tid), req)

	var code uint32
	if err != nil {
		buf.WriteString("\nerr: ")
		buf.WriteString(errs.StackTrace(err))

		// convert normal error to gRPC error
		if _, ok := status.FromError(err); !ok {
			e := errs.ToError(err)
			code = e.Code()
			err = status.Error(codes.Code(e.Code()), e.Message())
		}
	} else {
		buf.WriteString("\nresp: ")
		buf.Write(pb.MustMarshal(resp.(pb.Message)))
	}

	if logger.DebugEnabled() {
		logger.Debug(buf.String())
	} else if err != nil && code < errs.CodeBadRequest {
		logger.Error(buf.String())
	}

	return resp, err
}
