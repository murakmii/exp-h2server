package h2server

import (
	"errors"
	"fmt"
	"testing"
)

func TestUnwrapErrorCode(t *testing.T) {
	tests := []struct {
		in   error
		want ErrorCode
	}{
		{
			in:   NewH2Error(ProtocolError, "sample"),
			want: ProtocolError,
		},
		{
			in:   fmt.Errorf("error: %w", NewH2Error(ProtocolError, "sample")),
			want: ProtocolError,
		},
		{
			in:   errors.New("sample"),
			want: InternalError,
		},
	}

	for i, tt := range tests {
		got := UnwrapErrorCode(tt.in)
		if got != tt.want {
			t.Errorf("UnwrapErrorCode(#%d) got = '%s', want = '%s'", i+1, got.String(), tt.want.String())
		}
	}
}
