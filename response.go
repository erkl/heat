package wire

type Response struct {
	// Status code.
	Status int

	// Reason phrase.
	Reason string

	// Header fields.
	Headers HeaderFields
}
