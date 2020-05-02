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
		{name: []byte(":authority"), value: nil},
		{name: []byte(":method"), value: []byte("GET")},
		{name: []byte(":method"), value: []byte("POST")},
		{name: []byte(":path"), value: []byte("/")},
		{name: []byte(":path"), value: []byte("/index.html")},
		{name: []byte(":scheme"), value: []byte("http")},
		{name: []byte(":scheme"), value: []byte("https")},
		{name: []byte(":status"), value: []byte("200")},
		{name: []byte(":status"), value: []byte("204")},
		{name: []byte(":status"), value: []byte("206")},
		{name: []byte(":status"), value: []byte("304")},
		{name: []byte(":status"), value: []byte("400")},
		{name: []byte(":status"), value: []byte("404")},
		{name: []byte(":status"), value: []byte("500")},
		{name: []byte("accept-charset"), value: nil},
		{name: []byte("accept-encoding"), value: []byte("gzip, deflate")},
		{name: []byte("accept-language"), value: nil},
		{name: []byte("accept-ranges"), value: nil},
		{name: []byte("accept"), value: nil},
		{name: []byte("access-control-allow-origin"), value: nil},
		{name: []byte("age"), value: nil},
		{name: []byte("allow"), value: nil},
		{name: []byte("authorization"), value: nil},
		{name: []byte("cache-control"), value: nil},
		{name: []byte("content-disposition"), value: nil},
		{name: []byte("content-encoding"), value: nil},
		{name: []byte("content-language"), value: nil},
		{name: []byte("content-length"), value: nil},
		{name: []byte("content-location"), value: nil},
		{name: []byte("content-range"), value: nil},
		{name: []byte("content-type"), value: nil},
		{name: []byte("cookie"), value: nil},
		{name: []byte("date"), value: nil},
		{name: []byte("etag"), value: nil},
		{name: []byte("expect"), value: nil},
		{name: []byte("expires"), value: nil},
		{name: []byte("from"), value: nil},
		{name: []byte("host"), value: nil},
		{name: []byte("if-match"), value: nil},
		{name: []byte("if-modified-since"), value: nil},
		{name: []byte("if-none-match"), value: nil},
		{name: []byte("if-range"), value: nil},
		{name: []byte("if-unmodified-since"), value: nil},
		{name: []byte("last-modified"), value: nil},
		{name: []byte("link"), value: nil},
		{name: []byte("location"), value: nil},
		{name: []byte("max-forwards"), value: nil},
		{name: []byte("proxy-authenticate"), value: nil},
		{name: []byte("proxy-authorization"), value: nil},
		{name: []byte("range"), value: nil},
		{name: []byte("referer"), value: nil},
		{name: []byte("refresh"), value: nil},
		{name: []byte("retry-after"), value: nil},
		{name: []byte("server"), value: nil},
		{name: []byte("set-cookie"), value: nil},
		{name: []byte("strict-transport-security"), value: nil},
		{name: []byte("transfer-encoding"), value: nil},
		{name: []byte("user-agent"), value: nil},
		{name: []byte("vary"), value: nil},
		{name: []byte("via"), value: nil},
		{name: []byte("www-authenticate"), value: nil},
	}
)
