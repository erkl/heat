package wire

import (
	"github.com/erkl/xo"
)

type Request struct {
	// Request method.
	Method string

	// Request-URI.
	URI string

	// Header fields.
	Headers HeaderFields

	// Protocol scheme ("http" or "https").
	Scheme string

	// Remote address.
	RemoteAddr string
}

func WriteRequestHeader(w xo.Writer, req *Request) error {
	buf, err := w.Reserve(len(req.Method) + len(req.URI) + 12)
	if err != nil {
		return err
	}

	n := copy(buf[0:], req.Method)
	n += copy(buf[n:], " ")
	n += copy(buf[n:], req.URI)
	n += copy(buf[n:], " HTTP/1.1\r\n")

	if err := w.Commit(n); err != nil {
		return err
	}

	return writeHeaderFields(w, req.Headers)
}
