package scaffold

import (
	"net"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

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

// ServeGRPC is helper func to start gRPC server
func (m *Micro) ServeGRPC(bindAddr string, rpcServer, srv interface{}, opts ...grpc.ServerOption) {
	ln, err := net.Listen("tcp", bindAddr)
	if err != nil {
		m.ErrChan <- err
		return
	}

	m.AddResCloseFunc(ln.Close)

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

	// TODO maybe should handle panic error
	reflect.ValueOf(rpcServer).Call(params)

	go func() {
		err := server.Serve(ln)
		if err != nil {
			m.ErrChan <- err
		}
	}()
}

// FILO
func (m *Micro) releaseRes() {
	for i := len(m.resCloseFuncs) - 1; i > 0; i-- {
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
