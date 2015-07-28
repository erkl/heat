package heat

import (
	"bytes"
	"strings"

	"github.com/erkl/xo"
)

var crlf = []byte{'\r', '\n'}

// Field represents a header field.
type Field struct {
	Name, Value string
}

// Is performs a case-insensitive match on the field's name.
func (f *Field) Is(name string) bool {
	return strcaseeq(f.Name, name)
}

// The Fields type represents a list of header fields. All operations on the
// type uses case-insensitive matching of field names.
type Fields []Field

// Get returns the value of the first field matching the specified name.
// The second return value indicates whether a match was found.
func (fs *Fields) Get(name string) (string, bool) {
	if i := fs.Index(name, 0); i >= 0 {
		return (*fs)[i].Value, true
	} else {
		return "", false
	}
}

// Has returns true if the list contains a particular field.
func (fs *Fields) Has(name string) bool {
	return fs.Index(name, 0) >= 0
}

// Index returns the index of the first field with the specified name,
// starting the search at the index from.
func (fs *Fields) Index(name string, from int) int {
	for i := from; i < len(*fs); i++ {
		if (*fs)[i].Is(name) {
			return i
		}
	}

	return -1
}

// Add appends a field to the list.
func (fs *Fields) Add(name, value string) {
	*fs = append(*fs, Field{name, value})
}

// Set adds a field to the list, returning true if any previous fields were
// removed in the process.
func (fs *Fields) Set(name, value string) bool {
	if i := fs.Index(name, 0); i >= 0 {
		(*fs)[i] = Field{name, value}
		fs.remove(name, i+1)
		return true
	}

	fs.Add(name, value)
	return false
}

// Remove drops all fields with the specified name.
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

// Filter removes all fields for which fn returns false.
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

// Split parses a particular field value as a list of elements split over any
// number of individual fields, using sep as the separator token. The provided
// callback function will be invoked with each element, with any leading or
// trailing whitespace removed.
//
// This method is useful when parsing fields like "Accept-Language" or
// "Transfer-Encoding".
func (fs *Fields) Split(name string, sep byte, fn func(s string) bool) {
	for _, f := range *fs {
		if !f.Is(name) {
			continue
		}

		v := f.Value

		// Split the value.
		if i := strings.IndexByte(v, sep); i >= 0 {
			if !fn(strtrim(v[:i])) {
				return
			}
			v = v[i+1:]
		}

		// Forward the remainder.
		if !fn(strtrim(v)) {
			return
		}
	}
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
