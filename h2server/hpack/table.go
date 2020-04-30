package hpack

type (
	// DynamicTable is representation of "Dynamic Table"
	// See: https://tools.ietf.org/html/rfc7541#section-2.3.2
	DynamicTable struct {
		entries       []*HeaderField
		maxDataSize   int
		totalDataSize int
	}

	// IndexAddressSpace is representation of "Index Address Space"
	// See: https://tools.ietf.org/html/rfc7541#section-2.3.3
	IndexAddressSpace struct {
		dynamic *DynamicTable
	}
)

func NewIndexAddressSpace(maxDynTableSize int) *IndexAddressSpace {
	return &IndexAddressSpace{
		dynamic: &DynamicTable{
			entries:       []*HeaderField{},
			maxDataSize:   maxDynTableSize,
			totalDataSize: 0,
		},
	}
}

func (ias *IndexAddressSpace) Entry(index int) *HeaderField {
	if index < 1 || index > ias.IndexSize() {
		return nil
	}

	if index <= len(staticTable) {
		return staticTable[index-1]
	}

	return ias.dynamic.entry(index - len(staticTable))
}

func (ias *IndexAddressSpace) IndexSize() int {
	return len(staticTable) + ias.dynamic.entrySize()
}

func (ias *IndexAddressSpace) Dynamic() *DynamicTable {
	return ias.dynamic
}

func (dyn *DynamicTable) AddEntry(entry *HeaderField) {
	dyn.totalDataSize += entry.DataSize()
	dyn.entries = append(dyn.entries, entry)
	dyn.trimEntries()
}

func (dyn *DynamicTable) MaxDataSize() int {
	return dyn.maxDataSize
}

func (dyn *DynamicTable) UpdateMaxDataSize(maxDataSize int) {
	dyn.maxDataSize = maxDataSize
	dyn.trimEntries()
}

func (dyn *DynamicTable) entry(index int) *HeaderField {
	if index < 1 || index > dyn.entrySize() {
		return nil
	}

	return dyn.entries[dyn.entrySize()-index]
}

func (dyn *DynamicTable) entrySize() int {
	return len(dyn.entries)
}

func (dyn *DynamicTable) trimEntries() {
	if dyn.totalDataSize <= dyn.maxDataSize {
		return
	}

	total := dyn.totalDataSize
	for i := 0; i < len(dyn.entries); i++ {
		total -= dyn.entries[i].DataSize()
		if total <= dyn.maxDataSize {
			dyn.entries = dyn.entries[i+1:]
			dyn.totalDataSize = total
		}
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
