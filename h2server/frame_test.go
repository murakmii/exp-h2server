package h2server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	type want struct {
		typ      FrameType
		flags    uint8
		streamID uint32
		payload  []byte
		err      error
	}

	tests := []struct {
		in   io.Reader
		want want
	}{
		{
			in: bytes.NewBuffer([]byte{
				0x00, 0x00, 0x05, // Length
				0x04, 0x0f, // FrameType, Flags
				0x00, 0x00, 0xff, 0xff, // Stream Identifier
				0x48, 0x65, 0x6c, 0x6c, 0x6f, // Payload
			}),
			want: want{
				typ:      SettingsFrameType,
				flags:    0x0f,
				streamID: 0xffff,
				payload:  []byte("Hello"),
				err:      nil,
			},
		},
	}

	for _, tt := range tests {
		got, err := Read(tt.in)
		if !errors.Is(err, tt.want.err) {
			t.Errorf("Read() got error = %v, want = %v", err, tt.want.err)
			continue
		}
		if err != nil {
			continue
		}

		if got.Type() != tt.want.typ ||
			got.Flags() != tt.want.flags ||
			got.StreamID() != tt.want.streamID ||
			bytes.Compare(got.Payload(), tt.want.payload) != 0 {
			t.Errorf("Read() got = %+v, want = %+v", got, tt.want)
		}
	}
}

