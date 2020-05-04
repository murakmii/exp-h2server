package hpack

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

type (
	HeaderList []*HeaderField

	HeaderField struct {
		name  string
		value string
	}
)

const (
	// See: https://tools.ietf.org/html/rfc7541#section-4.1
	headerFieldManagingOverheadBytes = 32
)

func (hl HeaderList) Validate() error {
	for _, hf := range hl {
		if err := hf.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (hl HeaderList) Encode() []byte {
	buf := bytes.NewBuffer([]byte{0x10})
	for _, hf := range hl {
		buf.Write(encodeStringLiteral(hf.Name(), true))
		buf.Write(encodeStringLiteral(hf.Value(), true))
	}
	return buf.Bytes()
}

func NewHeaderField(name, value string) *HeaderField {
	return &HeaderField{
		name:  name,
		value: value,
	}
}

func (hf *HeaderField) Name() string {
	return hf.name
}

func (hf *HeaderField) Value() string {
	return hf.value
}

func (hf *HeaderField) DataSize() int {
	return len(hf.name) + len(hf.value) + headerFieldManagingOverheadBytes
}

func (hf *HeaderField) IsPseudo() bool {
	return hf.name == ":authority" ||
		hf.name == ":scheme" ||
		hf.name == ":method" ||
		hf.name == ":path" ||
		hf.name == ":status"
}

func (hf *HeaderField) Validate() error {
	if !hf.validHeaderName() || !hf.validHeaderValue() {
		return fmt.Errorf("%w: used invalid character", ErrHeader)
	}

	return nil
}

func (hf *HeaderField) validHeaderName() bool {
	if hf.IsPseudo() {
		return true
	}

	for _, c := range hf.name {
		if !availableHeaderNameChars[c] {
			return false
		}
	}
	return true
}

func (hf *HeaderField) validHeaderValue() bool {
	for _, c := range hf.value {
		if c >= utf8.RuneSelf || (c < ' ' && c != '\t') || c == 0x7f {
			return false
		}
	}

	return true
}

var (
	availableHeaderNameChars = &[256]bool{
		'!':  true,
		'#':  true,
		'$':  true,
		'%':  true,
		'&':  true,
		'\'': true,
		'*':  true,
		'+':  true,
		'-':  true,
		'.':  true,
		'0':  true,
		'1':  true,
		'2':  true,
		'3':  true,
		'4':  true,
		'5':  true,
		'6':  true,
		'7':  true,
		'8':  true,
		'9':  true,
		'^':  true,
		'_':  true,
		'`':  true,
		'a':  true,
		'b':  true,
		'c':  true,
		'd':  true,
		'e':  true,
		'f':  true,
		'g':  true,
		'h':  true,
		'i':  true,
		'j':  true,
		'k':  true,
		'l':  true,
		'm':  true,
		'n':  true,
		'o':  true,
		'p':  true,
		'q':  true,
		'r':  true,
		's':  true,
		't':  true,
		'u':  true,
		'v':  true,
		'w':  true,
		'x':  true,
		'y':  true,
		'z':  true,
		'|':  true,
		'~':  true,
	}
)
