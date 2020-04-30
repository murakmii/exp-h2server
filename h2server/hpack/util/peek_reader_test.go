package util

import (
	"bytes"
	"testing"
)

func TestPeekReader_PeekAndRead(t *testing.T) {
	in := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	r := NewPeekReader(bytes.NewReader(in))
	peek, err := r.Peek()
	if peek != 0x01 || err != nil {
		t.Errorf("Peek() got = %d/%v, want = 1/nil", peek, err)
	}

	buf := make([]byte, 2)
	if read, err := r.Read(buf); read != 2 || err != nil {
		t.Errorf("Read() got = %d/%v, want = 2/nil", peek, err)
	}

	if bytes.Compare(buf, in[0:2]) != 0 {
		t.Errorf("Read() read = %X, want = %X", buf, in[0:2])
	}

	peek, err = r.Peek()
	if peek != 0x03 || err != nil {
		t.Errorf("Peek() got = %d/%v, want = 3/nil", peek, err)
	}

	buf = make([]byte, 2)
	if read, err := r.Read(buf); read != 2 || err != nil {
		t.Errorf("Read() got = %d/%v, want = 2/nil", peek, err)
	}

	if bytes.Compare(buf, in[2:4]) != 0 {
		t.Errorf("Read() read = %X, want = %X", buf, in[2:4])
	}

	buf = make([]byte, 1)
	if read, err := r.Read(buf); read != 1 || err != nil {
		t.Errorf("Read() got = %d/%v, want = 2/nil", peek, err)
	}

	if bytes.Compare(buf, in[4:]) != 0 {
		t.Errorf("Read() read = %X, want = %X", buf, in[4:])
	}
}
