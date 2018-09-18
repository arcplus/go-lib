package scaffold

import (
	"context"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/scaffold/internal"

	"github.com/golang/protobuf/ptypes"
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

		spb := s.Proto()
		err = errs.NewRaw(errs.ErrCode(spb.Code), spb.Message)

		if len(spb.Details) != 0 {
			errInfo := &internal.ErrorInfo{}
			ptypes.UnmarshalAny(spb.Details[0], errInfo)
			if errInfo.Alert != "" {
				errs.WithAlert(err, errInfo.Alert)
			}
		}

		return err
	}

	return nil
}
