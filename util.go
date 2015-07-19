package heat

import (
	"errors"
	"io"
)

var errReadAfterClose = errors.New("heat: read after close")

type MonitoredReader struct {
	R io.Reader
	e error
}

func (m *MonitoredReader) Read(buf []byte) (int, error) {
	if m.e != nil {
		return 0, m.e
	}

	n, err := m.R.Read(buf)
	if err != nil {
		// Persist errors.
		if m.e == nil {
			m.e = err
		} else {
			err = m.e
		}

		// If the call yielded any data, delay the error.
		if n > 0 {
			err = nil
		}
	}

	return n, err
}

func (m *MonitoredReader) Close() error {
	if m.e != nil {
		m.e = errReadAfterClose
	}

	return nil
}

func (m *MonitoredReader) LastError() error {
	if m.e == errReadAfterClose {
		return nil
	}

	return m.e
}

func Closing(major, minor int, fields Fields) bool {
	iter := fields.iter("Connection", ',')

	if major == 1 && minor == 0 {
		for {
			if value, ok := iter.next(); !ok {
				return true
			} else if strcaseeq(value, "keep-alive") {
				return false
			}
		}
	} else {
		for {
			if value, ok := iter.next(); !ok {
				return false
			} else if strcaseeq(value, "close") {
				return true
			}
		}
	}
}
