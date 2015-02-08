package wire

import (
	"reflect"
	"testing"
)

var strcaseeqTests = []struct {
	a, b string
	out  bool
}{
	{"", "", true},
	{"x", "X", true},
	{"x", "y", false},
	{"_", "?", false},
}

func TestStrcaseeq(t *testing.T) {
	for _, test := range strcaseeqTests {
		out := strcaseeq(test.a, test.b)
		if out != test.out {
			t.Errorf("strcaseeq(%q, %q):", test.a, test.b)
			t.Errorf("  got  %v", out)
			t.Errorf("  want %v", test.out)
		}
	}
}

var strtrimTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{"x", "x"},
	{" x", "x"},
	{"x ", "x"},
	{" \tx y\t ", "x y"},
}

func TestStrtrim(t *testing.T) {
	for _, test := range strtrimTests {
		out := strtrim(test.in)
		if out != test.out {
			t.Errorf("strtrim(%q):", test.in)
			t.Errorf("  got  %q", out)
			t.Errorf("  want %q", test.out)
		}
	}
}

var strtokTests = []struct {
	in   []byte
	sep  byte
	tok  []byte
	rest []byte
}{
	{[]byte("x y z"), ' ', []byte("x"), []byte("y z")},
	{[]byte("y z"), ' ', []byte("y"), []byte("z")},
	{[]byte("z"), ' ', []byte("z"), nil},
	{[]byte{}, ' ', []byte{}, nil},
}

func TestStrtok(t *testing.T) {
	for _, test := range strtokTests {
		tok, rest := strtok(test.in, test.sep)
		if !reflect.DeepEqual(tok, test.tok) || !reflect.DeepEqual(rest, test.rest) {
			t.Errorf("strtok(%q, %q):", test.in, test.sep)
			t.Errorf("  got  %q, %q", tok, rest)
			t.Errorf("  want %q, %q", test.tok, test.rest)
		}
	}
}
