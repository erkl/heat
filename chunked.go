package wire

import (
	"io"

	"github.com/erkl/xo"
)

var hex = [16]byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'a', 'b', 'c', 'd', 'e', 'f',
}

var dehex = [256]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, // 0 1 2 3 4 5 6 7
	0x08, 0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 8 9 . . . . . .
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, // . A B C D E F .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, // . a b c d e f .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // . . . . . . . .
}

type chunkedWriter struct {
	w xo.Writer
	b [18]byte
}

func (cw *chunkedWriter) Write(chunk []byte) (int, error) {
	// Ignore empty chunks as they will be mistaken for the final chunk
	// in the stream.
	if len(chunk) == 0 {
		return 0, nil
	}

	// Write the chunk's size as hex at the end of the buffer.
	var i = 16

	for x := len(chunk); x > 0; x >>= 4 {
		i--
		cw.b[i] = hex[x&15]
	}

	if _, err := cw.w.Write(cw.b[i:]); err != nil {
		return 0, err
	}
	if _, err := cw.w.Write(chunk); err != nil {
		return 0, err
	}
	if _, err := cw.w.Write(cw.b[16:]); err != nil {
		return 0, err
	}

	return len(chunk), nil
}

func (cw *chunkedWriter) Close() error {
	cw.b[13] = '0'
	cw.b[14] = '\r'
	cw.b[15] = '\n'

	_, err := cw.w.Write(cw.b[13:])
	return err
}

type chunkedReader struct {
	r xo.Reader
	n int64
}

func (cr *chunkedReader) Read(buf []byte) (n int, err error) {
	// If we've fully consumed the previous chunk, read the size
	// of the next one.
	if cr.n == 0 {
		if err = cr.open(); err != nil {
			goto fail
		}

		// End the stream after an empty chunk.
		if cr.n == 0 {
			if err = cr.discardTrailers(); err != nil {
				return 0, err
			}
			return 0, io.EOF
		}
	}

	// Make sure we don't overshoot the end of this chunk.
	if int64(len(buf)) > cr.n {
		buf = buf[:int(cr.n)]
	}

	n, err = cr.r.Read(buf)
	if err != nil && n <= 0 {
		goto fail
	}

	// Consume trailing CRLF.
	if cr.n -= int64(n); cr.n == 0 {
		if err = cr.close(); err != nil {
			goto fail
		}
	}

	return n, nil

fail:
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return 0, err
}

func (cr *chunkedReader) open() error {
	buf, err := xo.PeekTo(cr.r, '\n', 0)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	// Quick check for invalid chunk size lines.
	if len(buf) < 3 || buf[len(buf)-2] != '\r' {
		return ErrInvalidChunkedEncoding
	}

	for i, c := range buf {
		// Decode hex characters.
		if x := dehex[c]; x <= 0xf {
			if cr.n > 0x07ffffffffffffff {
				return ErrInvalidChunkedEncoding
			}

			cr.n = cr.n<<4 | int64(x)
			continue
		}

		// Validate whatever's coming after the chunk size.
		if i > 0 {
			if c == '\r' && i == len(buf)-2 {
				break
			}

			// Chunk extensions are weird and seemingly unused, but RFC 2616
			// section 3.6.1 states:
			//
			//   All HTTP/1.1 applications MUST be able to receive and decode
			//   the "chunked" transfer-coding, and MUST ignore chunk-extension
			//   extensions they do not understand.
			//
			// ...so that means we'll just have to deal with them.
			if c == ';' {
				break
			}
		}

		// Any other case is an error.
		return ErrInvalidChunkedEncoding
	}

	return cr.r.Consume(len(buf))
}

func (cr *chunkedReader) close() error {
	crlf, err := cr.r.Peek(2)
	if err != nil {
		return err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return ErrInvalidChunkedEncoding
	}

	return cr.r.Consume(2)
}

func (cr *chunkedReader) discardTrailers() error {
	for {
		buf, err := xo.PeekTo(cr.r, '\n', 0)
		if err != nil {
			return err
		}

		// If the line is empty, we're done.
		done := (len(buf) < 2 || buf[0] == '\r')

		if err := cr.r.Consume(len(buf)); err != nil || done {
			return err
		}
	}
}
