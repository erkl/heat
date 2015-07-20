package heat

import (
	"errors"
)

var (
	ErrRequestHeader  = errors.New("malformed request header")
	ErrRequestVersion = errors.New("invalid or unsupported protocol version in request header")
	ErrRequestNoHost  = errors.New("request missing Host header field")

	ErrResponseHeader  = errors.New("malformed response header")
	ErrResponseVersion = errors.New("invalid or unsupported protocol version in response header")

	ErrInvalidChunkedEncoding = errors.New("invalid chunked encoding")
	ErrInvalidContentLength   = errors.New("invalid content length")

	ErrInvalidBodySize = errors.New("invalid body size")
	ErrNilBody         = errors.New("unexpected nil body body")

	// Internal errors.
	errMalformedHeader = errors.New("malformed header")
	errInvalidVersion  = errors.New("invalid version")
)
