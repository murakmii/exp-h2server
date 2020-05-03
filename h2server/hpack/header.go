package hpack

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

func (hf *HeaderField) Name() string {
	return hf.name
}

func (hf *HeaderField) Value() string {
	return hf.value
}

func (hf *HeaderField) DataSize() int {
	return len(hf.name) + len(hf.value) + headerFieldManagingOverheadBytes
}
