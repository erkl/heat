package heat

import (
	"bytes"
	"strings"

	"github.com/erkl/xo"
)

var crlf = []byte{'\r', '\n'}

type Fields []Field

func (fs *Fields) Add(name, value string) {
	*fs = append(*fs, Field{name, value})
}

func (fs *Fields) Get(name string) (string, bool) {
	if i := fs.Index(name, 0); i >= 0 {
		return (*fs)[i].Value, true
	} else {
		return "", false
	}
}

func (fs *Fields) Has(name string) bool {
	return fs.Index(name, 0) >= 0
}

func (fs *Fields) Set(name, value string) bool {
	if i := fs.Index(name, 0); i >= 0 {
		(*fs)[i] = Field{name, value}
		fs.remove(name, i+1)
		return true
	}

	fs.Add(name, value)
	return false
}

func (fs *Fields) Remove(name string) bool {
	return fs.remove(name, 0)
}

func (fs *Fields) remove(name string, from int) bool {
	var w, r int
	var n = len(*fs)

	// Scan for the first matching field.
	for r = from; r < n; r++ {
		if (*fs)[r].Is(name) {
			goto rewrite
		}
	}

	return false

	// Overwrite matching fields in place.
rewrite:
	for w, r = r, r+1; r < n; r++ {
		if f := (*fs)[r]; !f.Is(name) {
			(*fs)[w] = f
			w++
		}
	}

	*fs = (*fs)[:w]
	return true
}

func (fs *Fields) Filter(fn func(f Field) bool) bool {
	var w, r int
	var n = len(*fs)

	// Scan for the first matching field.
	for ; r < n; r++ {
		if f := (*fs)[r]; !fn(f) {
			goto rewrite
		}
	}

	return false

	// Overwrite matching fields in place.
rewrite:
	for w, r = r, r+1; r < n; r++ {
		if f := (*fs)[r]; fn(f) {
			(*fs)[w] = f
			w++
		}
	}

	*fs = (*fs)[:w]
	return true
}

func (fs *Fields) Index(name string, from int) int {
	for i := from; i < len(*fs); i++ {
		if (*fs)[i].Is(name) {
			return i
		}
	}

	return -1
}

func (fs *Fields) Split(name string, sep byte) []string {
	var values []string
	var it = iter{*fs, name, sep, 0, 0}

	for {
		if value, ok := it.next(); !ok {
			break
		} else if len(value) > 0 {
			values = append(values, value)
		}
	}

	return values
}

func (fs *Fields) iter(name string, sep byte) iter {
	return iter{*fs, name, sep, 0, 0}
}

type Field struct {
	Name, Value string
}

func (f *Field) Is(name string) bool {
	return strcaseeq(f.Name, name)
}

type iter struct {
	fields Fields
	name   string
	sep    byte

	// Current position.
	row int
	col int
}

func (it *iter) next() (string, bool) {
	if it.col == 0 {
		it.row = it.fields.Index(it.name, it.row)
		if it.row < 0 {
			return "", false
		}
	}

	value := it.fields[it.row].Value
	if i := strings.IndexByte(value, it.sep); i >= 0 {
		it.col = i + 1
		return strtrim(value[it.col:i]), true
	}

	it.row++
	it.col = 0
	return strtrim(value[it.col:]), true
}

func writeHeader(w xo.Writer, fields Fields) error {
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

func readHeader(r xo.Reader) (Fields, error) {
	var fields Fields

	for {
		buf, err := xo.PeekTo(r, '\n', 0)
		if err != nil {
			return nil, err
		}

		if c := buf[0]; c == '\n' || (c == '\r' && len(buf) == 2) {
			if err := r.Consume(len(buf)); err != nil {
				return nil, err
			} else {
				return fields, nil
			}
		} else if c == ' ' || c == '\t' {
			// Because the loop below will consume all continuation lines,
			// taking this branch must mean that the first fields field has
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

		fields = append(fields, Field{
			Name:  stringify(name),
			Value: string(value),
		})

		// Consume the bytes we just parsed.
		if err := r.Consume(len(buf)); err != nil {
			return nil, err
		}
	}
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
