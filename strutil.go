package wire

import (
	"bytes"
)

func strtok(buf []byte, sep byte) ([]byte, []byte, bool) {
	if i := bytes.IndexByte(buf, sep); i >= 0 {
		return buf[:i], buf[i+1:], true
	} else {
		return buf, nil, false
	}
}
