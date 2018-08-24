package scaffold

import (
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

// TODO config file
type Micro struct {
	*sync.Mutex
	ErrChan       chan error
	resCloseFuncs []func() error // for close func list
}

func NewMicro() *Micro {
	m := &Micro{
		Mutex:         &sync.Mutex{},
		ErrChan:       make(chan error, 1),
		resCloseFuncs: make([]func() error, 0, 8),
	}

	m.resCloseFuncs = append(m.resCloseFuncs, log.Close)

	return m
}

// AddResCloseFunc add ResCloseFunc
func (m *Micro) AddResCloseFunc(f func() error) {
	m.Lock()
	m.resCloseFuncs = append(m.resCloseFuncs, f)
	m.Unlock()
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
		m.ErrChan <- err
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
			m.ErrChan <- errors.New(fmt.Sprint(err))
		}
	}()

	reflect.ValueOf(rpcServer).Call(params)

	go func() {
		err := server.Serve(ln)
		if err != nil {
			m.ErrChan <- err
		}
	}()
}

// TODO other params can optimize
func (m *Micro) ServeHTTP(bindAddr string, handler http.Handler) {
	ln, err := m.createListener(bindAddr)
	if err != nil {
		m.ErrChan <- err
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
			m.ErrChan <- err
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
		log.Infof("receive signal: '%v'", s)
	case e := <-m.ErrChan:
		log.Errorf("receive err signal: '%v'", e)
	}
}
