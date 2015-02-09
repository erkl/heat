package wire

type Transport interface {
	RoundTrip(req *Request, cancel <-chan error) (*Response, error)
}
