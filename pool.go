package wire

import (
	"sync"
	"time"
)

func NewPool(dialer Dialer, idleTimeout time.Duration) *Pool {
	if dialer == nil {
		panic("wire.NewPool: dialer == 0")
	}
	if idleTimeout <= 0 {
		panic("wire.NewPool: idleTimeout <= 0")
	}

	return &Pool{
		dialer:  dialer,
		timeout: idleTimeout,
	}
}

// A Pool extends a Dialer with support for reusing connections.
type Pool struct {
	// Underlying Dialer.
	dialer Dialer

	// Connection keep-alive duration.
	timeout time.Duration

	// Protects internal fields.
	mu sync.Mutex

	// Idle TCP and TLS connections (stored as linked lists).
	idleTCP map[string]*poolConn
	idleTLS map[string]*poolConn

	// Whether or not the garbage collection loop is currently running.
	looping bool
}

func (p *Pool) DialTCP(addr string) (Conn, error) {
	conn := p.first(p.idleTCP, addr)
	if conn != nil {
		return conn, nil
	}

	// Attempt to establish a new connection.
	conn, err := p.dialer.DialTCP(addr)
	if err != nil {
		return nil, err
	}

	return &poolConn{
		Conn: conn,
		pool: p,
		addr: addr,
		tls:  false,
	}, nil
}

func (p *Pool) DialTLS(addr string) (Conn, error) {
	conn := p.first(p.idleTLS, addr)
	if conn != nil {
		return conn, nil
	}

	// Attempt to establish a new connection.
	conn, err := p.dialer.DialTLS(addr)
	if err != nil {
		return nil, err
	}

	return &poolConn{
		Conn: conn,
		pool: p,
		addr: addr,
		tls:  true,
	}, nil
}

func (p *Pool) first(idle map[string]*poolConn, addr string) Conn {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, ok := idle[addr]
	if !ok {
		return nil
	}

	// Remove the connection from the map.
	if conn.next != nil {
		idle[addr] = conn.next
		conn.next = nil
	} else {
		delete(idle, addr)
	}

	return conn
}

// CloseIdle closes all of the pool's idle connections.
func (p *Pool) CloseIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.idleTCP {
		closeConns(c)
	}
	for _, c := range p.idleTLS {
		closeConns(c)
	}

	p.idleTCP = nil
	p.idleTLS = nil
}

func (p *Pool) recycle(c *poolConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	c.idleSince = time.Now()

	// Idle connections are stored in per-host linked list, with the most
	// recently used connection at the head.
	if !c.tls {
		if p.idleTCP == nil {
			p.idleTCP = make(map[string]*poolConn)
		}
		c.next = p.idleTCP[c.addr]
		p.idleTCP[c.addr] = c
	} else {
		if p.idleTLS == nil {
			p.idleTLS = make(map[string]*poolConn)
		}
		c.next = p.idleTLS[c.addr]
		p.idleTLS[c.addr] = c
	}

	// If the garbage collector loop isn't running yet, start it.
	if !p.looping {
		p.looping = true
		go p.loop()
	}
}

func (p *Pool) loop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		cutoff := (<-ticker.C).Add(-p.timeout)
		p.mu.Lock()

		closeIdle(p.idleTCP, cutoff)
		closeIdle(p.idleTLS, cutoff)

		// If all idle connections have been closed, stop the goroutine.
		if len(p.idleTCP)+len(p.idleTLS) == 0 {
			p.looping = false
			p.mu.Unlock()
			return
		}

		p.mu.Unlock()
	}
}

func closeIdle(m map[string]*poolConn, cutoff time.Time) {
	for k, c := range m {
		if c.idleSince.Before(cutoff) {
			closeConns(c)
			delete(m, k)
		} else {
			for ; c.next != nil; c = c.next {
				if c.next.idleSince.Before(cutoff) {
					closeConns(c.next)
					c.next = nil
					break
				}
			}
		}
	}
}

func closeConns(c *poolConn) {
	for c != nil {
		c.Close()
		c, c.next = c.next, nil
	}
}

type poolConn struct {
	Conn

	// Connection pool.
	pool *Pool

	// Connection identifiers.
	addr string
	tls  bool

	// Next item in linked list.
	next *poolConn

	// When did the connection go idle?
	idleSince time.Time
}

func (c *poolConn) Recycle() error {
	c.pool.recycle(c)
	return nil
}

func (c *poolConn) Close() error {
	return c.Conn.Close()
}
