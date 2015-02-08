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
