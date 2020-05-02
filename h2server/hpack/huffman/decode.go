package huffman

import (
	"bytes"
	"errors"
	"sync"
)

type (
	codeTreeNode struct {
		match    bool
		symbol   byte
		children *[2]*codeTreeNode
	}
)

var (
	codeTreeRoot      *codeTreeNode
	buildCodeTreeOnce sync.Once

	Err = errors.New("invalid huffman encoded data")
)

// Decode decodes Huffman encoded data that is compliant with HPACK specification.
func Decode(encoded []byte) ([]byte, error) {
	decoded := bytes.NewBuffer(nil)
	checkingBits := bytes.NewBuffer(nil)
	node := codeTree()

	// Follow tree to check whether matches bits and Huffman code.
	for _, b := range encoded {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			node = node.children[bit]
			if node == nil {
				return nil, Err
			}

			if node.match {
				decoded.WriteByte(node.symbol)
				checkingBits.Reset()
				node = codeTree()
			} else {
				checkingBits.WriteByte(bit)
			}
		}
	}

	// Check padding(EOS)
	padding := checkingBits.Bytes()
	for i, p := range padding {
		if i >= 7 || p != 1 {
			return nil, Err
		}
	}

	return decoded.Bytes(), nil
}

func codeTree() *codeTreeNode {
	buildCodeTreeOnce.Do(buildCodeTree)
	return codeTreeRoot
}

// Build once tree(binary trie) for decoding.
func buildCodeTree() {
	codeTreeRoot = newCodeTreeNode()
	for symbol, tableItem := range codeTable {
		codeTreeRoot.addChild(byte(symbol), tableItem.code, tableItem.bitsLen)
	}
}

func newCodeTreeNode() *codeTreeNode {
	return &codeTreeNode{}
}

func (node *codeTreeNode) addChild(symbol byte, code int, bitsLen uint8) {
	if bitsLen == 0 {
		node.match = true
		node.symbol = symbol
		return
	}

	if node.children == nil {
		node.children = &[2]*codeTreeNode{}
	}

	bit := (code >> (bitsLen - 1)) & 1

	child := node.children[bit]
	if child == nil {
		child = newCodeTreeNode()
		node.children[bit] = child
	}

	child.addChild(symbol, code, bitsLen-1)
}
