package wire

import (
	"testing"
)

var itoaTests = []struct {
	in  int64
	out string
}{
	{0, "0"},
	{1, "1"},
	{9223372036854775807, "9223372036854775807"},
	{-1, "-1"},
	{-9223372036854775808, "-9223372036854775808"},
}

func TestItoa(t *testing.T) {
	for _, test := range itoaTests {
		buf := make([]byte, 64)
		out := string(buf[:itoa(buf, test.in)])
		if out != test.out {
			t.Errorf("itoa(%d):", test.in)
			t.Errorf("  got  %q", out)
			t.Errorf("  want %q", test.out)
		}
	}
}

var atoiTests = []struct {
	in  string
	out int64
	ok  bool
}{
	{"", 0, false},
	{"0", 0, true},
	{"1", 1, true},
	{"12345678", 12345678, true},
	{"foo bar", 0, false},
	{"-", 0, false},
	{"-0", 0, false},
	{"-1", 0, false},
	{"9223372036854775807", 9223372036854775807, true},
	{"9223372036854775808", 0, false},
}

func TestAtoi(t *testing.T) {
	for _, test := range atoiTests {
		out, ok := atoi([]byte(test.in))
		if out != test.out || ok != test.ok {
			t.Errorf("atoi(%q):", test.in)
			t.Errorf("  got  %d, %v", out, ok)
			t.Errorf("  want %d, %v", test.out, test.ok)
		}
	}
}
