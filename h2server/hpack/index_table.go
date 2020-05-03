package hpack

import (
	"fmt"
)

type (
	// IndexTable is representation of "Index Table"
	// See: https://tools.ietf.org/html/rfc7541#section-2.3
	IndexTable struct {
		maxProtocolDataSize int
		maxDataSize         int
		currentDataSize     int
		dynamicTable        []*HeaderField
	}
)

func NewIndexTable(maxProtocolDataSize int) *IndexTable {
	return &IndexTable{
		maxProtocolDataSize: maxProtocolDataSize,
		maxDataSize:         maxProtocolDataSize,
		currentDataSize:     0,
		dynamicTable:        make([]*HeaderField, 0),
	}
}

func (table *IndexTable) EntriesCount() int {
	return len(staticTable) + len(table.dynamicTable)
}

func (table *IndexTable) Entry(index int) *HeaderField {
	if index < 1 || index > table.EntriesCount() {
		return nil
	}

	if index <= len(staticTable) {
		return staticTable[index-1]
	}

	return table.dynamicTable[len(table.dynamicTable)-(index-len(staticTable))]
}

func (table *IndexTable) AddEntry(hf *HeaderField) {
	table.dynamicTable = append(table.dynamicTable, hf)
	table.currentDataSize += hf.DataSize()
	table.evictEntries()
}

func (table *IndexTable) MaxProtocolDataSize() int {
	return table.maxProtocolDataSize
}

func (table *IndexTable) UpdateMaxProtocolDataSize(n int) {
	table.maxProtocolDataSize = n
	if table.maxDataSize > n {
		table.maxDataSize = n
	}
	table.evictEntries()
}

func (table *IndexTable) MaxDataSize() int {
	return table.maxDataSize
}

func (table *IndexTable) UpdateMaxDataSize(n int) error {
	if n > table.maxProtocolDataSize {
		return fmt.Errorf("%w: new max data size(%d) is over protocol limitation(%d)", ErrDataSize, n, table.maxProtocolDataSize)
	}

	table.maxDataSize = n
	table.evictEntries()
	return nil
}

// See: https://tools.ietf.org/html/rfc7541#section-4.3
//      https://tools.ietf.org/html/rfc7541#section-4.4
func (table *IndexTable) evictEntries() {
	evictUntil := 0
	for ; table.currentDataSize > table.maxDataSize; evictUntil++ {
		table.currentDataSize -= table.dynamicTable[evictUntil].DataSize()
	}

	if evictUntil > 0 {
		table.dynamicTable = table.dynamicTable[evictUntil:]
	}
}

var (
	// See: https://tools.ietf.org/html/rfc7541#appendix-A
	staticTable = []*HeaderField{
		{name: ":authority", value: ""},
		{name: ":method", value: "GET"},
		{name: ":method", value: "POST"},
		{name: ":path", value: "/"},
		{name: ":path", value: "/index.html"},
		{name: ":scheme", value: "http"},
		{name: ":scheme", value: "https"},
		{name: ":status", value: "200"},
		{name: ":status", value: "204"},
		{name: ":status", value: "206"},
		{name: ":status", value: "304"},
		{name: ":status", value: "400"},
		{name: ":status", value: "404"},
		{name: ":status", value: "500"},
		{name: "accept-charset", value: ""},
		{name: "accept-encoding", value: "gzip, deflate"},
		{name: "accept-language", value: ""},
		{name: "accept-ranges", value: ""},
		{name: "accept", value: ""},
		{name: "access-control-allow-origin", value: ""},
		{name: "age", value: ""},
		{name: "allow", value: ""},
		{name: "authorization", value: ""},
		{name: "cache-control", value: ""},
		{name: "content-disposition", value: ""},
		{name: "content-encoding", value: ""},
		{name: "content-language", value: ""},
		{name: "content-length", value: ""},
		{name: "content-location", value: ""},
		{name: "content-range", value: ""},
		{name: "content-type", value: ""},
		{name: "cookie", value: ""},
		{name: "date", value: ""},
		{name: "etag", value: ""},
		{name: "expect", value: ""},
		{name: "expires", value: ""},
		{name: "from", value: ""},
		{name: "host", value: ""},
		{name: "if-match", value: ""},
		{name: "if-modified-since", value: ""},
		{name: "if-none-match", value: ""},
		{name: "if-range", value: ""},
		{name: "if-unmodified-since", value: ""},
		{name: "last-modified", value: ""},
		{name: "link", value: ""},
		{name: "location", value: ""},
		{name: "max-forwards", value: ""},
		{name: "proxy-authenticate", value: ""},
		{name: "proxy-authorization", value: ""},
		{name: "range", value: ""},
		{name: "referer", value: ""},
		{name: "refresh", value: ""},
		{name: "retry-after", value: ""},
		{name: "server", value: ""},
		{name: "set-cookie", value: ""},
		{name: "strict-transport-security", value: ""},
		{name: "transfer-encoding", value: ""},
		{name: "user-agent", value: ""},
		{name: "vary", value: ""},
		{name: "via", value: ""},
		{name: "www-authenticate", value: ""},
	}
)
