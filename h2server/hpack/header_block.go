package hpack

import (
	"io"
)

// peekReader is io.Reader implementation and allows peeking byte.
type peekReader struct {
	r    io.Reader
	peek *byte
}

var (
	_ io.Reader = (*peekReader)(nil)
)

func DecodeHeaderBlock(table *IndexTable, r *io.LimitedReader) (HeaderList, error) {
	pkr := newPeekReader(r)
	headerList := HeaderList{}

	for r.N > 0 {
		peeked, err := pkr.Peek()
		if err != nil {
			return nil, err
		}

		var hf *HeaderField
		switch {
		case peeked >= 128: // Indexed Header Field
			hf, err = decodeIndexedHeaderField(table, pkr)

		case peeked >= 64: // Literal Header Field with Incremental Indexing
			hf, err = decodeLiteralHeaderField(table, pkr, 6, true)

		case peeked >= 32: // Dynamic Table Size Update
			_, newDataSize, err := decodePrefixedInt(pkr, 5)
			if err == nil {
				err = table.UpdateMaxDataSize(int(newDataSize))
			}

		default: // Literal Header Field without Indexing, Literal Header Field Never Indexed
			hf, err = decodeLiteralHeaderField(table, pkr, 4, false)
		}

		if err != nil {
			return nil, err
		}
		if hf != nil {
			headerList = append(headerList, hf)
		}
	}

	return headerList, nil
}

func decodeIndexedHeaderField(table *IndexTable, r io.Reader) (*HeaderField, error) {
	_, value, err := decodePrefixedInt(r, 7)
	if err != nil {
		return nil, err
	}

	indexed := table.Entry(int(value))
	if indexed == nil {
		return nil, ErrTableEntryNotFound
	}

	return indexed, nil
}

func decodeLiteralHeaderField(table *IndexTable, r io.Reader, prefixedIntN int, beIndexed bool) (*HeaderField, error) {
	_, nameIndex, err := decodePrefixedInt(r, prefixedIntN)
	if err != nil {
		return nil, err
	}

	var name []byte
	if nameIndex > 0 {
		indexed := table.Entry(int(nameIndex))
		if indexed == nil {
			return nil, ErrTableEntryNotFound
		}
		name = indexed.Name()
	} else {
		name, err = decodeStringLiteral(r, table.MaxDataSize())
		if err != nil {
			return nil, err
		}
	}

	// Probably, we should consider 32 bytes overhead to compute max length
	value, err := decodeStringLiteral(r, table.MaxDataSize())
	if err != nil {
		return nil, err
	}

	hf := &HeaderField{name: name, value: value}
	if beIndexed {
		table.AddEntry(hf)
	}

	return hf, nil
}

func newPeekReader(r io.Reader) *peekReader {
	return &peekReader{r: r}
}

func (pk *peekReader) Peek() (byte, error) {
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

func (pk *peekReader) Read(buf []byte) (int, error) {
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