func TestOutgoingFrame_WriteTo(t *testing.T) {
	type want struct {
		n   int64
		err error
	}

	tests := []struct {
		original []byte
		want     want
	}{
		{
			original: []byte{
				0x00, 0x00, 0x05, // Length
				0x04, 0x0f, // FrameType, Flags
				0x00, 0x00, 0xff, 0xff, // Stream Identifier
				0x48, 0x65, 0x6c, 0x6c, 0x6f, // Payload
			},
			want: want{
				n:   14,
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		f, err := Read(bytes.NewBuffer(tt.original))
		if err != nil {
			t.FailNow()
		}

		wrote := bytes.NewBuffer(nil)
		got, err := f.WriteTo(wrote)
		if !errors.Is(err, tt.want.err) {
			t.Errorf("WriteTo() got error = %v, want = %v", err, tt.want.err)
			continue
		}

		if got != int64(len(tt.original)) || bytes.Compare(wrote.Bytes(), tt.original) != 0 {
			t.Errorf("WriteTo() got = %v, want = %v", wrote.Bytes(), tt.original)
		}
	}
}

func TestSettingsFrame_IsACK(t *testing.T) {
	tests := []struct {
		settingsFrame *SettingsFrame
		want          bool
	}{
		{
			settingsFrame: &SettingsFrame{frame: &frame{typ: SettingsFrameType, flags: 0x00}},
			want:          false,
		},
		{
			settingsFrame: &SettingsFrame{frame: &frame{typ: SettingsFrameType, flags: 0x01}},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("flags=0x%X", tt.settingsFrame.flags), func(t *testing.T) {
			got := tt.settingsFrame.IsACK()
			if got != tt.want {
				t.Errorf("IsACK() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestSettingsFrame_Verify(t *testing.T) {
	tests := []struct {
		name          string
		settingsFrame *SettingsFrame
		want          ErrorCode
	}{
		{
			name:          "valid",
			settingsFrame: &SettingsFrame{frame: &frame{typ: SettingsFrameType}},
			want:          NoError,
		},
		{
			name:          "invalid_stream_ID",
			settingsFrame: &SettingsFrame{frame: &frame{typ: SettingsFrameType, streamID: 123}},
			want:          ProtocolError,
		},
		{
			name: "ack_with_payload",
			settingsFrame: &SettingsFrame{
				frame: &frame{
					typ:     SettingsFrameType,
					flags:   0x01,
					payload: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				},
			},
			want: FrameSizeError,
		},
		{
			name: "invalid_payload_len",
			settingsFrame: &SettingsFrame{
				frame: &frame{
					typ:     SettingsFrameType,
					payload: []byte{0x00, 0x00, 0x00},
				},
			},
			want: FrameSizeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnwrapErrorCode(tt.settingsFrame.Verify())
			if got != tt.want {
				t.Errorf("Verify() got = %s, want = %s", got.String(), tt.want.String())
			}
		})
	}
}

func TestSettingsFrame_Params(t *testing.T) {
	tests := []struct {
		frame *frame
		want  []*SettingsFrameParam
	}{
		{
			frame: &frame{
				typ:      SettingsFrameType,
				flags:    0,
				streamID: 0,
				payload: []byte{
					0x00, 0x01, 0x12, 0x34, 0x56, 0x78,
					0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x03, 0x87, 0x65, 0x43, 0x21,
					0xFF,
				},
			},
			want: []*SettingsFrameParam{
				{ID: HeaderTableSizeSetting, Value: 0x12345678},
				{ID: MaxConcurrentStreamsSetting, Value: 0x87654321},
			},
		},
	}

	for _, tt := range tests {
		got := (&SettingsFrame{frame: tt.frame}).Params()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Params() got = %+v, want = %+v", got, tt.want)
		}
	}
}

func TestSettingsFrameParam_IsUnknown(t *testing.T) {
	tests := []struct {
		param *SettingsFrameParam
		want  bool
	}{
		{param: &SettingsFrameParam{ID: 0x00}, want: true},
		{param: &SettingsFrameParam{ID: HeaderTableSizeSetting}, want: false},
		{param: &SettingsFrameParam{ID: MaxHeaderListSizeSetting}, want: false},
		{param: &SettingsFrameParam{ID: 0x07}, want: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("0x%X", tt.param.ID), func(t *testing.T) {
			got := tt.param.IsUnknown()
			if got != tt.want {
				t.Errorf("IsUnknown() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestSettingsFrameParam_Verify(t *testing.T) {
	tests := []struct {
		param *SettingsFrameParam
		want  ErrorCode
	}{
		{param: &SettingsFrameParam{ID: EnablePushSetting, Value: 0}, want: NoError},
		{param: &SettingsFrameParam{ID: EnablePushSetting, Value: 1}, want: NoError},
		{param: &SettingsFrameParam{ID: EnablePushSetting, Value: 2}, want: ProtocolError},
		{param: &SettingsFrameParam{ID: InitialWindowSizeSetting, Value: (1 << 16) - 1}, want: NoError},
		{param: &SettingsFrameParam{ID: InitialWindowSizeSetting, Value: (1 << 31) - 1}, want: NoError},
		{param: &SettingsFrameParam{ID: InitialWindowSizeSetting, Value: 1 << 31}, want: FlowControlError},
		{param: &SettingsFrameParam{ID: MaxFrameSizeSetting, Value: (1 << 14) - 1}, want: FlowControlError},
		{param: &SettingsFrameParam{ID: MaxFrameSizeSetting, Value: 1 << 14}, want: NoError},
		{param: &SettingsFrameParam{ID: MaxFrameSizeSetting, Value: (1 << 24) - 1}, want: NoError},
		{param: &SettingsFrameParam{ID: MaxFrameSizeSetting, Value: 1 << 24}, want: FlowControlError},
		{param: &SettingsFrameParam{ID: HeaderTableSizeSetting}, want: NoError},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id=0x%X,value=%d", tt.param.ID, tt.param.Value), func(t *testing.T) {
			got := UnwrapErrorCode(tt.param.Verify())
			if got != tt.want {
				t.Errorf("Verify() got = %s, want = %s", got.String(), tt.want.String())
			}
		})
	}
}

func TestSettingsFrameBuilder_Build(t *testing.T) {
	type want struct {
		frame *frame
		err   ErrorCode
	}

	tests := []struct {
		name    string
		builder *SettingsFrameBuilder
		want    want
	}{
		{
			name:    "empty",
			builder: NewSettingsFrameBuilder(),
			want: want{
				frame: &frame{
					typ:      SettingsFrameType,
					flags:    0,
					streamID: 0,
					payload:  []byte{},
				},
			},
		},
		{
			name:    "ack",
			builder: NewSettingsFrameBuilder().ACK(),
			want: want{
				frame: &frame{
					typ:      SettingsFrameType,
					flags:    0x01,
					streamID: 0,
					payload:  []byte{},
				},
			},
		},
		{
			name:    "ack_with_param",
			builder: NewSettingsFrameBuilder().ACK().Add(&SettingsFrameParam{ID: HeaderTableSizeSetting}),
			want:    want{err: FrameSizeError},
		},
		{
			name: "params",
			builder: NewSettingsFrameBuilder().
				Add(&SettingsFrameParam{ID: HeaderTableSizeSetting, Value: 0x12345678}).
				Add(&SettingsFrameParam{ID: MaxConcurrentStreamsSetting, Value: 0x12345678}),
			want: want{
				frame: &frame{
					typ:      SettingsFrameType,
					flags:    0x00,
					streamID: 0,
					payload: []byte{
						0x00, 0x01, 0x12, 0x34, 0x56, 0x78,
						0x00, 0x03, 0x12, 0x34, 0x56, 0x78,
					},
				},
			},
		},
		{
			name:    "invalid_param",
			builder: NewSettingsFrameBuilder().Add(&SettingsFrameParam{ID: EnablePushSetting, Value: 0x02}),
			want: want{
				err: ProtocolError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.builder.Build()

			errCode := UnwrapErrorCode(gotErr)
			if errCode != tt.want.err {
				t.Errorf("Build() got = %s, want = %v", errCode.String(), tt.want.err.String())
			}
			if gotErr != nil {
				return
			}

			w := &SettingsFrame{frame: tt.want.frame}
			if !reflect.DeepEqual(got, w) {
				t.Errorf("Build() got = %+v, want = %+v", got, w)
			}
		})
	}
}

func TestHeadersFrame_Verify(t *testing.T) {
	tests := []struct {
		name  string
		frame *frame
		want  ErrorCode
	}{
		{
			name: "valid",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0x28,
				streamID: 0x1234,
				payload: []byte{
					0x05,
					0x00, 0x00, 0x00, 0x01,
					0xFF,
					0x12, 0x34, 0x56,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				},
			},
			want: NoError,
		},
		{
			name: "zero-stream-id",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0,
				streamID: 0x00,
				payload:  []byte{},
			},
			want: ProtocolError,
		},
		{
			name: "invalid-pad",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0x08,
				streamID: 0x1234,
				payload:  []byte{0x05, 0xFF, 0xFF, 0xFF, 0xFF},
			},
			want: ProtocolError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnwrapErrorCode((&HeadersFrame{frame: tt.frame}).Verify())
			if got != tt.want {
				t.Errorf("Verify() got = %s, want = %s", got.String(), tt.want.String())
			}
		})
	}
}

func TestHeadersFrame_Fragment(t *testing.T) {
	tests := []struct {
		name  string
		frame *frame
		want  []byte
	}{
		{
			name: "no-flags",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0,
				streamID: 0x1234,
				payload:  []byte{0x12, 0x34, 0x56},
			},
			want: []byte{0x12, 0x34, 0x56},
		},
		{
			name: "padded",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0x08,
				streamID: 0x1234,
				payload: []byte{
					0x05,
					0x12, 0x34, 0x56,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				},
			},
			want: []byte{0x12, 0x34, 0x56},
		},
		{
			name: "prioritized",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0x20,
				streamID: 0x1234,
				payload: []byte{
					0x00, 0x00, 0x00, 0x01,
					0xFF,
					0x12, 0x34, 0x56,
				},
			},
			want: []byte{0x12, 0x34, 0x56},
		},
		{
			name: "padded+prioritized",
			frame: &frame{
				typ:      HeadersFrameType,
				flags:    0x28,
				streamID: 0x1234,
				payload: []byte{
					0x05,
					0x00, 0x00, 0x00, 0x01,
					0xFF,
					0x12, 0x34, 0x56,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				},
			},
			want: []byte{0x12, 0x34, 0x56},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&HeadersFrame{frame: tt.frame}).Fragment()
			if bytes.Compare(got, tt.want) != 0 {
				t.Errorf("Fragment() got = %+v, want = %+v", got, tt.want)
			}
		})
	}
}

func TestPingFrame_ACK(t *testing.T) {
	type want struct {
		frame *frame
		err   ErrorCode
	}

	tests := []struct {
		frame *frame
		want  want
	}{
		{
			frame: &frame{
				typ:      PingFrameType,
				flags:    0,
				streamID: 0,
				payload:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			want: want{
				frame: &frame{
					typ:      PingFrameType,
					flags:    0x01,
					streamID: 0,
					payload:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
			},
		},
	}

	for _, tt := range tests {
		ack, err := (&PingFrame{frame: tt.frame}).ACK()
		errCode := UnwrapErrorCode(err)

		if errCode != tt.want.err {
			t.Errorf("ACK() got = %s, want = %s", errCode.String(), tt.want.err.String())
		}
		if err != nil {
			continue
		}

		want := &PingFrame{frame: tt.want.frame}
		if !reflect.DeepEqual(ack, want) {
			t.Errorf("ACK() got = %+v, want = %+v", ack, want)
		}
	}
}
