package h2server

import (
	"encoding/binary"
	"fmt"
	"io"
)

type (
	FrameType uint8

	Frame interface {
		io.WriterTo
		Type() FrameType
		Flags() uint8
		StreamID() uint32
		Payload() []byte
	}

	IncomingFrame interface {
		Frame
		Verify() error
	}

	frame struct {
		typ      FrameType
		flags    uint8
		streamID uint32
		payload  []byte
	}

	SettingsFrame struct {
		*frame
	}

	SettingsFrameParamID uint16

	SettingsFrameParam struct {
		ID    SettingsFrameParamID
		Value uint32
	}

	SettingsFrameBuilder struct {
		ack    bool
		params []*SettingsFrameParam
	}

	DataFrame struct {
		*frame
	}

	HeadersFrame struct {
		*frame
	}

	RstStreamFrame struct {
		*frame
	}

	PushPromiseFrame struct {
		*frame
	}

	PingFrame struct {
		*frame
	}

	GoAwayFrame struct {
		*frame
	}

	ContinuationFrame struct {
		*frame
	}

	UnknownFrame struct {
		*frame
	}
)

const (
	streamIDMask = 0x7fffffff

	DataFrameType         FrameType = 0x00
	HeadersFrameType      FrameType = 0x01
	PriorityFrameType     FrameType = 0x02
	RstStreamFrameType    FrameType = 0x03
	SettingsFrameType     FrameType = 0x04
	PushPromiseFrameType  FrameType = 0x05
	PingFrameType         FrameType = 0x06
	GoAwayFrameType       FrameType = 0x07
	WindowUpdateFrameType FrameType = 0x08
	ContinuationFrameType FrameType = 0x09

	HeaderTableSizeSetting      SettingsFrameParamID = 0x01
	EnablePushSetting           SettingsFrameParamID = 0x02
	MaxConcurrentStreamsSetting SettingsFrameParamID = 0x03
	InitialWindowSizeSetting    SettingsFrameParamID = 0x04
	MaxFrameSizeSetting         SettingsFrameParamID = 0x05
	MaxHeaderListSizeSetting    SettingsFrameParamID = 0x06
)

func (typ FrameType) IsUnknown() bool {
	return typ > ContinuationFrameType
}

func Read(r io.Reader) (IncomingFrame, error) {
	buf := make([]byte, 9)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	payloadLen := 0
	for i := 0; i < 3; i++ {
		payloadLen |= int(buf[i]) << ((2 - i) * 8)
	}

	f := &frame{
		typ:      FrameType(buf[3]),
		flags:    buf[4],
		streamID: binary.BigEndian.Uint32(buf[5:9]) & streamIDMask,
		payload:  make([]byte, payloadLen),
	}

	if _, err := r.Read(f.payload); err != nil {
		return nil, err
	}

	var ret IncomingFrame
	switch f.typ {
	case DataFrameType:
		ret = &DataFrame{frame: f}

	case HeadersFrameType:
		ret = &HeadersFrame{frame: f}

	case PushPromiseFrameType:
		ret = &PushPromiseFrame{frame: f}

	case ContinuationFrameType:
		ret = &ContinuationFrame{frame: f}

	case SettingsFrameType:
		ret = &SettingsFrame{frame: f}

	case RstStreamFrameType:
		ret = &RstStreamFrame{frame: f}

	case PingFrameType:
		ret = &PingFrame{frame: f}

	case GoAwayFrameType:
		ret = &GoAwayFrame{frame: f}

	default:
		ret = &UnknownFrame{frame: f}
	}

	return ret, nil
}

func (la *frame) Type() FrameType {
	return la.typ
}

func (la *frame) Flags() uint8 {
	return la.flags
}

func (la *frame) StreamID() uint32 {
	return la.streamID
}

func (la *frame) Payload() []byte {
	return la.payload
}

func (la *frame) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 9)
	for i := 0; i < 3; i++ {
		buf[i] = byte((len(la.payload) >> (8 * (2 - i))) & 0xff)
	}
	buf[3] = byte(la.typ)
	buf[4] = la.flags
	binary.BigEndian.PutUint32(buf[5:], la.streamID&streamIDMask)

	var sum int64
	n, err := w.Write(buf)
	if err != nil {
		return int64(n), err
	}
	sum += int64(n)

	n, err = w.Write(la.payload)
	return sum + int64(n), err
}

func (data *DataFrame) IsEOS() bool {
	return (data.flags & 0x01) > 0
}

func (data *DataFrame) IsPadded() bool {
	return (data.flags & 0x08) > 0
}

