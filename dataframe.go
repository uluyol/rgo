package rgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// SimpleData is the the data type used for DataFrames.
// A string, int, float, or bool (of any size) is a
// valid value for a SimpleData variable.
type SimpleData interface{}

func ensureSimpleData(x SimpleData) {
	switch x.(type) {
	case string, bool:
	case int, int8, int16, int32, int64: // include rune and byte
	case uint, uint8, uint16, uint32, uint64, uintptr:
	case float32, float64:
	default:
		panic(fmt.Sprintf("%s is not a valid SimpleData value", x))
	}
}

// DataFrames are used to hold tabular data. This type
// can be easily marshaled into an R dataframe.
// DataFrames can only store data that meet the
// requirements set by SimpleData.
//
// This type is still experimental.
type DataFrame struct {
	cols         [][]SimpleData
	colNames     []string
	rowNames     []string
	namelessRows bool
}

// Remember to update this, MarshalJSON, UnmarshalJSON,
// SendDF, and validate when updating DataFrame.
type dataFrameJSON struct {
	Cols         [][]SimpleData `json:"cols"`
	ColNames     []string       `json:"colNames"`
	RowNames     []string       `json:"rowNames"`
	NamelessRows bool           `json:"namelessRows"`
}

func (df *DataFrame) ValidateColumns() (err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else if estr, ok := e.(string); ok {
				err = errors.New(estr)
			} else {
				panic(err)
			}
		}
	}()
	for coli, col := range df.cols {
		if len(col) == 0 {
			continue
		}

		ensureSimpleData(col[0])
		cType := reflect.TypeOf(col[0])
		for i := 1; i < len(col); i++ {
			ensureSimpleData(col[i])
			if !reflect.TypeOf(col[i]).AssignableTo(cType) {
				return fmt.Errorf("found different data types in column %d", coli)
			}
		}
	}
	return nil
}

func (df *DataFrame) MarshalJSON() ([]byte, error) {
	// don't use Key: Value syntax here so that this
	// will break if we forget to update this when
	// we update the type.
	d := dataFrameJSON{
		df.cols,
		df.colNames,
		df.rowNames,
		df.namelessRows,
	}
	return json.Marshal(&d)
}

func (df *DataFrame) UnmarshalJSON(data []byte) error {
	var d dataFrameJSON
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	df.cols = d.Cols
	df.colNames = d.ColNames
	df.rowNames = d.RowNames
	df.namelessRows = d.NamelessRows
	return nil
}

func (df *DataFrame) SetCols(colNames ...string) {
	if df.colNames != nil {
		panic("already set columns on this dataframe")
	}
	df.colNames = colNames
	df.cols = make([][]SimpleData, len(colNames))
}

func (df *DataFrame) AppendRow(name string, vals ...SimpleData) {
	if df.namelessRows {
		panic("cannot add named row: rows are nameless")
	}
	if len(vals) != len(df.cols) {
		panic("incorrect number of values being appended")
	}
	df.rowNames = append(df.rowNames, name)
	for i := range df.cols {
		df.cols[i] = append(df.cols[i], vals[i])
	}
}

func (df *DataFrame) AppendURow(vals ...SimpleData) {
	if df.rowNames != nil {
		panic("cannot add nameless row: rows are named")
	}
	if len(vals) != len(df.cols) {
		panic("incorrect number of values being appended")
	}
	for i := range df.cols {
		df.cols[i] = append(df.cols[i], vals[i])
	}
}
