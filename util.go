package heat

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

func ReasonPhrase(status int) string {
	if s, ok := reasonPhrases[status]; ok {
		return s
	}
	return "Unknown"
}
