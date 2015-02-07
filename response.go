package wire

import (
	"github.com/erkl/xo"
)

type Response struct {
	// Status code.
	Status int

	// Reason phrase.
	Reason string

	// Header fields.
	Headers HeaderFields
}

func WriteResponseHeader(w xo.Writer, resp *Response) error {
	buf, err := w.Reserve(len(resp.Reason) + 12 + 20)
	if err != nil {
		return err
	}

	n := copy(buf[0:], "HTTP/1.1 ")
	n += itoa(buf[n:], int64(resp.Status))
	n += copy(buf[n:], " ")
	n += copy(buf[n:], resp.Reason)
	n += copy(buf[n:], "\r\n")

	if err := w.Commit(n); err != nil {
		return err
	}

	return writeHeaderFields(w, resp.Headers)
}
