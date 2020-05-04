package hpack

import (
	"errors"
	"testing"
)

func TestHeaderField_IsPseudo(t *testing.T) {
	tests := []struct {
		headerField *HeaderField
		want        bool
	}{
		{headerField: &HeaderField{name: ":authority", value: ""}, want: true},
		{headerField: &HeaderField{name: ":scheme", value: ""}, want: true},
		{headerField: &HeaderField{name: ":method", value: ""}, want: true},
		{headerField: &HeaderField{name: ":path", value: ""}, want: true},
		{headerField: &HeaderField{name: ":status", value: ""}, want: true},
		{headerField: &HeaderField{name: ":foo", value: ""}, want: false},
		{headerField: &HeaderField{name: "age", value: ""}, want: false},
	}

	for _, tt := range tests {
		got := tt.headerField.IsPseudo()
		if got != tt.want {
			t.Errorf("IsPseudo() got = %v, want = %v", got, tt.want)
		}
	}
}

func TestHeaderField_Validate(t *testing.T) {
	tests := []struct {
		headerField *HeaderField
		want        error
	}{
		{headerField: &HeaderField{name: ":method", value: "GET"}, want: nil},
		{headerField: &HeaderField{name: "cache-control", value: "no-cache"}, want: nil},
		{headerField: &HeaderField{name: ":foo", value: "bar"}, want: ErrHeader},
		{headerField: &HeaderField{name: "Age", value: "300"}, want: ErrHeader},
		{headerField: &HeaderField{name: "x-custom-header", value: "あいうえお"}, want: ErrHeader},
	}

	for _, tt := range tests {
		got := tt.headerField.Validate()
		if !errors.Is(got, tt.want) {
			t.Errorf("Validate() got = %+v, want = %+v", got, tt.want)
		}
	}
}
