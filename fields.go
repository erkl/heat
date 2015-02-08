package wire

import (
	"bytes"

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

func readHeaderFields(r xo.Reader) (HeaderFields, error) {
	var fields HeaderFields

	for {
		buf, err := xo.PeekTo(r, '\n', 0)
		if err != nil {
			return nil, err
		}

		if c := buf[0]; c == '\n' || (c == '\r' && len(buf) == 2) {
			return nil, r.Consume(len(buf))
		} else if c == ' ' || c == '\t' {
			// Because the loop below will consume all continuation lines,
			// taking this branch must mean that the first header field has
			// leading whitespace, which is illegal.
			return nil, errMalformedHeader
		}

		colon := bytes.IndexByte(buf, ':')
		if colon == -1 {
			return nil, errMalformedHeader
		}

		// Lines beginning with horizontal whitespace are continuations of
		// the field value on the previous line, meaning we have to read all
		// of them before we have a full field value.
		for off := len(buf); ; off = len(buf) {
			peek, err := xo.PeekTo(r, '\n', off)
			if err != nil {
				return nil, err
			}

			if c := peek[off]; c == ' ' || c == '\t' {
				buf = peek
			} else {
				break
			}
		}

		// Trim the field's name and value. The shrinkValue call will modify
		// buf in place, which is referencing the xo.Reader's internal storage.
		// This isn't ideal, but it will only matter if the Consume call fails,
		// which is impossible for correct xo.Readers.
		name := shrinkName(buf[:colon])
		if len(name) == 0 {
			return nil, errMalformedHeader
		}

		value := shrinkValue(buf[colon+1:])

		fields = append(fields, HeaderField{
			Name:  string(name),
			Value: string(value),
		})

		// Consume the bytes we just parsed.
		if err := r.Consume(len(buf)); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

func shrinkName(buf []byte) []byte {
	for len(buf) > 0 && buf[len(buf)-1] == ' ' {
		buf = buf[:len(buf)-1]
	}
	return buf
}

func shrinkValue(buf []byte) []byte {
	var c byte
	var r, w int
	var m = -1

	// Trim leading whitespace.
	for len(buf) > 0 && buf[0] == ' ' {
		buf = buf[1:]
	}

	for r < len(buf) {
		if c = buf[r]; c == '\r' || c == '\n' {
			// Replace all trailing whitespace on this line with a single
			// space character.
			if m != -1 {
				buf[m] = ' '
				w = m + 1
			}

			// Fast-forward past all upcoming whitespace.
			for r++; r < len(buf); r++ {
				if c = buf[r]; c != ' ' && c != '\t' && c != '\r' && c != '\n' {
					break
				}
			}

			continue
		}

		buf[w] = buf[r]
		w, r = w+1, r+1

		// Remember the position of the last non-whitespace character.
		if c != ' ' && c != '\t' {
			m = w
		}
	}

	if m < 0 {
		return nil
	} else {
		return buf[:m]
	}
}
