package hpack

type (
	HeaderList []*HeaderField

	HeaderField struct {
		name  []byte
		value []byte
	}
)

const (
	// See: https://tools.ietf.org/html/rfc7541#section-4.1
	headerFieldManagingOverheadBytes = 32
)

func (hf *HeaderField) Name() []byte {
	return hf.name
}

func (hf *HeaderField) Value() []byte {
	return hf.value
}

func (hf *HeaderField) DataSize() int {
	return len(hf.name) + len(hf.value) + headerFieldManagingOverheadBytes
}
