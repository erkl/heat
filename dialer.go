package wire

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/erkl/xo"
)

type Dialer interface {
	DialTCP(addr string) (Conn, error)
	DialTLS(addr string) (Conn, error)
}

type Conn interface {
	xo.ReadWriter

	// Close closes the connection. The recycle parameters should be true if
	// the last request-response cycle terminated cleanly, and the connection
	// can be reused (at the Dialer's discretion).
	Close(recycle bool) error
}

type xDialer struct {
	dialer  net.Dialer
	tlsConf *tls.Config
}

func NewDialer(timeout time.Duration, conf *tls.Config) Dialer {
	var dialer = new(xDialer)
	if timeout > 0 {
		dialer.dialer.Timeout = timeout
	}
	return dialer
}

func (d *xDialer) DialTCP(addr string) (Conn, error) {
	conn, err := d.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newConn(conn), nil
}

func (d *xDialer) DialTLS(addr string) (Conn, error) {
	conn, err := tls.DialWithDialer(&d.dialer, "tcp", addr, d.tlsConf)
	if err != nil {
		return nil, err
	}

	return newConn(conn), nil
}

// Pool of buffers used by xConn instances.
var bufpool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 8192)
		return [2][]byte{buf[:4096], buf[4096:]}
	},
}

type xConn struct {
	// The underlying net.Conn.
	conn net.Conn

	// Embedded buffered reader/writer.
	xo.Reader
	xo.Writer

	// Read/write buffers, returned to bufpool after the connection
	// is closed.
	bufs [2][]byte
}

func newConn(conn net.Conn) *xConn {
	bufs := bufpool.Get().([2][]byte)

	return &xConn{
		conn:   conn,
		Reader: xo.NewReader(conn, bufs[0]),
		Writer: xo.NewWriter(conn, bufs[1]),
		bufs:   bufs,
	}
}

func (c *xConn) Close(recycle bool) error {
	bufpool.Put(c.bufs)
	return c.conn.Close()
}
