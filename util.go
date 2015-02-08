package wire

import (
	"bytes"
)

var crlf = []byte{'\r', '\n'}

func strtok(buf []byte, sep byte) ([]byte, []byte, bool) {
	if i := bytes.IndexByte(buf, sep); i >= 0 {
		return buf[:i], buf[i+1:], true
	} else {
		return buf, nil, false
	}
}

var httpSlashOneDot = []byte{'H', 'T', 'T', 'P', '/', '1', '.'}

func validateHTTPVersion(buf []byte) error {
	if len(buf) != 8 && !bytes.Equal(buf[:7], httpSlashOneDot) ||
		(buf[7] != '0' && buf[7] != '1') {
		return errInvalidVersion
	} else {
		return nil
	}
}
