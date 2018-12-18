package grpcx

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type grpcErrorWrapper struct {
	s *status.Status
}

func (e *grpcErrorWrapper) Code() uint32 {
	return uint32(e.s.Code())
}

func (e *grpcErrorWrapper) Message() string {
	return e.s.Message()
}

func (e *grpcErrorWrapper) Error() string {
	return e.s.Err().Error()
}

func (e *grpcErrorWrapper) GRPCStatus() *status.Status {
	return e.s
}

func ClientErrorConvertor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		// this must be gRPC error
		s, ok := status.FromError(err)
		if !ok {
			return err
		}

		return &grpcErrorWrapper{
			s: s,
		}
	}

	return nil
}
