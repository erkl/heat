package wire

import (
	"net"
	"sync"

	"github.com/erkl/xo"
)

type Conn interface {
	xo.ReadWriter

	// Close closes the connection. The recycle parameters should be true if
	// the last request-response cycle terminated cleanly, and the connection
	// can be reused (at the Dialer's discretion).
	Close(recycle bool)
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

func (c *xConn) Close(recycle bool) {
	bufpool.Put(c.bufs)
	c.conn.Close()
}
