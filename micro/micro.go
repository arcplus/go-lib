package micro

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/arcplus/go-lib/log"
)

// go build -ldflags -X
var version, gitCommit, buildDate string

// VersionInfo return visualized version info
func VersionInfo() string {
	buff := &bytes.Buffer{}
	w := tabwriter.NewWriter(buff, 0, 0, 0, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "version: \t"+version)
	fmt.Fprintln(w, "gitCommit: \t"+gitCommit)
	fmt.Fprintln(w, "buildDate: \t"+buildDate) // trailing tab
	w.Flush()
	return buff.String()
}

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

var mode = os.Getenv("mode")

func ProdMode() bool {
	return mode == "prod" || mode == "production"
}

func Mode() string {
	if mode == "" {
		return "dev"
	}
	return mode
}

// New create Micro, moduleName.0 is module name.
func New(moduleName ...string) Micro {
	m := &micro{
		mu:            &sync.Mutex{},
		errChan:       make(chan error, 1),
		serveFuncs:    make([]func(), 0),
		resCloseFuncs: list.New(),
	}

	kv := map[string]interface{}{}

	if len(moduleName) != 0 {
		kv["module"] = moduleName[0]
	}

	if mode != "" {
		kv["mode"] = mode
	}

	if gitCommit != "" {
		kv["version"] = version + "_" + gitCommit
	}

	log.SetAttachment(kv)

	level := getLogLevel(os.Getenv("log_level"))
	log.SetGlobalLevel(level)

	ws := []io.Writer{}

	async := os.Getenv("log_sync") == "true"

	if os.Getenv("log_std_disable") != "true" {
		ws = append(ws, log.ConsoleWriter(
			log.ConsoleConfig{
				Async: async,
			},
		))
	}

	if rds, key := os.Getenv("log_rds_dsn"), os.Getenv("log_rds_key"); rds != "" && key != "" {
		ws = append(ws, log.RedisWriter(log.RedisConfig{
			Level:  getLogLevel(os.Getenv("log_rds_level")),
			DSN:    rds,
			LogKey: key,
			Async:  async,
		}))
	}

	if len(ws) != 0 {
		log.SetOutput(ws...)
	}

	m.AddResCloseFunc(log.Close)

	return m
}

// Close close all added resource FILO
func (m *micro) Close() {
	m.mu.Lock()
	for m.resCloseFuncs.Len() != 0 {
		e := m.resCloseFuncs.Back()
		if f, ok := e.Value.(func() error); ok && f != nil {
			err := f()
			if err != nil {
				log.Errorf("close resource err: %s", err.Error())
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

	log.Info("micro start")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, WatchSignal...)
	select {
	case s := <-ch:
		log.Skip(1).Infof("micro receive stop signal: %s", s)
	case e := <-m.errChan:
		log.Skip(1).Errorf("micro receive err signal: %s", e)
	}
}

// bind is a helper func to read env port and returns bind addr
func Bind(port string, envName ...string) string {
	env := "p"
	if len(envName) != 0 {
		env = envName[0]
	}

	if p := os.Getenv(env); p != "" {
		port = p
	}

	// :port, it panics if port is empty
	if port[0] == ':' {
		return port
	}

	// port only, yes, it may return :0, caution
	if _, err := strconv.Atoi(port); err == nil {
		return ":" + port
	}

	return port
}

func getLogLevel(str string) log.Level {
	switch str {
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.DebugLevel
	}
}
