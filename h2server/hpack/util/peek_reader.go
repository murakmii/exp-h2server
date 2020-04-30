package util

import "io"

// PeekReader is io.Reader implementation and allows peeking byte.
type PeekReader struct {
	r    io.Reader
	peek *byte
}

var _ io.Reader = (*PeekReader)(nil)

func NewPeekReader(r io.Reader) *PeekReader {
	return &PeekReader{r: r}
}

func (pk *PeekReader) Peek() (byte, error) {
	if pk.peek != nil {
		return *pk.peek, nil
	}

	buf := make([]byte, 1)
	if _, err := pk.r.Read(buf); err != nil {
		return 0, err
	}

	pk.peek = &buf[0]
	return buf[0], nil
}

func (pk *PeekReader) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	offset := 0
	if pk.peek != nil {
		buf[0] = *pk.peek
		offset = 1
	}

	read, err := pk.r.Read(buf[offset:])
	if err != nil {
		return 0, err
	}

	pk.peek = nil
	return offset + read, nil
}
