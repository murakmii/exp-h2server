package hpack

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestProcessHeaderBlockWithSampleInRFC(t *testing.T) {
	type in struct {
		ias          *IndexAddressSpace
		headerBlocks [][]byte
	}

	type want struct {
		headerLists []HeaderList
		errors      []error
	}

	tests := []struct {
		name string
		in   in
		want want
	}{
		{
			name: "https://tools.ietf.org/html/rfc7541#appendix-C.3",
			in: in{
				ias: NewIndexAddressSpace(4096),
				headerBlocks: [][]byte{
					{
						0x82, 0x86, 0x84, 0x41, 0x0f, 0x77, 0x77, 0x77,
						0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
						0x2e, 0x63, 0x6f, 0x6d,
					},
					{
						0x82, 0x86, 0x84, 0xbe, 0x58, 0x08, 0x6e, 0x6f,
						0x2d, 0x63, 0x61, 0x63, 0x68, 0x65,
					},
					{
						0x82, 0x87, 0x85, 0xbf, 0x40, 0x0a, 0x63, 0x75,
						0x73, 0x74, 0x6f, 0x6d, 0x2d, 0x6b, 0x65, 0x79,
						0x0c, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x2d,
						0x76, 0x61, 0x6c, 0x75, 0x65,
					},
				},
			},
			want: want{
				headerLists: []HeaderList{
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("http")},
						&HeaderField{name: []byte(":path"), value: []byte("/")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
					},
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("http")},
						&HeaderField{name: []byte(":path"), value: []byte("/")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
						&HeaderField{name: []byte("cache-control"), value: []byte("no-cache")},
					},
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("https")},
						&HeaderField{name: []byte(":path"), value: []byte("/index.html")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
						&HeaderField{name: []byte("custom-key"), value: []byte("custom-value")},
					},
				},
				errors: []error{nil, nil, nil},
			},
		},
		{
			name: "https://tools.ietf.org/html/rfc7541#appendix-C.4",
			in: in{
				ias: NewIndexAddressSpace(4096),
				headerBlocks: [][]byte{
					{
						0x82, 0x86, 0x84, 0x41, 0x8c, 0xf1, 0xe3, 0xc2,
						0xe5, 0xf2, 0x3a, 0x6b, 0xa0, 0xab, 0x90, 0xf4,
						0xff,
					},
					{
						0x82, 0x86, 0x84, 0xbe, 0x58, 0x86, 0xa8, 0xeb,
						0x10, 0x64, 0x9c, 0xbf,
					},
					{
						0x82, 0x87, 0x85, 0xbf, 0x40, 0x88, 0x25, 0xa8,
						0x49, 0xe9, 0x5b, 0xa9, 0x7d, 0x7f, 0x89, 0x25,
						0xa8, 0x49, 0xe9, 0x5b, 0xb8, 0xe8, 0xb4, 0xbf,
					},
				},
			},
			want: want{
				headerLists: []HeaderList{
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("http")},
						&HeaderField{name: []byte(":path"), value: []byte("/")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
					},
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("http")},
						&HeaderField{name: []byte(":path"), value: []byte("/")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
						&HeaderField{name: []byte("cache-control"), value: []byte("no-cache")},
					},
					{
						&HeaderField{name: []byte(":method"), value: []byte("GET")},
						&HeaderField{name: []byte(":scheme"), value: []byte("https")},
						&HeaderField{name: []byte(":path"), value: []byte("/index.html")},
						&HeaderField{name: []byte(":authority"), value: []byte("www.example.com")},
						&HeaderField{name: []byte("custom-key"), value: []byte("custom-value")},
					},
				},
				errors: []error{nil, nil, nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, input := range tt.in.headerBlocks {
				r := io.LimitReader(bytes.NewReader(input), int64(len(input)))
				got, err := ProcessHeaderBlock(tt.in.ias, r.(*io.LimitedReader))
				if !errors.Is(err, tt.want.errors[i]) {
					t.Errorf("[%d] ProcessHeaderBlock() return error = %v, want = %v", i, err, tt.want.errors[i])
					return
				}

				if !reflect.DeepEqual(got, tt.want.headerLists[i]) {
					t.Errorf("[%d] ProcessHeaderBlock() got = %v, want = %v", i, got, tt.want.headerLists[i])
				}
			}
		})
	}
}
