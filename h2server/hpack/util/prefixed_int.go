package util

import (
	"errors"
	"io"
	"math/bits"
)

var ErrBigPrefixedInt = errors.New("big prefixed int(over 63 bits)")

// DecodePrefixedInt decodes integer representation with prefix
// See:https://tools.ietf.org/html/rfc7541#section-5.1
func DecodePrefixedInt(r io.Reader, prefixBits int) (byte, uint64, error) {
	buf := make([]byte, 1)

	if _, err := r.Read(buf); err != nil {
		return 0, 0, err
	}

	prefix := buf[0] & (0xff << prefixBits)
	prefixedInt := buf[0] & (0xff >> (8 - prefixBits))

	if prefixedInt < (1<<prefixBits)-1 {
		return prefix, uint64(prefixedInt), nil
	}

	value := uint64(prefixedInt)
	shift := 0

	for {
		if shift >= 63 {
			return 0, 0, ErrBigPrefixedInt
		}

		if _, err := r.Read(buf); err != nil {
			return 0, 0, err
		}

		value += uint64(buf[0]&0x7f) << shift
		if (buf[0] >> 7) == 0 {
			break
		}

		shift += 7
	}

	return prefix, value, nil
}

// EncodePrefixedInt encodes integer to integer representation with prefix
func EncodePrefixedInt(prefixBits int, value uint64) []byte {
	if value < ((1 << prefixBits) - 1) {
		return []byte{byte(value)}
	}

	value -= (1 << prefixBits) - 1
	encoded := []byte{(1 << prefixBits) - 1}
	remain := 64 - bits.LeadingZeros64(value)

	for remain > 0 {
		msb := byte(1)
		if remain <= 7 {
			msb = 0
		}

		encoded = append(encoded, (msb<<7)|byte(value&0x7f))
		value >>= 7
		remain -= 7
	}

	return encoded
}
