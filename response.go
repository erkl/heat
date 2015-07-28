package heat

import (
	"io"

	"github.com/erkl/xo"
)

// The Response struct represents an HTTP request.
type Response struct {
	// Status code and reason phrase.
	Status int
	Reason string

	// HTTP version, represented as major and minor version numbers.
	// Only 1.0 and 1.1 are officially supported.
	Major int
	Minor int

	// Associated header fields.
	Fields Fields

	// Optional message body.
	Body io.ReadCloser
}

// The NewResponse function constructs a minimal Response instance given
// a status code and reason phrase.
func NewResponse(status int, reason string) *Response {
	return &Response{
		Status: status,
		Reason: reason,
		Major:  1,
		Minor:  1,
	}
}

// WriteResponseHeader writes an HTTP response header to w.
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

	return writeHeader(w, resp.Fields)
}

// ReadResponseHeader reads an HTTP response header from r.
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
	resp.Reason = stringify(rest)

	// Consume the Status-Line.
	if err := r.Consume(len(buf)); err != nil {
		return nil, err
	}

	// Read header fields.
	resp.Fields, err = readHeader(r)
	if err != nil {
		if err == errMalformedHeader {
			err = ErrResponseHeader
		}
		return nil, err
	}

	return resp, nil
}
