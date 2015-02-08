package wire

import (
	"bytes"

	"github.com/erkl/xo"
)

var crlf = []byte{'\r', '\n'}

type HeaderFields []HeaderField

type HeaderField struct {
	Name, Value string
}

func (h *HeaderFields) Add(name, value string) {
	*h = append(*h, HeaderField{name, value})
}

func (h *HeaderFields) Get(name string) (string, bool) {
	if i := h.index(name, 0); i >= 0 {
		return (*h)[i].Value, true
	} else {
		return "", false
	}
}

func (h *HeaderFields) Has(name string) bool {
	return h.index(name, 0) >= 0
}

func (h *HeaderFields) Set(name, value string) bool {
	if i := h.index(name, 0); i >= 0 {
		(*h)[i] = HeaderField{name, value}
		h.remove(name, i+1)
		return true
	}

	h.Add(name, value)
	return false
}

func (h *HeaderFields) Remove(name string) bool {
	return h.remove(name, 0)
}

func (h *HeaderFields) remove(name string, i int) bool {
	if i = h.index(name, i); i < 0 {
		return false
	}

	var e = i
	for i++; i < len(*h); i++ {
		if strcaseeq((*h)[i].Name, name) {
			(*h)[e] = (*h)[i]
			e++
		}
	}

	*h = (*h)[:e]
	return true
}

func (h *HeaderFields) index(name string, from int) int {
	for i := from; i < len(*h); i++ {
		if strcaseeq((*h)[i].Name, name) {
			return i
		}
	}
	return -1
}

func (h *HeaderFields) split(name string) *fieldSplitter {
	return &fieldSplitter{*h, name, 0, 0}
}

type fieldSplitter struct {
	fields HeaderFields
	name   string
	line   int
	offset int
}

func (s *fieldSplitter) next() (string, bool) {
	if s.offset == 0 {
		s.line = s.fields.index(s.name, s.line)
		if s.line < 0 {
			return "", false
		}
	}

	raw := s.fields[s.line].Value

	for i := s.offset; ; i++ {
		if i == len(raw) {
			s.line++
			s.offset = 0
			return strtrim(raw[s.offset:]), true
		}
		if raw[i] == ',' {
			s.offset = i + 1
			return strtrim(raw[s.offset:i]), true
		}
	}
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
