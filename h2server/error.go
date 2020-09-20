package h2server

import (
	"errors"
	"fmt"
)

type (
	H2Error struct {
		code    ErrorCode
		wrapped error
	}

	ErrorCode uint32
)

var (
	_ error = (*H2Error)(nil)

	UnknownFrameErr     = errors.New("unknown frame type")
	ACKSettingsFrameErr = NewH2Error(FrameSizeError, "ack settings frame's payload length must be 0")
)

const (
	NoError                 ErrorCode = 0x00
	ProtocolError           ErrorCode = 0x01
	InternalError           ErrorCode = 0x02
	FlowControlError        ErrorCode = 0x03
	SettingsTimeoutError    ErrorCode = 0x04
	StreamClosedError       ErrorCode = 0x05
	FrameSizeError          ErrorCode = 0x06
	RefusedStreamError      ErrorCode = 0x07
	CancelError             ErrorCode = 0x08
	CompressionError        ErrorCode = 0x09
	ConnectError            ErrorCode = 0x0a
	EnhanceYourCalmError    ErrorCode = 0x0b
	InadequateSecurityError ErrorCode = 0x0c
	HTTP11RequiredError     ErrorCode = 0x0d
)

func NewH2Error(code ErrorCode, format string, args ...interface{}) error {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(format, args...)
	} else {
		err = errors.New(format)
	}

	return &H2Error{code: code, wrapped: err}
}

func UnwrapErrorCode(err error) ErrorCode {
	if err == nil {
		return NoError
	}

	if h2Error, ok := err.(*H2Error); ok {
		return h2Error.code
	}

	if wrapErr, ok := err.(interface{ Unwrap() error }); ok {
		err = wrapErr.Unwrap()
		if err == nil {
			return InternalError
		}

		return UnwrapErrorCode(err)
	}

	return InternalError
}

func (err *H2Error) Code() ErrorCode {
	return err.code
}

func (err *H2Error) Error() string {
	return err.code.String() + ": " + err.wrapped.Error()
}

func (err *H2Error) Unwrap() error {
	return err.wrapped
}

func (code ErrorCode) IsUnknown() bool {
	return code > HTTP11RequiredError
}

func (code ErrorCode) String() string {
	switch code {
	case NoError:
		return "no error"
	case ProtocolError:
		return "protocol error"
	case InternalError:
		return "internal error"
	case FlowControlError:
		return "flow control error"
	case SettingsTimeoutError:
		return "settings timeout"
	case StreamClosedError:
		return "stream closed"
	case FrameSizeError:
		return "frame size error"
	case RefusedStreamError:
		return "refused stream"
	case CancelError:
		return "cancel error"
	case CompressionError:
		return "compression error"
	case ConnectError:
		return "connect error"
	case EnhanceYourCalmError:
		return "enhance your calm"
	case InadequateSecurityError:
		return "inadequate security error"
	case HTTP11RequiredError:
		return "http1.1 required"
	default:
		return "unknown error"
	}
}

func NewUnknownFrameError(typ FrameType) error {
	return fmt.Errorf("%w: 0x%X", UnknownFrameErr, typ)
}
