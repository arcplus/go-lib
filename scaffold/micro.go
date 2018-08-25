package scaffold

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/arcplus/go-lib/log"
	"google.golang.org/grpc"
)

type Micro struct {
	mu            *sync.Mutex
	errChan       chan error
	serveFuncs    []func()
	resCloseFuncs []func() error // for close func list
}

func NewMicro() *Micro {
	m := &Micro{
		mu:            &sync.Mutex{},
		errChan:       make(chan error, 1),
		serveFuncs:    make([]func(), 0),
		resCloseFuncs: make([]func() error, 0, 8),
	}

	m.AddResCloseFunc(log.Close)

	return m
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

// TODO ln reuse?
func (m *Micro) createListener(bindAddr string) (net.Listener, error) {
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

// AddResCloseFunc add ResCloseFunc
func (m *Micro) AddResCloseFunc(f func() error) {
	m.mu.Lock()
	m.resCloseFuncs = append(m.resCloseFuncs, f)
	m.mu.Unlock()
}

// ServeGRPC is helper func to start gRPC server
func (m *Micro) ServeGRPC(bindAddr string, server *grpc.Server) {
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

		log.Debugf("grpc version: %s", grpc.Version)

		err = server.Serve(ln)
		if err != nil {
			m.errChan <- err
		}
	})
}

// TODO other params can optimize
func (m *Micro) ServeHTTP(bindAddr string, handler http.Handler) {
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
func (m *Micro) Start() {
	defer m.releaseRes()

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
