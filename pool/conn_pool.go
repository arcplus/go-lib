package pool

import (
	"container/list"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ErrClientPoolExhausted err
var ErrClientPoolExhausted = errors.New("connection pool exhausted")

// ErrClientPoolClosed err
var ErrClientPoolClosed = errors.New("connection pool closed")

// ClientPool is client pool
type ClientPool struct {
	// Dial is an application supplied function for creating new connections.
	Dial func() (interface{}, error)

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c interface{}, t time.Time) error

	// Close is an application supplied function for closing connections.
	Close func(c interface{}) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When le zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// MaxLiveTime is the max live time after conn first established. if zero, no limit
	// MaxLiveTime is used for ease unbalance of tcp conn
	MaxLiveTime time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning, or it may return exhaust err
	Wait bool

	// mu protects fields defined below.
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
	active int

	// Stack of idleConn with most recently used at the front.
	idle list.List
}

// Client client
type Client interface{}

// PooledClient client
type PooledClient interface {
	RawClient() Client
	MarkUnusable()
	Close() error
}

type idleConn struct {
	client   Client
	pool     *ClientPool
	unusable bool
	ft       time.Time // first init time
	lt       time.Time // last active time
}

func init() {
	rand.Seed(time.Now().UnixNano()) // set rand seed
}

func (p *idleConn) RawClient() Client {
	return p.client
}

func (p *idleConn) MarkUnusable() {
	p.unusable = true
}

func (p *idleConn) Close() error {
	return p.pool.put(p.client, p.ft, p.unusable)
}

// Get gets a connection. The application must close the returned connection.
// This method always returns a valid connection so that applications can defer
// error handling to the first use of the connection.
func (p *ClientPool) Get(forceNew ...bool) (PooledClient, error) {
	return p.get(len(forceNew) != 0 && forceNew[0])
}

func (p *ClientPool) get(forceNew bool) (PooledClient, error) {
	p.mu.Lock()
	// if closed
	if p.closed {
		p.mu.Unlock()
		return nil, ErrClientPoolClosed
	}

	// Prune stale connections
	if timeout := p.IdleTimeout; timeout > 0 {
		for i, n := 0, p.idle.Len(); i < n; i++ {
			e := p.idle.Back()
			if e == nil {
				break
			}
			ic := e.Value.(*idleConn)
			if ic.lt.Add(timeout).After(time.Now()) {
				break
			}
			p.idle.Remove(e)
			p.active--
			if p.cond != nil {
				p.cond.Signal()
			}
			p.mu.Unlock()
			p.Close(ic.client)
			p.mu.Lock()
		}
	}

	for {
		if !forceNew {
			// Get idle connection.
			for i, n := 0, p.idle.Len(); i < n; i++ {
				fmt.Println("get idl")
				e := p.idle.Front()
				if e == nil {
					break
				}
				ic := e.Value.(*idleConn)
				p.idle.Remove(e)
				test := p.TestOnBorrow
				p.mu.Unlock()
				if test == nil || test(ic.client, ic.lt) == nil {
					if p.MaxLiveTime <= 0 || ic.ft.Add(withRand(p.MaxLiveTime)).After(time.Now()) {
						return ic, nil
					}
				}
				p.Close(ic.client)
				p.mu.Lock()
				p.active--
				if p.cond != nil {
					p.cond.Signal()
				}
			}
		}

		if p.closed {
			p.mu.Unlock()
			return nil, ErrClientPoolClosed
		}

		if p.MaxActive <= 0 || p.active < p.MaxActive {
			fmt.Println("new one")
			dial := p.Dial
			p.active++
			p.mu.Unlock()
			c, err := dial()
			if err != nil {
				p.mu.Lock()
				p.active--
				if p.cond != nil {
					p.cond.Signal()
				}
				p.mu.Unlock()
				c = nil
			}
			return &idleConn{
				client: c,
				pool:   p,
				ft:     time.Now(),
			}, err
		}

		if !p.Wait {
			p.mu.Unlock()
			return nil, ErrClientPoolExhausted
		}

		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}
		p.cond.Wait()
	}
}

// Put adds conn back to the pool, use forceClose to close the connection force
func (p *ClientPool) put(c interface{}, firstInit time.Time, forceClose bool) error {
	p.mu.Lock()
	if !p.closed && !forceClose {
		p.idle.PushFront(&idleConn{client: c, pool: p, ft: firstInit, lt: time.Now()})
		if p.idle.Len() > p.MaxIdle {
			c = p.idle.Remove(p.idle.Back()).(*idleConn).client
		} else {
			c = nil
		}
	}

	if c == nil {
		if p.cond != nil {
			p.cond.Signal()
		}
		p.mu.Unlock()
		return nil
	}

	p.active--
	if p.cond != nil {
		p.cond.Signal()
	}
	p.mu.Unlock()
	return p.Close(c)
}

// ActiveCount returns the number of active connections in the pool.
func (p *ClientPool) ActiveCount() int {
	p.mu.Lock()
	active := p.active
	p.mu.Unlock()
	return active
}

// IdleCount returns the number of idle connections in the pool.
func (p *ClientPool) IdleCount() int {
	p.mu.Lock()
	idle := p.idle.Len()
	p.mu.Unlock()
	return idle
}

// Release releases the resources used by the pool.
func (p *ClientPool) Release() error {
	p.mu.Lock()
	idle := p.idle
	p.idle.Init()
	p.closed = true
	p.active -= idle.Len()
	if p.cond != nil {
		p.cond.Broadcast()
	}
	p.mu.Unlock()
	for e := idle.Front(); e != nil; e = e.Next() {
		p.Close(e.Value.(*idleConn).client)
	}
	return nil
}

func withRand(t time.Duration) time.Duration {
	if t <= time.Minute*5 {
		return t + time.Second*time.Duration(rand.Intn(30))
	}
	return t + time.Second*time.Duration(rand.Intn(60))
}
