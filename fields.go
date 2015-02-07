package wire

import (
	"github.com/erkl/xo"
)

type HeaderFields []HeaderField

type HeaderField struct {
	Name, Value string
}

func writeHeaderFields(w xo.Writer, fields HeaderFields) error {
	for _, f := range fields {
		buf, err := w.Reserve(len(f.Name) + len(f.Value) + 4)
		if err != nil {
			return err
		}

		n := copy(buf[0:], f.Name)
		n += copy(buf[n:], ": ")
		n += copy(buf[n:], f.Value)
		n += copy(buf[n:], "\r\n")

		if err := w.Commit(n); err != nil {
			return err
		}
	}

	_, err := w.Write(crlf)
	return err
}