func (data *DataFrame) Verify() error {
	if data.streamID == 0x00 {
		return NewH2Error(ProtocolError, "data frame's stream ID(0) is invalid")
	}

	if !data.IsPadded() {
		return nil
	}

	pLen := len(data.payload)
	if pLen == 0 || pLen < int(data.payload[0])+1 {
		return NewH2Error(FrameSizeError, "data frame's payload length is invalid")
	}

	return nil
}

func (data *DataFrame) Data() []byte {
	if data.IsPadded() {
		return data.payload[1 : len(data.payload)-int(data.payload[0])]
	}

	return data.payload
}

func (settings *SettingsFrame) IsACK() bool {
	return (settings.Flags() & 0x01) > 0
}

func (settings *SettingsFrame) Verify() error {
	if settings.StreamID() != 0x0 {
		return NewH2Error(ProtocolError, "settings frame's stream ID must be 0x00")
	}

	pLen := len(settings.payload)
	if settings.IsACK() && pLen != 0 {
		return ACKSettingsFrameErr
	}

	if pLen%6 != 0 {
		return NewH2Error(FrameSizeError, "settings frame's payload length is invalid")
	}

	return nil
}

func (settings *SettingsFrame) Params() []*SettingsFrameParam {
	n := len(settings.payload) / 6
	params := make([]*SettingsFrameParam, 0, n)

	for i := 0; i < n; i++ {
		offset := i * 6
		param := &SettingsFrameParam{
			ID:    SettingsFrameParamID(binary.BigEndian.Uint16(settings.payload[offset : offset+2])),
			Value: binary.BigEndian.Uint32(settings.payload[offset+2 : offset+6]),
		}

		if param.IsUnknown() {
			continue
		}
		params = append(params, param)
	}

	return params
}

func (settingsParam *SettingsFrameParam) IsUnknown() bool {
	return settingsParam.ID == 0x00 || settingsParam.ID > 0x06
}

func (settingsParam *SettingsFrameParam) Verify() error {
	switch settingsParam.ID {
	case EnablePushSetting:
		if settingsParam.Value > 1 {
			return NewH2Error(ProtocolError, "enable push settings value is invalid(%d)", settingsParam.Value)
		}

	case InitialWindowSizeSetting:
		if settingsParam.Value > ((1 << 31) - 1) {
			return NewH2Error(FlowControlError, "initial window size settings value is invalid(%d)", settingsParam.Value)
		}

	case MaxFrameSizeSetting:
		if settingsParam.Value < (1<<14) || settingsParam.Value > ((1<<24)-1) {
			return NewH2Error(FlowControlError, "max frame size settings value is invalid(%d)", settingsParam.Value)
		}
	}

	return nil
}

func NewSettingsFrameBuilder() *SettingsFrameBuilder {
	return &SettingsFrameBuilder{
		ack:    false,
		params: make([]*SettingsFrameParam, 0, 6),
	}
}

func (sfb *SettingsFrameBuilder) ACK() *SettingsFrameBuilder {
	sfb.ack = true
	return sfb
}

func (sfb *SettingsFrameBuilder) Add(params ...*SettingsFrameParam) *SettingsFrameBuilder {
	for _, param := range params {
		sfb.params = append(sfb.params, param)
	}
	return sfb
}

func (sfb *SettingsFrameBuilder) Build() (Frame, error) {
	pLen := len(sfb.params)
	if sfb.ack && pLen > 0 {
		return nil, ACKSettingsFrameErr
	}

	f := &frame{
		typ:      SettingsFrameType,
		flags:    0,
		streamID: 0,
		payload:  make([]byte, pLen*6),
	}

	if sfb.ack {
		f.flags = 0x01
	}

	for i, param := range sfb.params {
		if err := param.Verify(); err != nil {
			return nil, fmt.Errorf("can't build settings frame: %w", err)
		}

		offset := i * 6
		binary.BigEndian.PutUint16(f.payload[offset:], uint16(param.ID))
		binary.BigEndian.PutUint32(f.payload[offset+2:], param.Value)
	}

	return &SettingsFrame{frame: f}, nil
}

func (headers *HeadersFrame) IsEOS() bool {
	return (headers.flags & 0x01) > 0
}

func (headers *HeadersFrame) IsEndHeaders() bool {
	return (headers.flags & 0x04) > 0
}

func (headers *HeadersFrame) IsPadded() bool {
	return (headers.flags & 0x08) > 0
}

func (headers *HeadersFrame) IsPrioritized() bool {
	return (headers.flags & 0x20) > 0
}

