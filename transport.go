package wire

import (
	"io"
	"net"
	"sync"
)

type Transport interface {
	RoundTrip(req *Request, cancel <-chan error) (*Response, error)
}

type xTransport struct {
	dialer Dialer
}

func NewTransport(d Dialer) Transport {
	return &xTransport{d}
}

func (t *xTransport) RoundTrip(req *Request, cancel <-chan error) (*Response, error) {
	// Grab a connection.
	conn, err := t.dial(req.Scheme, req.RemoteAddr)
	if err != nil {
		return nil, err
	}

	// Recycle the connection if the request has already been cancelled.
	select {
	case err := <-cancel:
		if err == nil {
			err = ErrNilCancel
		}
		conn.Close(true)
		return nil, err

	default:
	}

	// Channels for synchronization.
	wait := make(chan error, 1)
	werr := make(chan error, 1)
	rerr := make(chan error, 1)

	// Perform all I/O operations in a separate goroutine.
	var resp *Response
	var rsize MessageSize

	go func() {
		wsize, err := RequestMessageSize(req)
		if err != nil {
			goto done
		}

		// Write the request header.
		if err = WriteRequestHeader(conn, req); err != nil {
			goto done
		}
		if err = conn.Flush(); err != nil {
			goto done
		}

		// Send the request message body.
		if wsize == 0 {
			werr <- nil
		} else {
			go func() {
				err := WriteMessageBody(conn, req.Body, wsize)
				if err != nil {
					werr <- err
				} else {
					werr <- conn.Flush()
				}
			}()
		}

		// Read the response headers.
		resp, err = ReadResponseHeader(conn)
		if err != nil {
			goto done
		}
		rsize, err = ResponseMessageSize(resp, req.Method)

		// Signal that the headers have been exchanged.
	done:
		wait <- err
		if err != nil {
			return
		}

		// "Reuse" this goroutine for closing the connection when we're
		// done with it.
		if <-werr != nil || <-rerr != io.EOF || rsize == Unbounded {
			conn.Close(true)
		} else {
			conn.Close(false)
		}
	}()

	// Wait for the request headers to be sent and the response headers to be
	// received, or for the round-trip to be cancelled.
	select {
	case err := <-cancel:
		if err == nil {
			err = ErrNilCancel
		}
		conn.Close(false)
		return nil, err

	case err := <-wait:
		if err != nil {
			conn.Close(true)
			return nil, err
		}
	}

	// Avoid a bit of work if the response body is empty.
	if rsize == 0 {
		rerr <- io.EOF
		resp.Body = empty{}
		return resp, nil
	}

	r, err := ReadMessageBody(conn, rsize)
	if err != nil {
		panic("unreachable")
	}

	resp.Body = &bodyReader{r: r, ch: rerr}
	return resp, nil
}

func (t *xTransport) dial(scheme, addr string) (Conn, error) {
	switch scheme {
	case "http":
		if !hasPort(addr) {
			addr = net.JoinHostPort(addr, "80")
		}
		return t.dialer.DialTCP(addr)

	case "https":
		if !hasPort(addr) {
			addr = net.JoinHostPort(addr, "443")
		}
		return t.dialer.DialTLS(addr)
	}

	return nil, ErrUnsupportedScheme
}

func hasPort(addr string) bool {
	if len(addr) == 0 {
		return false
	}

	var colons int
	var rbrack bool

	for i, c := range addr {
		if c == ':' {
			colons++
			rbrack = addr[i-1] == ']'
		}
	}

	switch colons {
	case 0:
		return false
	case 1:
		return true
	default:
		return addr[0] == '[' && rbrack
	}
}

type bodyReader struct {
	r io.Reader

	// Persisted error.
	err error

	// Channel for reporting the first encountered non-nil error,
	// including io.EOF.
	ch chan<- error

	// Guards the r and err fields.
	rdmu sync.Mutex
	ermu sync.Mutex
}

func (br *bodyReader) Read(buf []byte) (int, error) {
	br.rdmu.Lock()
	defer br.rdmu.Unlock()

	// Has a previous read already failed?
	br.ermu.Lock()
	err := br.err
	br.ermu.Unlock()
	if err != nil {
		return 0, err
	}

	// Perform the actual read.
	n, err := br.r.Read(buf)
	if err != nil {
		br.ermu.Lock()
		if br.err != nil {
			err = br.err
		} else {
			br.err = err
			br.ch <- err
		}
		br.ermu.Unlock()

		// Delay reporting the error if n > 0.
		if n > 0 {
			err = nil
		}
	}

	return n, err
}

func (br *bodyReader) Close() error {
	br.ermu.Lock()
	defer br.ermu.Unlock()

	if br.err == nil {
		br.err = errReadAfterClose
		br.ch <- errClosedBeforeEOF
	}

	return nil
}
