package heat

// Closing takes the HTTP version and header fields of a request or response,
// and returns true if they indicate that the sender expects the connection to
// be closed after this particular round trip.
func Closing(major, minor int, fields Fields) bool {
	var closing bool

	switch {
	case major == 1 && minor == 1:
		closing = false

		fields.Split("Connection", ',', func(s string) bool {
			if strcaseeq(s, "close") {
				closing = true
				return false
			}
			return true
		})

	case major == 1 && minor == 0:
		closing = true

		fields.Split("Connection", ',', func(s string) bool {
			if strcaseeq(s, "keep-alive") {
				closing = false
				return false
			}
			return true
		})

	default:
		// Default to non-keep-alive connections for unknown HTTP versions.
		closing = true
	}

	return closing
}

var reasonPhrases = map[int]string{
	100: "Continue",
	101: "Switching Protocols",

	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",

	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",

	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request URI Too Long",
	415: "Unsupported Media Type",
	416: "Requested Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",

	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	511: "Network Authentication Required",
}

// ReasonPhrase returns the standard reason phrase for a given status code,
// defaulting to "Unknown" for unsupported status codes.
func ReasonPhrase(status int) string {
	if s, ok := reasonPhrases[status]; ok {
		return s
	}
	return "Unknown"
}
