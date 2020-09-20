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

func TestNewUnknownFrameError(t *testing.T) {
	err := NewUnknownFrameError(0xFF)
	expectMsg := "unknown frame type: 0xFF"

	if err.Error() != expectMsg {
		t.Errorf("NewUnknownFrameError() got %s, want = %s", err.Error(), expectMsg)
	}

	if !errors.Is(err, UnknownFrameErr) {
		t.Errorf("NewUnknownFrameError() return error is NOT UnknownFrameErr")
	}
}
