package wire

type HeaderFields []HeaderField

type HeaderField struct {
	Name, Value string
}
