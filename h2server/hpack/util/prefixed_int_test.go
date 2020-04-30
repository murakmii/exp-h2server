package util

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadPrefixedInt(t *testing.T) {
	type in struct {
		r          io.Reader
		prefixBits int
	}

	type want struct {
		prefix byte
		value  uint64
		err    error
	}

	tests := []struct {
		in   in
		want want
	}{
		{
			in:   in{r: bytes.NewReader([]byte{0xbf, 0x9a, 0x0a}), prefixBits: 5},
			want: want{prefix: 0xa0, value: 1337, err: nil},
		},
		{
			in:   in{r: bytes.NewReader([]byte{0x2a}), prefixBits: 5},
			want: want{prefix: 0x20, value: 10, err: nil},
		},
		{
			in:   in{r: bytes.NewReader([]byte{0x2a}), prefixBits: 8},
			want: want{prefix: 0x00, value: 42, err: nil},
		},
		{
			in:   in{r: bytes.NewReader([]byte{0x82}), prefixBits: 7},
			want: want{prefix: 0x80, value: 2, err: nil},
		},
	}

	for _, tt := range tests {
		prefix, value, err := DecodePrefixedInt(tt.in.r, tt.in.prefixBits)
		if prefix != tt.want.prefix || value != tt.want.value || !errors.Is(err, tt.want.err) {
			t.Errorf("DecodePrefixedInt() got = {%d %d %v}, want = %v", prefix, value, err, tt.want)
		}
	}
}

func TestEncodePrefixedInt(t *testing.T) {
	type in struct {
		prefixBits int
		value      uint64
	}

	tests := []struct {
		in   in
		want []byte
	}{
		{
			in:   in{prefixBits: 5, value: 1337},
			want: []byte{0x1f, 0x9a, 0x0a},
		},
	}

	for _, tt := range tests {
		encoded := EncodePrefixedInt(tt.in.prefixBits, tt.in.value)
		if bytes.Compare(encoded, tt.want) != 0 {
			t.Errorf("EncodePrefixedInt() got = %X, want = %X", encoded, tt.want)
		}
	}
}
