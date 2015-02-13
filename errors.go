package wire

import (
	"errors"
)

var (
	ErrRequestHeader  = errors.New("wire: malformed request header")
	ErrRequestVersion = errors.New("wire: invalid or unsupported protocol version in request header")
	ErrRequestNoHost  = errors.New("wire: request missing Host header field")

	ErrResponseHeader  = errors.New("wire: malformed response header")
	ErrResponseVersion = errors.New("wire: invalid or unsupported protocol version in response header")

	ErrInvalidChunkedEncoding = errors.New("wire: invalid chunked encoding")
	ErrInvalidContentLength   = errors.New("wire: invalid content length")

	ErrNilCancel          = errors.New("wire: round-trip cancelled with nil error")
	ErrUnsupportedScheme  = errors.New("wire: unsupported scheme")
	ErrInvalidMessageSize = errors.New("wire: invalid message size")
	ErrNilMessageBody     = errors.New("wire: unexpected nil message body")

	// Internal errors.
	errMalformedHeader = errors.New("wire: malformed header")
	errInvalidVersion  = errors.New("wire: invalid version")
	errReadAfterClose  = errors.New("wire: read after close")
	errClosedBeforeEOF = errors.New("wire: closed before EOF")
)
