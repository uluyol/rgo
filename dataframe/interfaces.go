package dataframe

import "encoding/json"

type DataFrame interface {
	AppendRow(name string, vals ...SimpleData)
	AppendURow(vals ...SimpleData)
	Col(colName string) Column
	ColIndex(i int) Column
	ColNames() []string
	Row(rowName string) Row
	RowIndex(i int) Row
	RowNames() (names []string, hasRowNames bool)
}

// Column is an immutable view of a column of a
// DataFrame. It remains up-to-date even as the
// values in the DataFrame are changed.
//
// This type is still experimental.
type Column interface {
	json.Marshaler
	Get(rowName string, x SimpleData)
	GetIndex(i int, x SimpleData)
	GetIndexSD(i int) SimpleData
	GetSD(rowName string) SimpleData
	Len() int
}

// Row is an immutable copy of a row of DataFrame.
// As it is a copy, it will not reflect changes
// later made to the DataFrame.
//
// This type is still experimental.
type Row interface {
	Get(colName string, x SimpleData)
	GetIndex(i int, x SimpleData)
	GetIndexSD(i int) SimpleData
	GetSD(colName string) SimpleData
}
