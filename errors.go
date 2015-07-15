package heat

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

	ErrInvalidBodySize = errors.New("wire: invalid body size")
	ErrNilBody         = errors.New("wire: unexpected nil body body")

	// Internal errors.
	errMalformedHeader = errors.New("wire: malformed header")
	errInvalidVersion  = errors.New("wire: invalid version")
)
