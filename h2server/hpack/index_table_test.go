package hpack

import (
	"errors"
	"reflect"
	"testing"
)

func TestIndexTable_EntriesCount(t *testing.T) {
	staticTableLen := len(staticTable)

	tests := []struct {
		name  string
		table func() *IndexTable
		want  int
	}{
		{
			name: "Default",
			table: func() *IndexTable {
				return NewIndexTable(4096)
			},
			want: staticTableLen,
		},
		{
			name: "Dynamic table has 1 entry",
			table: func() *IndexTable {
				table := NewIndexTable(4096)
				table.AddEntry(&HeaderField{name: []byte("foo"), value: []byte("bar")})
				return table
			},
			want: staticTableLen + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.table().EntriesCount()
			if got != tt.want {
				t.Errorf("EntriesCount() got = %d, want = %d", got, tt.want)
			}
		})
	}
}

func TestIndexTable_Entry(t *testing.T) {
	tests := []struct {
		name  string
		table func() *IndexTable
		in    int
		want  *HeaderField
	}{
		{
			name: "Index:1 for default index table",
			table: func() *IndexTable {
				return NewIndexTable(4096)
			},
			in:   1,
			want: &HeaderField{name: []byte(":authority"), value: nil},
		},
		{
			name: "Index:61 for default index table",
			table: func() *IndexTable {
				return NewIndexTable(4096)
			},
			in:   61,
			want: &HeaderField{name: []byte("www-authenticate"), value: nil},
		},
		{
			name: "Index:0 for default index table",
			table: func() *IndexTable {
				return NewIndexTable(4096)
			},
			in:   0,
			want: nil,
		},
		{
			name: "Index:62 for default index table",
			table: func() *IndexTable {
				return NewIndexTable(4096)
			},
			in:   62,
			want: nil,
		},
		{
			name: "Index:62 for index table with dynamic table contains 3 entries",
			table: func() *IndexTable {
				table := NewIndexTable(4096)
				table.AddEntry(&HeaderField{name: []byte("foo"), value: []byte("bar")})
				table.AddEntry(&HeaderField{name: []byte("x-forwarded-for"), value: []byte("192.168.0.1")})
				table.AddEntry(&HeaderField{name: []byte("x-frame-options"), value: []byte("deny")})
				return table
			},
			in:   62,
			want: &HeaderField{name: []byte("x-frame-options"), value: []byte("deny")},
		},
		{
			name: "Index:64 for index table with dynamic table contains 3 entries",
			table: func() *IndexTable {
				table := NewIndexTable(4096)
				table.AddEntry(&HeaderField{name: []byte("foo"), value: []byte("bar")})
				table.AddEntry(&HeaderField{name: []byte("x-forwarded-for"), value: []byte("192.168.0.1")})
				table.AddEntry(&HeaderField{name: []byte("x-frame-options"), value: []byte("deny")})
				return table
			},
			in:   64,
			want: &HeaderField{name: []byte("foo"), value: []byte("bar")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.table().Entry(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Entry() got = %+v, want = %+v", got, tt.want)
			}
		})
	}
}

func TestIndexTable_AddEntry(t *testing.T) {
	tests := []struct {
		table func() *IndexTable
		in    *HeaderField
		want  []*HeaderField
	}{
		{
			table: func() *IndexTable {
				return NewIndexTable(100)
			},
			in: &HeaderField{name: []byte("foo"), value: []byte("bar")},
			want: []*HeaderField{
				{name: []byte("foo"), value: []byte("bar")},
			},
		},
		{
			table: func() *IndexTable {
				table := NewIndexTable(100)

				// The data size of this header field is 50 bytes(name = 9 bytes, value = 9 bytes, implicit overhead = 32 bytes)
				table.AddEntry(&HeaderField{name: []byte("111111111"), value: []byte("111111111")})
				table.AddEntry(&HeaderField{name: []byte("222222222"), value: []byte("222222222")})
				return table
			},
			in: &HeaderField{name: []byte("333333333"), value: []byte("333333333")},
			want: []*HeaderField{
				{name: []byte("333333333"), value: []byte("333333333")},
				{name: []byte("222222222"), value: []byte("222222222")},
			},
		},
	}

	for _, tt := range tests {
		table := tt.table()
		table.AddEntry(tt.in)
		testDynamicTableEntries(t, table, tt.want)
	}
}

func TestIndexTable_UpdateMaxProtocolDataSize(t *testing.T) {
	table := NewIndexTable(100)
	table.AddEntry(&HeaderField{name: []byte("111111111"), value: []byte("111111111")})
	table.AddEntry(&HeaderField{name: []byte("222222222"), value: []byte("222222222")})

	table.UpdateMaxProtocolDataSize(50)

	if table.MaxDataSize() != 50 {
		t.Errorf("UpdateMaxProtocolDataSize() update max data size = %d, want = %d", table.MaxDataSize(), 50)
	}

	testDynamicTableEntries(t, table, []*HeaderField{
		{name: []byte("222222222"), value: []byte("222222222")},
	})
}

func TestIndexTable_UpdateMaxDataSize(t *testing.T) {
	type want struct {
		dynamicTable []*HeaderField
		err          error
	}

	tests := []struct {
		name  string
		table func() *IndexTable
		in    int
		want  want
	}{
		{
			name: "Valid max data size",
			table: func() *IndexTable {
				table := NewIndexTable(100)
				table.AddEntry(&HeaderField{name: []byte("111111111"), value: []byte("111111111")})
				table.AddEntry(&HeaderField{name: []byte("222222222"), value: []byte("222222222")})
				return table
			},
			in: 50,
			want: want{
				dynamicTable: []*HeaderField{
					{name: []byte("222222222"), value: []byte("222222222")},
				},
				err: nil,
			},
		},
		{
			name: "Invalid max data size",
			table: func() *IndexTable {
				return NewIndexTable(100)
			},
			in: 101,
			want: want{
				err: ErrDataSize,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := tt.table()
			gotErr := table.UpdateMaxDataSize(tt.in)
			if !errors.Is(gotErr, tt.want.err) {
				t.Errorf("UpdateMaxDataSize() return error = %+v, want = %+v", gotErr, tt.want.err)
				return
			}
			if gotErr != nil {
				return
			}

			testDynamicTableEntries(t, table, tt.want.dynamicTable)
		})
	}
}

func testDynamicTableEntries(t *testing.T, table *IndexTable, expect []*HeaderField) {
	t.Helper()

	staticCount := len(staticTable)
	wantCount := staticCount + len(expect)
	if table.EntriesCount() != wantCount {
		t.Errorf("unexpected index table entries count = %d, want = %d", table.EntriesCount(), wantCount)
		return
	}

	for i, wantEntry := range expect {
		index := staticCount + i + 1
		gotEntry := table.Entry(index)
		if !reflect.DeepEqual(gotEntry, wantEntry) {
			t.Errorf("unexpected entry in index table at %d = %+v, want = %+v", index, gotEntry, wantEntry)
			return
		}
	}
}
