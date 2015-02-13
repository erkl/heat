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

	// Major and minor version numbers.
	Major int
	Minor int

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
		Major:  1,
		Minor:  1,
		Headers: HeaderFields{
			{"Host", u.Host},
		},
		Scheme:     u.Scheme,
		RemoteAddr: u.Host,
	}
}

func (r *Request) ParseURL() (*url.URL, error) {
	host, ok := r.Headers.Get("Host")
	if !ok {
		return nil, ErrRequestNoHost
	}

	u, err := url.ParseRequestURI(r.URI)
	if err != nil {
		return nil, err
	}

	u.Scheme = r.Scheme
	u.Host = host

	return u, nil
}

func WriteRequestHeader(w xo.Writer, req *Request) error {
	buf, err := w.Reserve(len(req.Method) + len(req.URI) + 10 + 20 + 20)
	if err != nil {
		return err
	}

	n := copy(buf[0:], req.Method)
	n += copy(buf[n:], " ")
	n += copy(buf[n:], req.URI)
	n += copy(buf[n:], " HTTP/")
	n += itoa(buf[n:], int64(req.Major))
	n += copy(buf[n:], ".")
	n += itoa(buf[n:], int64(req.Minor))
	n += copy(buf[n:], "\r\n")

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

	req.Major, req.Minor, err = parseHTTPVersion(rest)
	if err != nil {
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

func parseHTTPVersion(buf []byte) (int, int, error) {
	if len(buf) == 8 && bytes.Equal(buf[:7], httpSlashOneDot) &&
		(buf[7] == '0' || buf[7] == '1') {
		return 1, int(buf[7] - '0'), nil
	} else {
		return 0, 0, errInvalidVersion
	}
}
