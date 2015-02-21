package wire

import (
	"net"
)

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Close closes the listener.
	Close() error

	// Addr returns the listener's network address.
	Addr() net.Addr
}

type xListener struct {
	net.Listener
}

func (l *xListener) Accept() (Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewConn(conn), nil
}

func ListenTCP(addr string) (Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &xListener{ln}, nil
}
