package micro

import (
	"container/list"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/arcplus/go-lib/log"
)

type Micro interface {
	AddResCloseFunc(f func() error)
	Close()
	ServeGRPC(bindAddr string, server GRPCServer)
	ServeHTTP(bindAddr string, handler http.Handler)
	Start()
}

type micro struct {
	mu            *sync.Mutex
	errChan       chan error
	serveFuncs    []func()
	resCloseFuncs *list.List
}

func New() Micro {
	m := &micro{
		mu:            &sync.Mutex{},
		errChan:       make(chan error, 1),
		serveFuncs:    make([]func(), 0),
		resCloseFuncs: list.New(),
	}

	m.AddResCloseFunc(log.Close)

	return m
}

// FILO
func (m *micro) Close() {
	m.mu.Lock()
	for e := m.resCloseFuncs.Back(); e != nil; {
		if f, ok := e.Value.(func() error); ok && f != nil {
			err := f()
			if err != nil {
				log.Errorf("release res err: %s", err.Error())
			}
		}
		m.resCloseFuncs.Remove(e)
	}
	m.mu.Unlock()
}

// TODO ln reuse?
func (m *micro) createListener(bindAddr string) (net.Listener, error) {
	m.mu.Lock()
	ln, err := net.Listen("tcp", bindAddr)
	if err != nil {
		m.mu.Unlock()
		return nil, err
	}
	m.mu.Unlock()

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

// AddResCloseFunc add resource close func
func (m *micro) AddResCloseFunc(f func() error) {
	m.mu.Lock()
	m.resCloseFuncs.PushBack(f)
	m.mu.Unlock()
}

// GRPCServer
type GRPCServer interface {
	Serve(net.Listener) error
	GracefulStop()
}

// ServeGRPC is helper func to start gRPC server
func (m *micro) ServeGRPC(bindAddr string, server GRPCServer) {
	m.serveFuncs = append(m.serveFuncs, func() {
		ln, err := m.createListener(bindAddr)
		if err != nil {
			m.errChan <- err
			return
		}

		m.AddResCloseFunc(func() error {
			server.GracefulStop()
			return nil
		})

		err = server.Serve(ln)
		if err != nil {
			m.errChan <- err
		}
	})
}

// TODO other params can optimize
func (m *micro) ServeHTTP(bindAddr string, handler http.Handler) {
	m.serveFuncs = append(m.serveFuncs, func() {
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

		err = server.Serve(ln)
		if err != nil {
			m.errChan <- err
		}
	})
}

// WatchSignal notify signal to stop running
var WatchSignal = []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT}

// Wait util signal
func (m *micro) Start() {
	defer m.Close()

	for i := range m.serveFuncs {
		go m.serveFuncs[i]()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, WatchSignal...)
	select {
	case s := <-ch:
		log.Infof("receive stop signal: %s", s)
	case e := <-m.errChan:
		log.Errorf("receive err signal: %s", e)
	}
}
