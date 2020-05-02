package hpack

import (
	"fmt"
	"io"
	"math/bits"

	"github.com/murakmii/exp-h2server/h2server/hpack/huffman"
)

// decodePrefixedInt decodes integer representation with prefix
// See:https://tools.ietf.org/html/rfc7541#section-5.1
func decodePrefixedInt(r io.Reader, n int) (byte, uint64, error) {
	buf := make([]byte, 1)

	if _, err := r.Read(buf); err != nil {
		return 0, 0, err
	}

	prefix := buf[0] & (0xff << n)
	prefixedInt := buf[0] & (0xff >> (8 - n))

	if prefixedInt < (1<<n)-1 {
		return prefix, uint64(prefixedInt), nil
	}

	value := uint64(prefixedInt)
	shift := 0

	for {
		if shift >= 63 {
			return 0, 0, fmt.Errorf("%w: too long", ErrPrefixedInt)
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

// encodePrefixedInt encodes integer to integer representation with prefix
func encodePrefixedInt(n int, value uint64) []byte {
	if value < ((1 << n) - 1) {
		return []byte{byte(value)}
	}

	value -= (1 << n) - 1
	encoded := []byte{(1 << n) - 1}
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

// decodeStringLiteral decodes string literal
// See: https://tools.ietf.org/html/rfc7541#section-5.2
func decodeStringLiteral(r io.Reader, maxLength int) ([]byte, error) {
	encodedFlag, length, err := decodePrefixedInt(r, 7)
	if err != nil {
		return nil, err
	}

	if length > uint64(maxLength) {
		return nil, fmt.Errorf("%w: too long", ErrStringLiteral)
	}

	str := make([]byte, length)
	if _, err := r.Read(str); err != nil {
		return nil, err
	}

	if (encodedFlag >> 7) == 1 {
		str, err = huffman.Decode(str)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrHPACK, err.Error())
		}
	}

	return str, nil
}

// encodeStringLiteral encodes string to string literal
func encodeStringLiteral(str []byte, encodeHuffman bool) []byte {
	if encodeHuffman {
		str = huffman.Encode(str)
	}

	encoded := encodePrefixedInt(7, uint64(len(str)))
	if encodeHuffman {
		encoded[0] |= 1 << 7
	}

	return encoded
}
