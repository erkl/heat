package wire

type Request struct {
	// Request method.
	Method string

	// Request-URI.
	URI string

	// Header fields.
	Headers HeaderFields

	// Protocol scheme ("http" or "https").
	Scheme string

	// Remote address.
	RemoteAddr string
}
