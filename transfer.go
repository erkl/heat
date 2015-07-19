package heat

import (
	"io"

	"github.com/erkl/xo"
)

type BodySize int64

const (
	Chunked   = -1 // Terminated by an empty chunk and trailers.
	Multipart = -2 // Terminated by boundary.
	Unbounded = -3 // Terminated by closing the connection.

	invalid = -4
)

func RequestBodySize(req *Request) (BodySize, error) {
	n, err := genericBodySize(req.Fields)
	if n == Unbounded {
		n = 0
	}
	return n, err
}

func ResponseBodySize(resp *Response, method string) (BodySize, error) {
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

	return genericBodySize(resp.Fields)
}

func genericBodySize(fields Fields) (BodySize, error) {
	// TODO: Support for Content-Type: multipart/byteranges.

	if isChunkedTransfer(fields) {
		return Chunked, nil
	}

	if n, err := parseContentLength(fields); err != nil {
		return 0, err
	} else if n >= 0 {
		return BodySize(n), nil
	}

	return Unbounded, nil
}

func isChunkedTransfer(fields Fields) bool {
	iter := fields.iter("Transfer-Encoding", ',')

	// According to RFC 2616, any Transfer-Encoding value other than
	// "identity" means the body is "chunked".
	for {
		if value, ok := iter.next(); !ok {
			break
		} else if value != "" && !strcaseeq(value, "identity") {
			return true
		}
	}

	return false
}

func parseContentLength(fields Fields) (int64, error) {
	var n int64 = -1
	var i int

	for {
		// Find the next Content-Length field.
		if i = fields.Index("Content-Length", i+1); i < 0 {
			break
		}

		value := strtrim(fields[i].Value)
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

func WriteBody(dst xo.Writer, src io.Reader, size BodySize) error {
	// TODO: Add support for the Multipart size.

	if size == 0 {
		return nil
	} else if src == nil && size > invalid {
		return ErrNilBody
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
		return ErrInvalidBodySize
	}
}

func ReadBody(src xo.Reader, size BodySize) (io.Reader, error) {
	switch {
	case size == 0:
		return nil, nil
	case size > 0:
		return &fixedReader{src, int64(size)}, nil
	case size == Chunked:
		return &chunkedReader{src, 0}, nil
	case size == Unbounded:
		return src, nil
	default:
		return nil, ErrInvalidBodySize
	}
}

type fixedReader struct {
	r io.Reader
	n int64
}

func (fr *fixedReader) Read(buf []byte) (int, error) {
	if fr.n <= 0 {
		return 0, io.EOF
	}

	if int64(len(buf)) > fr.n {
		buf = buf[:fr.n]
	}

	n, err := fr.r.Read(buf)
	if err != nil {
		if n > 0 {
			err = nil
		} else if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}

	fr.n -= int64(n)
	return n, err
}

func Closing(major, minor int, fields Fields) bool {
	iter := fields.iter("Connection", ',')

	if major == 1 && minor == 0 {
		for {
			if value, ok := iter.next(); !ok {
				return true
			} else if strcaseeq(value, "keep-alive") {
				return false
			}
		}
	} else {
		for {
			if value, ok := iter.next(); !ok {
				return false
			} else if strcaseeq(value, "close") {
				return true
			}
		}
	}
}
