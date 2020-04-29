package huffman

import (
	"bytes"
)

// Encode encodes any data to Huffman encoded data that is compliant with HPACK specification.
func Encode(data []byte) []byte {
	encoded := bytes.NewBuffer(nil)

	var curByte byte
	var wroteBits uint8

	for _, b := range data {
		code := codeTable[b].code
		bits := codeTable[b].bitsLen

		for bits > 0 {
			writeBits := 8 - wroteBits
			if writeBits > bits {
				writeBits = bits
			}

			curByte |= (byte(code >> (bits - writeBits))) << (8 - wroteBits - writeBits)
			wroteBits += writeBits
			bits -= writeBits
			code &= (1 << bits) - 1

			if wroteBits == 8 {
				encoded.WriteByte(curByte)
				curByte = 0
				wroteBits = 0
			}
		}
	}

	if wroteBits != 0 {
		curByte |= (1 << (8 - wroteBits)) - 1
		encoded.WriteByte(curByte)
	}

	return encoded.Bytes()
}
