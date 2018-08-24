package scaffold

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/arcplus/go-lib/log"
	"google.golang.org/grpc"
)

type Micro struct {
	mu            *sync.Mutex
	errChan       chan error
	resCloseFuncs []func() error // for close func list
}

func NewMicro() *Micro {
	m := &Micro{
		mu:            &sync.Mutex{},
		errChan:       make(chan error, 1),
		resCloseFuncs: make([]func() error, 0, 8),
	}

	m.AddResCloseFunc(log.Close)

	return m
}

// AddResCloseFunc add ResCloseFunc
func (m *Micro) AddResCloseFunc(f func() error) {
	m.mu.Lock()
	m.resCloseFuncs = append(m.resCloseFuncs, f)
	m.mu.Unlock()
}

// TODO ln reuse?
func (m *Micro) createListener(bindAddr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	m.AddResCloseFunc(func() error {
		err := ln.Close()
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				return nil
			}
			return err
		}
		return nil
	})

	return ln, nil
}

// ServeGRPC is helper func to start gRPC server
func (m *Micro) ServeGRPC(bindAddr string, rpcServer, srv interface{}, opts ...grpc.ServerOption) {
	ln, err := m.createListener(bindAddr)
	if err != nil {
		m.errChan <- err
		return
	}

	opts = append(opts, UnaryInterceptor)

	server := grpc.NewServer(opts...)

	m.AddResCloseFunc(func() error {
		server.GracefulStop()
		return nil
	})

	params := []reflect.Value{
		reflect.ValueOf(server),
		reflect.ValueOf(srv),
	}

	defer func() {
		if err := recover(); err != nil {
			m.errChan <- errors.New(fmt.Sprint(err))
		}
	}()

	// check is impl (TODO: optimize)
	if !reflect.TypeOf(srv).Implements(reflect.TypeOf(rpcServer).In(1)) {
		xMap := make(map[string]reflect.Type)

		svcRef := reflect.TypeOf(srv)
		for i, l := 0, svcRef.NumMethod(); i < l; i++ {
			xMap[svcRef.Method(i).Name] = svcRef.Method(i).Type
		}

		rpcRef := reflect.TypeOf(rpcServer).In(1)
		for i, l := 0, rpcRef.NumMethod(); i < l; i++ {
			method := rpcRef.Method(i)

			t, ok := xMap[method.Name]
			if !ok {
				m.errChan <- fmt.Errorf("rpc method %q missing", method.Name)
				return
			}

			if method.Type.NumIn() != t.NumIn()-1 {
				m.errChan <- fmt.Errorf("rpc method %q want:%s, have:%s", method.Name, method.Type, t)
				return
			}

			rpcBuff := &bytes.Buffer{}
			rpcBuff.WriteString("func(")
			implBuff := &bytes.Buffer{}
			implBuff.WriteString("func(")

			var failed bool
			for i, l := 0, method.Type.NumIn(); i < l; i++ {
				if method.Type.In(i) != t.In(i+1) {
					failed = true
				}

				rpcBuff.WriteString(method.Type.In(i).String())
				implBuff.WriteString(t.In(i + 1).String())
				if i != l-1 {
					rpcBuff.WriteString(", ")
					implBuff.WriteString(", ")
				} else {
					rpcBuff.WriteString(") (")
					implBuff.WriteString(") (")
				}
			}

			for i, l := 0, method.Type.NumOut(); i < l; i++ {
				if method.Type.Out(i) != t.Out(i) {
					failed = true
				}

				rpcBuff.WriteString(method.Type.Out(i).String())
				implBuff.WriteString(t.Out(i).String())
				if i != l-1 {
					rpcBuff.WriteString(", ")
					implBuff.WriteString(", ")
				} else {
					rpcBuff.WriteString(")")
					implBuff.WriteString(")")
				}
			}

			if failed {
				m.errChan <- fmt.Errorf("rpc method %q want:%s, have:%s", method.Name, rpcBuff.String(), implBuff.String())
				return
			}
		}

		m.errChan <- errors.New("rpc impl error")
		return
	}

	reflect.ValueOf(rpcServer).Call(params)

	go func() {
		err := server.Serve(ln)
		if err != nil {
			m.errChan <- err
		}
	}()
}

// TODO other params can optimize
func (m *Micro) ServeHTTP(bindAddr string, handler http.Handler) {
	ln, err := m.createListener(bindAddr)
	if err != nil {
		m.errChan <- err
		return
	}

	server := &http.Server{
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 2 << 15, // 64k
	}

	m.AddResCloseFunc(func() error {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
		defer cancelFunc()
		return server.Shutdown(ctx)
	})

	go func() {
		err := server.Serve(ln)
		if err != nil {
			m.errChan <- err
		}
	}()
}

// FILO
func (m *Micro) releaseRes() {
	for i := len(m.resCloseFuncs) - 1; i >= 0; i-- {
		if f := m.resCloseFuncs[i]; f != nil {
			err := f()
			if err != nil {
				log.Errorf("release res err: %s", err.Error())
			}
		}
	}
}

// WatchSignal notify signal to stop running
var WatchSignal = []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT}

// Wait util signal
func (m *Micro) Start() {
	defer m.releaseRes()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, WatchSignal...)
	select {
	case s := <-ch:
		log.Infof("receive stop signal: %s", s)
	case e := <-m.errChan:
		log.Errorf("receive err signal: %s", e)
	}
}
