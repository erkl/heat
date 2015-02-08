package wire

import (
	"io"

	"github.com/erkl/xo"
)

type MessageSize int64

const (
	Chunked   MessageSize = iota - 1 // Terminated by an empty chunk and trailers.
	Multipart MessageSize = iota - 1 // Terminated by boundary.
	Unbounded MessageSize = iota - 1 // Terminated by closing the connection.

	invalid MessageSize = iota - 1
)

func RequestMessageSize(req *Request) (MessageSize, error) {
	n, err := genericMessageSize(req.Headers)
	if n == Unbounded {
		n = 0
	}
	return n, err
}

func ResponseMessageSize(resp *Response, method string) (MessageSize, error) {
	switch {
	case method == "HEAD":
		return 0, nil
	case 100 <= resp.Status && resp.Status <= 199:
		return 0, nil
	case resp.Status == 204:
		return 0, nil
	case resp.Status == 304:
		return 0, nil
	}

	return genericMessageSize(resp.Headers)
}

func genericMessageSize(headers HeaderFields) (MessageSize, error) {
	// TODO: Support for Content-Type: multipart/byteranges.

	if isChunkedTransfer(headers) {
		return Chunked, nil
	}

	if n, err := parseContentLength(headers); err != nil {
		return 0, err
	} else if n >= 0 {
		return MessageSize(n), nil
	}

	return Unbounded, nil
}

func isChunkedTransfer(headers HeaderFields) bool {
	it := headers.split("Transfer-Encoding")

	// According to RFC 2616, any Transfer-Encoding value other than
	// "identity" means the body is "chunked".
	for {
		if value, ok := it.next(); !ok {
			break
		} else if value != "" && !strcaseeq(value, "identity") {
			return true
		}
	}

	return false
}

func parseContentLength(headers HeaderFields) (int64, error) {
	var n int64 = -1
	var i int = -1

	for {
		// Find the next Content-Length field.
		if i = headers.index("Content-Length", i+1); i < 0 {
			break
		}

		value := strtrim(headers[i].Value)
		if value == "" {
			continue
		}

		// Convert the value to a 64-bit integer.
		var x int64

		for _, c := range value {
			if !('0' <= c && c <= '9') {
				return 0, ErrInvalidContentLength
			}

			if y := x*10 + int64(c-'0'); y/10 != x {
				return 0, ErrInvalidContentLength
			} else {
				x = y
			}
		}

		// Did we already find a conflicting Content-Length?
		if n >= 0 && x != n {
			return 0, ErrInvalidContentLength
		} else {
			n = x
		}
	}

	return n, nil
}

func WriteMessageBody(dst xo.Writer, src io.Reader, size MessageSize) error {
	// TODO: Add support for the Multipart size.

	if size == 0 {
		return nil
	} else if src == nil && size > invalid {
		return ErrNilMessageBody
	}

	switch {
	case size > 0:
		_, err := io.CopyN(dst, src, int64(size))
		return err
	case size == Chunked:
		cw := &chunkedWriter{dst, [18]byte{16: '\r', 17: '\n'}}
		if _, err := io.Copy(cw, src); err != nil {
			return err
		}
		return cw.Close()
	case size == Unbounded:
		_, err := io.Copy(dst, src)
		return err
	default:
		return ErrInvalidMessageSize
	}
}

func ReadMessageBody(src xo.Reader, size MessageSize) (io.Reader, error) {
	switch {
	case size == 0:
		return empty{}, nil
	case size > 0:
		return &io.LimitedReader{src, int64(size)}, nil
	case size == Chunked:
		return &chunkedReader{src, 0}, nil
	case size == Unbounded:
		return src, nil
	default:
		return nil, ErrInvalidMessageSize
	}
}

type empty struct{}

func (empty) Read(buf []byte) (int, error) {
	return 0, io.EOF
}
