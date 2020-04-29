package huffman

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		in   []byte
		want []byte
	}{
		{
			in:   []byte("www.example.com"),
			want: []byte{0xf1, 0xe3, 0xc2, 0xe5, 0xf2, 0x3a, 0x6b, 0xa0, 0xab, 0x90, 0xf4, 0xff},
		},
		{
			in:   []byte("no-cache"),
			want: []byte{0xa8, 0xeb, 0x10, 0x64, 0x9c, 0xbf},
		},
		{
			in:   []byte("custom-key"),
			want: []byte{0x25, 0xa8, 0x49, 0xe9, 0x5b, 0xa9, 0x7d, 0x7f},
		},
		{
			in:   []byte("custom-value"),
			want: []byte{0x25, 0xa8, 0x49, 0xe9, 0x5b, 0xb8, 0xe8, 0xb4, 0xbf},
		},
	}

	for _, tt := range tests {
		got := Encode(tt.in)
		if bytes.Compare(got, tt.want) != 0 {
			t.Errorf("Encode('%X') got = %X, want = %X", tt.in, got, tt.want)
		}
	}
}
