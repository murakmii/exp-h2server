package hpack

import (
	"errors"
	"io"

	"github.com/murakmii/exp-h2server/h2server/hpack/huffman"

	"github.com/murakmii/exp-h2server/h2server/hpack/util"
)

type (
	HeaderList []*HeaderField

	HeaderField struct {
		name  []byte
		value []byte
	}
)

var (
	errNoEntry             = errors.New("no table entry")
	errInvalidStringLength = errors.New("invalid string length")
)

func ProcessHeaderBlock(ias *IndexAddressSpace, r *io.LimitedReader) (HeaderList, error) {
	pkr := util.NewPeekReader(r)
	headerList := HeaderList{}

	for r.N > 0 {
		peeked, err := pkr.Peek()
		if err != nil {
			return nil, err
		}

		var hf *HeaderField
		switch {
		case peeked >= 128: // Indexed Header Field
			hf, err = processIndexedHeaderField(ias, pkr)

		case peeked >= 64: // Literal Header Field with Incremental Indexing
			hf, err = processLiteralHeaderField(ias, pkr, 6, true)

		case peeked >= 32: // Dynamic Table Size Update
			// TODO: update dynamic table size

		default: // Literal Header Field without Indexing, Literal Header Field Never Indexed
			hf, err = processLiteralHeaderField(ias, pkr, 4, false)
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

func processIndexedHeaderField(ias *IndexAddressSpace, r io.Reader) (*HeaderField, error) {
	_, value, err := util.DecodePrefixedInt(r, 7)
	if err != nil {
		return nil, err
	}

	indexed := ias.Entry(int(value))
	if indexed == nil {
		return nil, errNoEntry
	}

	return indexed, nil
}

func processLiteralHeaderField(ias *IndexAddressSpace, r io.Reader, prefixBits int, beIndexed bool) (*HeaderField, error) {
	_, nameIndex, err := util.DecodePrefixedInt(r, prefixBits)
	if err != nil {
		return nil, err
	}

	var name []byte
	if nameIndex > 0 {
		indexed := ias.Entry(int(nameIndex))
		if indexed == nil {
			return nil, errNoEntry
		}
		name = indexed.Name()
	} else {
		name, err = decodeStringLiteral(r, ias.Dynamic().MaxDataSize())
		if err != nil {
			return nil, err
		}
	}

	// Probably, we should consider 32 bytes overhead to compute max length
	value, err := decodeStringLiteral(r, ias.Dynamic().MaxDataSize())
	if err != nil {
		return nil, err
	}

	hf := &HeaderField{name: name, value: value}
	if beIndexed {
		ias.Dynamic().AddEntry(hf)
	}

	return hf, nil
}

func decodeStringLiteral(r io.Reader, maxLength int) ([]byte, error) {
	encodedFlag, length, err := util.DecodePrefixedInt(r, 7)
	if err != nil {
		return nil, err
	}

	if length > uint64(maxLength) {
		return nil, errInvalidStringLength
	}

	str := make([]byte, length)
	if _, err := r.Read(str); err != nil {
		return nil, err
	}

	if (encodedFlag >> 7) == 1 {
		str, err = huffman.Decode(str)
		if err != nil {
			return nil, err
		}
	}

	return str, nil
}

func (hf *HeaderField) Name() []byte {
	return hf.name
}

func (hf *HeaderField) Value() []byte {
	return hf.value
}

func (hf *HeaderField) DataSize() int {
	return len(hf.name) + len(hf.value) + 32
}
