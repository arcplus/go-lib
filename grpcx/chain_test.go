package grpcx

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// test case from: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/chain_test.go

var (
	someValue       = 1024
	parentContext   = context.WithValue(context.Background(), "parent", someValue)
	someServiceName = "SomeService.StreamMethod"
	parentUnaryInfo = &grpc.UnaryServerInfo{FullMethod: someServiceName}

	parentStreamInfo = &grpc.StreamServerInfo{
		FullMethod:     someServiceName,
		IsServerStream: true,
	}
)

func TestChainUnaryServer(t *testing.T) {
	noChain := true
	input := "input"
	output := "output"

	first := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ctx.Value("parent").(int) != someValue {
			t.Fatal("first interceptor must know the parent context value")
		}

		if !reflect.DeepEqual(parentUnaryInfo, info) {
			t.Fatal("first interceptor must know the someUnaryServerInfo")
		}

		ctx = context.WithValue(ctx, "first", 1)
		return handler(ctx, req)
	}

	second := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ctx.Value("parent").(int) != someValue {
			t.Fatal("second interceptor must know the parent context value")
		}

		if ctx.Value("first") == nil {
			t.Fatal("second interceptor must know the first context value")
		}

		if !reflect.DeepEqual(parentUnaryInfo, info) {
			t.Fatal("second interceptor must know the someUnaryServerInfo")
		}

		ctx = context.WithValue(ctx, "second", 1)
		return handler(ctx, req)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		if req.(string) != input {
			t.Fatal("handler must get the input")
		}

		if ctx.Value("parent").(int) != someValue {
			t.Fatal("handler must know the parent context value")
		}

		if noChain {
			return noChain, nil
		}

		if ctx.Value("first") == nil {
			t.Fatal("handler must know the first context value")
		}

		if ctx.Value("second") == nil {
			t.Fatal("handler must know the second context value")
		}

		return output, nil
	}

	chain := ChainUnaryServer()
	out, err := chain(parentContext, input, parentUnaryInfo, handler)
	if err != nil {
		t.Fatal(err)
	}
	if !out.(bool) {
		t.Fatal("chain must return handler's noChan")
	}

	noChain = false
	chain = ChainUnaryServer(first, second)
	out, err = chain(parentContext, input, parentUnaryInfo, handler)
	if err != nil {
		t.Fatal(err)
	}
	if out.(string) != out {
		t.Fatal("chain must return handler's output")
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx         context.Context
	recvMessage interface{}
	sentMessage interface{}
}

func (f *fakeServerStream) Context() context.Context {
	return f.ctx
}

func (f *fakeServerStream) SendMsg(m interface{}) error {
	if f.sentMessage != nil {
		return grpc.Errorf(codes.AlreadyExists, "fakeServerStream only takes one message, sorry")
	}
	f.sentMessage = m
	return nil
}

func (f *fakeServerStream) RecvMsg(m interface{}) error {
	if f.recvMessage == nil {
		return grpc.Errorf(codes.NotFound, "fakeServerStream has no message, sorry")
	}
	return nil
}

func TestChainStreamServer(t *testing.T) {
	noChain := true
	someService := &struct{}{}
	recvMessage := "received"
	sentMessage := "sent"
	outputError := errors.New("some error")

	first := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if stream.Context().Value("parent").(int) != someValue {
			t.Fatal("first interceptor must know the parent context value")
		}

		if !reflect.DeepEqual(info, parentStreamInfo) {
			t.Fatal("first interceptor must know the parentStreamInfo")
		}

		if !reflect.DeepEqual(srv, someService) {
			t.Fatal("first interceptor must know someService")
		}

		wrapped := WrapServerStream(stream)
		wrapped.WrappedContext = context.WithValue(stream.Context(), "first", 1)
		return handler(srv, wrapped)
	}

	second := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if stream.Context().Value("parent").(int) != someValue {
			t.Fatal("second interceptor must know the parent context value")
		}

		if stream.Context().Value("first") == nil {
			t.Fatal("second interceptor must know the first context value")
		}

		if !reflect.DeepEqual(info, parentStreamInfo) {
			t.Fatal("second interceptor must know the parentStreamInfo")
		}

		if !reflect.DeepEqual(srv, someService) {
			t.Fatal("second interceptor must know someService")
		}

		wrapped := WrapServerStream(stream)
		wrapped.WrappedContext = context.WithValue(stream.Context(), "second", 1)
		return handler(srv, wrapped)
	}

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		if stream.Context().Value("parent").(int) != someValue {
			t.Fatal("handler must know the parent context value")
		}

		if !reflect.DeepEqual(srv, someService) {
			t.Fatal("handler must know someService")
		}

		err := stream.RecvMsg(recvMessage)
		if err != nil {
			t.Fatal("handler must have access to recv stream messages")
		}

		if noChain {
			return nil
		}

		err = stream.SendMsg(sentMessage)
		if err != nil {
			t.Fatal("handler must have access to send stream messages", err)
		}

		if stream.Context().Value("first") == nil {
			t.Fatal("handler must know the first context value")
		}

		if stream.Context().Value("second") == nil {
			t.Fatal("handler must know the second context value")
		}

		return outputError
	}

	fakeStream := &fakeServerStream{
		ctx:         parentContext,
		recvMessage: recvMessage,
	}

	chain := ChainStreamServer()
	err := chain(someService, fakeStream, parentStreamInfo, handler)
	if err != nil {
		t.Fatal(err)
	}

	noChain = false
	chain = ChainStreamServer(first, second)
	err = chain(someService, fakeStream, parentStreamInfo, handler)
	if err != outputError {
		t.Fatal("chain must return handler's error")
	}

	if fakeStream.sentMessage.(string) != sentMessage {
		t.Fatal("handler's sent message must propagate to stream")
	}
}
