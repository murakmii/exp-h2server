package hpack

import (
	"errors"
	"fmt"
)

var (
	ErrHPACK = errors.New("hpack")

	ErrTableEntryNotFound = fmt.Errorf("%w: specified table entry not found", ErrHPACK)

	ErrDataSize = fmt.Errorf("%w: data size", ErrHPACK)

	ErrPrefixedInt   = fmt.Errorf("%w: prefixed int", ErrHPACK)
	ErrStringLiteral = fmt.Errorf("%w: string literal", ErrHPACK)
)