func (headers *HeadersFrame) Verify() error {
	if headers.streamID == 0x00 {
		return NewH2Error(ProtocolError, "headers frame's stream ID(0) is invalid")
	}

	frgLen := len(headers.payload)
	if headers.IsPadded() {
		frgLen -= 1
		if frgLen < 0 {
			return NewH2Error(ProtocolError, "headers frame's payload length is invalid")
		}

		frgLen -= int(headers.payload[0])
	}

	if headers.IsPrioritized() {
		frgLen -= 5
	}

	if frgLen < 0 {
		return NewH2Error(ProtocolError, "headers frame's payload length is invalid")
	}

	return nil
}

func (headers *HeadersFrame) Fragment() []byte {
	frgEnd := len(headers.payload)
	offset := 0

	if headers.IsPadded() {
		offset += 1
		frgEnd -= int(headers.payload[0])
	}

	if headers.IsPrioritized() {
		offset += 5
	}

	return headers.payload[offset:frgEnd]
}

func (push *PushPromiseFrame) IsEndOfHeaders() bool {
	return (push.flags & 0x01) > 0
}

func (push *PushPromiseFrame) IsPadded() bool {
	return (push.flags & 0x08) > 0
}

func (push *PushPromiseFrame) Verify() error {
	if push.streamID == 0x00 {
		return NewH2Error(ProtocolError, "push promise frame's stream ID(0) is invalid")
	}

	reqLen := 4
	pLen := len(push.payload)
	if push.IsPadded() {
		if pLen == 0 {
			return NewH2Error(FrameSizeError, "push promise's payload length is invalid")
		}

		reqLen += 1 + int(push.payload[0])
	}

	if pLen < reqLen {
		return NewH2Error(FrameSizeError, "push promise's payload length is invalid")
	}

	return nil
}

func (push *PushPromiseFrame) PromisedStreamID() uint32 {
	if push.IsPadded() {
		return binary.BigEndian.Uint32(push.payload[1:])
	}

	return binary.BigEndian.Uint32(push.payload)
}

func (push *PushPromiseFrame) HeaderFragment() []byte {
	frgEnd := len(push.payload)
	offset := 4

	if push.IsPadded() {
		offset += 1
		frgEnd -= int(push.payload[0])
	}

	return push.payload[offset:frgEnd]
}

func (rst *RstStreamFrame) Verify() error {
	if rst.streamID == 0x00 {
		return NewH2Error(ProtocolError, "rst frame's stream ID(0) is invalid")
	}

	pLen := len(rst.payload)
	if pLen != 4 {
		return NewH2Error(FrameSizeError, "rst frame's payload length(%d octets) is invalid", pLen)
	}

	return nil
}

func (rst *RstStreamFrame) ErrorCode() ErrorCode {
	return ErrorCode(binary.BigEndian.Uint32(rst.payload))
}

func (ping *PingFrame) IsACK() bool {
	return (ping.flags & 0x01) > 0
}

func (ping *PingFrame) Verify() error {
	if ping.streamID != 0x00 {
		return NewH2Error(ProtocolError, "ping frame's stream ID(%d) is invalid", ping.streamID)
	}

	pLen := len(ping.payload)
	if pLen != 8 {
		return NewH2Error(FrameSizeError, "ping frame's payload length(%d octets) is invalid", pLen)
	}

	return nil
}

func (ping *PingFrame) ACK() (Frame, error) {
	if ping.IsACK() {
		return nil, NewH2Error(InternalError, "can't ack to ack ping frame")
	}

	return &PingFrame{
		frame: &frame{
			typ:      PingFrameType,
			flags:    0x01,
			streamID: 0,
			payload:  ping.payload,
		},
	}, nil
}

func (cont *ContinuationFrame) IsEndHeaders() bool {
	return (cont.flags & 0x04) > 0
}

func (cont *ContinuationFrame) Verify() error {
	if cont.streamID == 0x00 {
		return NewH2Error(ProtocolError, "continuation frame's stream ID(0) is invalid")
	}

	return nil
}

func (cont *ContinuationFrame) Fragment() []byte {
	return cont.payload
}

func (goAway *GoAwayFrame) Verify() error {
	if goAway.streamID != 0x00 {
		return NewH2Error(ProtocolError, "go away frame's stream ID must be 0x00")
	}

	pLen := len(goAway.payload)
	if pLen < 8 {
		return NewH2Error(FrameSizeError, "go away frame's payload length must be greater than 8 octets")
	}

	return nil
}

func (goAway *GoAwayFrame) LastStreamID() uint32 {
	return binary.BigEndian.Uint32(goAway.payload) & streamIDMask
}

func (goAway *GoAwayFrame) ErrorCode() ErrorCode {
	return ErrorCode(binary.BigEndian.Uint32(goAway.payload[4:]))
}

func (goAway *GoAwayFrame) DebugData() []byte {
	return goAway.payload[8:]
}

func (unknown *UnknownFrame) Verify() error {
	return nil
}
