package wire

import (
	"bytes"
	"io"
	"net/url"

	"github.com/erkl/xo"
)

type Request struct {
	// Request method.
	Method string

	// Request-URI.
	URI string

	// Header fields.
	Headers HeaderFields

	// Message body.
	Body io.Reader

	// Protocol scheme ("http" or "https").
	Scheme string

	// Remote address.
	RemoteAddr string
}

func NewRequest(method string, u *url.URL) *Request {
	return &Request{
		Method: method,
		URI:    u.RequestURI(),
		Headers: HeaderFields{
			{"Host", u.Host},
		},
		Scheme:     u.Scheme,
		RemoteAddr: u.Host,
	}
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

func ReadRequestHeader(r xo.Reader) (*Request, error) {
	var req = new(Request)

	// Fetch the whole Request-Line.
	buf, err := xo.PeekTo(r, '\n', 0)
	if err != nil {
		return nil, err
	}

	method, rest := strtok(buf, ' ')
	if len(method) == 0 || rest == nil {
		return nil, ErrRequestHeader
	}

	uri, rest := strtok(rest, ' ')
	if len(uri) == 0 || rest == nil {
		return nil, ErrRequestHeader
	}

	// Trim trailing CRLF/LF.
	if len(rest) < 2 {
		return nil, ErrRequestHeader
	} else if rest[len(rest)-2] == '\r' {
		rest = rest[:len(rest)-2]
	} else {
		rest = rest[:len(rest)-1]
	}

	if err := validateHTTPVersion(rest); err != nil {
		return nil, ErrRequestVersion
	}

	req.Method = string(method)
	req.URI = string(uri)

	// Consume the Request-Line.
	if err := r.Consume(len(buf)); err != nil {
		return nil, err
	}

	// Read header fields.
	req.Headers, err = readHeaderFields(r)
	if err != nil {
		if err == errMalformedHeader {
			err = ErrRequestHeader
		}
		return nil, err
	}

	return req, nil
}

var httpSlashOneDot = []byte{'H', 'T', 'T', 'P', '/', '1', '.'}

func validateHTTPVersion(buf []byte) error {
	if len(buf) != 8 && !bytes.Equal(buf[:7], httpSlashOneDot) ||
		(buf[7] != '0' && buf[7] != '1') {
		return errInvalidVersion
	} else {
		return nil
	}
}
