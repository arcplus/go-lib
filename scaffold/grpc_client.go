package scaffold

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

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
