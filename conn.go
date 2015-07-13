package wire

import (
	"net"
	"sync"

	"github.com/erkl/xo"
)

type Conn interface {
	xo.ReadWriter

	// RawConn returns the Conn's underlying net.Conn instance, if there
	// is one. Returns nil otherwise.
	RawConn() net.Conn

	// Close closes the connection. If keepAlive is true the last request-
	// response cycle terminated cleanly, and the connection may be reused.
	Close(keepAlive bool) error
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

func NewConn(conn net.Conn) Conn {
	bufs := bufpool.Get().([2][]byte)

	return &xConn{
		conn:   conn,
		Reader: xo.NewReader(conn, bufs[0]),
		Writer: xo.NewWriter(conn, bufs[1]),
		bufs:   bufs,
	}
}

func (c *xConn) RawConn() net.Conn {
	return c.conn
}

func (c *xConn) Close(keepAlive bool) error {
	bufpool.Put(c.bufs)
	return c.conn.Close()
}
