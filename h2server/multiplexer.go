package h2server

import (
	"fmt"
)

type (
	Multiplexer interface {
		Received(Frame)
		Terminated()
	}

	HttpMultiplexer struct {
		conn   Conn
		logger Logger
	}
)

func DefaultMultiplexer(logger Logger) func(Conn) Multiplexer {
	return func(conn Conn) Multiplexer {
		return &HttpMultiplexer{
			conn:   conn,
			logger: logger,
		}
	}
}

func (hmp *HttpMultiplexer) Received(frame Frame) {
	// TODO: Verify
	hmp.log(DebugLog, "received type=0x%X id=%d flags=0x%X payload=%d B", frame.Type(), frame.StreamID(), frame.Flags(), len(frame.Payload()))

	switch f := frame.(type) {
	case *SettingsFrame:
		hmp.handleSettings(f)
	}
}

func (hmp *HttpMultiplexer) Terminated() {

}

func (hmp *HttpMultiplexer) log(level LogLevel, format string, args ...interface{}) {
	hmp.logger.Write(level, fmt.Sprintf("<%s> ", hmp.conn.RemoteAddr().String())+format, args...)
}

func (hmp *HttpMultiplexer) handleSettings(f *SettingsFrame) {
	// TODO: Verify

}
