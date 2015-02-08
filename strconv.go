package wire

const (
	// Number of bits in an int.
	intSize = 32 << uint(^uint(0)>>63)

	// Maximum and minimum int values.
	maxInt = 1<<(intSize-1) - 1
	minInt = -1 << (intSize - 1)
)

func itoa(dst []byte, x int64) int {
	const dig01 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	const dig10 = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"

	var u, q, o uint64
	var n, i int

	// Negative numbers need special treatment.
	if x < 0 {
		u = uint64(-x)
		n = 1 + digits10(u)
		dst[0] = '-'
	} else {
		u = uint64(x)
		n = digits10(u)
	}

	i = n - 1

	// Principal write loop.
	for u >= 100 {
		q = u / 100
		o = u - q*100
		dst[i-0] = dig01[o]
		dst[i-1] = dig10[o]
		u = q
		i -= 2
	}

	// Write the last one or two digits.
	if u >= 10 {
		dst[i-0] = dig01[u]
		dst[i-1] = dig10[u]
	} else {
		dst[i] = dig01[u]
	}

	return n
}

func digits10(x uint64) int {
	// Optimize for the common case of a 3-digit x.
	if x < 1000 {
		if x < 100 {
			if x < 10 {
				return 1
			}
			return 2
		}
		return 3
	}

	// Large values of x are going to be incredibly rare,
	// so let's not bother with a binary search.
	switch {
	case x < 10000:
		return 4
	case x < 100000:
		return 5
	case x < 1000000:
		return 6
	case x < 10000000:
		return 7
	case x < 100000000:
		return 8
	}

	return 8 + digits10(x/100000000)
}

func atoi(buf []byte) (int64, bool) {
	const cutoff = (1<<64-1)/10 + 1

	var x int64

	// Empty input slices are obviously invalid.
	if len(buf) == 0 {
		return 0, false
	}

	for _, c := range buf {
		if !('0' <= c && c <= '9') {
			return 0, false
		}

		if x >= cutoff {
			return 0, false
		} else {
			x *= 10
		}

		if y := x + int64(c-'0'); y < x {
			return 0, false
		} else {
			x = y
		}
	}

	return x, true
}
