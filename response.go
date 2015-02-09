package wire

import (
	"io"

	"github.com/erkl/xo"
)

type Response struct {
	// Status code.
	Status int

	// Reason phrase.
	Reason string

	// Major and minor version numbers.
	Major int
	Minor int

	// Header fields.
	Headers HeaderFields

	// Message body.
	Body io.ReadCloser
}

func WriteResponseHeader(w xo.Writer, resp *Response) error {
	buf, err := w.Reserve(len(resp.Reason) + 10 + 20 + 20 + 20)
	if err != nil {
		return err
	}

	n := copy(buf[0:], "HTTP/")
	n += itoa(buf[n:], int64(resp.Major))
	n += copy(buf[n:], ".")
	n += itoa(buf[n:], int64(resp.Minor))
	n += copy(buf[n:], " ")
	n += itoa(buf[n:], int64(resp.Status))
	n += copy(buf[n:], " ")
	n += copy(buf[n:], resp.Reason)
	n += copy(buf[n:], "\r\n")

	if err := w.Commit(n); err != nil {
		return err
	}

	return writeHeaderFields(w, resp.Headers)
}

func ReadResponseHeader(r xo.Reader) (*Response, error) {
	var resp = new(Response)

	// Fetch the Status-Line.
	buf, err := xo.PeekTo(r, '\n', 0)
	if err != nil {
		return nil, err
	}

	version, rest := strtok(buf, ' ')
	if len(version) == 0 || rest == nil {
		return nil, ErrResponseHeader
	}

	resp.Major, resp.Minor, err = parseHTTPVersion(version)
	if err != nil {
		return nil, ErrResponseVersion
	}

	status, rest := strtok(rest, ' ')
	if len(status) == 0 || rest == nil {
		return nil, ErrResponseHeader
	}

	code, ok := atoi(status)
	if !ok || code > maxInt {
		return nil, ErrResponseHeader
	}

	// Trim trailing CRLF/LF.
	if len(rest) < 2 {
		return nil, ErrResponseHeader
	} else if rest[len(rest)-2] == '\r' {
		rest = rest[:len(rest)-2]
	} else {
		rest = rest[:len(rest)-1]
	}

	resp.Status = int(code)
	resp.Reason = string(rest)

	// Consume the Status-Line.
	if err := r.Consume(len(buf)); err != nil {
		return nil, err
	}

	// Read header fields.
	resp.Headers, err = readHeaderFields(r)
	if err != nil {
		if err == errMalformedHeader {
			err = ErrRequestHeader
		}
		return nil, err
	}

	return resp, nil
}
