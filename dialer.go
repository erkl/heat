package wire

import (
	"crypto/tls"
	"net"
	"time"
)

type Dialer interface {
	DialTCP(addr string) (Conn, error)
	DialTLS(addr string) (Conn, error)
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

	return NewConn(conn), nil
}

func (d *xDialer) DialTLS(addr string) (Conn, error) {
	conn, err := tls.DialWithDialer(&d.dialer, "tcp", addr, d.tlsConf)
	if err != nil {
		return nil, err
	}

	return NewConn(conn), nil
}
