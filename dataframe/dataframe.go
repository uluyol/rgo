package dataframe

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

//go:generate python3 gen_simple_data.py simple_data_util.generated.go other=string,bool numeric=int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,uintptr,float32,float64

// SimpleData is the the data type used for DataFrames.
// A string, int, float, or bool (of any size) is a
// valid value for a SimpleData variable.
type SimpleData interface{}

// Row is an immutable copy of a row of DataFrame.
// As it is a copy, it will not reflect changes
// later made to the DataFrame.
//
// This type is still experimental.
type Row struct {
	colNames []string
	v        []SimpleData
}

// Get will return the value of the given column.
func (r Row) Get(colName string, x SimpleData) {
	for i, name := range r.colNames {
		if name == colName {
			r.GetIndex(i, x)
			return
		}
	}
	panic("unable to find column")
}

// GetIndex will return the value of the given column.
func (r Row) GetIndex(i int, x SimpleData) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		panic("x needs to be a pointer or interface")
	}
	v2 := reflect.ValueOf(r.v[i])
	e := v.Elem()
	if !v2.Type().AssignableTo(e.Type()) {
		panic("cannot assign the value to x: wrong type")
	}
	e.Set(v2)
}

// GetSD will return the value of the given column as a
// SimpleData. Users will typically use Get instead of this.
func (r Row) GetSD(colName string) SimpleData {
	for i, name := range r.colNames {
		if name == colName {
			return r.GetIndexSD(i)
		}
	}
	panic("unable to find column")
}

// GetIndexSD will return the value of the given column as a
// SimpleData. Users will typically use GetIndex instead of this.
func (r Row) GetIndexSD(i int) SimpleData {
	return r.v[i]
}

// Column is an immutable view of a column of a
// DataFrame. It remains up-to-date even as the
// values in the DataFrame are changed.
//
// This type is still experimental.
type Column struct {
	v            *[]SimpleData
	rowNames     *[]string
	rowNameIndex map[string]int
}

// Get will return the value of the given row.
func (c Column) Get(rowName string, x SimpleData) {
	if c.rowNameIndex != nil {
		if i, ok := c.rowNameIndex[rowName]; ok {
			c.GetIndex(i, x)
			return
		}
	}
	for i, name := range *c.rowNames {
		if name == rowName {
			c.GetIndex(i, x)
			return
		}
	}
	panic("unable to find row")
}

// GetIndex will return the value of the given row.
func (c Column) GetIndex(i int, x SimpleData) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		panic("x needs to be a pointer or interface")
	}
	v2 := reflect.ValueOf((*c.v)[i])
	e := v.Elem()
	if !v2.Type().AssignableTo(e.Type()) {
		panic("cannot assign the value to x: wrong type")
	}
	e.Set(v2)
}

// GetSD will return the value of the given row as a SimpleData.
// Users will typically call Get instead of this.
func (c Column) GetSD(rowName string) SimpleData {
	if c.rowNameIndex != nil {
		if i, ok := c.rowNameIndex[rowName]; ok {
			return c.GetIndexSD(i)
		}
	}
	for i, name := range *c.rowNames {
		if name == rowName {
			return c.GetIndexSD(i)
		}
	}
	panic("unable to find row")
}

// GetIndexSD will return the value of the given row as a
// SimpleData. Users will typically call GetIndex instead of this.
func (c Column) GetIndexSD(i int) SimpleData {
	return (*c.v)[i]
}

func (c Column) Len() int {
	return len(*c.v)
}

func (c Column) MarshalJSON() ([]byte, error) {
	return json.Marshal(*c.v)
}

type dfColumn struct {
	v *[]SimpleData
}

func (c dfColumn) get(i int) SimpleData    { return (*c.v)[i] }
func (c dfColumn) set(i int, v SimpleData) { (*c.v)[i] = v }
func (c dfColumn) len() int                { return len(*c.v) }

func (c dfColumn) append(v SimpleData) dfColumn {
	sl := append(*c.v, v)
	return dfColumn{&sl}
}

func (c dfColumn) String() string {
	return fmt.Sprint(*c.v)
}

func (c dfColumn) dtype() string {
	if len(*c.v) == 0 {
		return "empty"
	}
	return reflect.TypeOf((*c.v)[0]).Kind().String()
}

func (c dfColumn) MarshalJSON() ([]byte, error) {
	s := struct {
		D string       `json:"dtype"`
		V []SimpleData `json:"val"`
	}{
		c.dtype(),
		*c.v,
	}
	return json.Marshal(s)
}

func (c *dfColumn) UnmarshalJSON(data []byte) error {
	var dtype struct {
		D string `json:"dtype"`
	}
	if err := json.Unmarshal(data, &dtype); err != nil {
		return err
	}
	v, err := slicePtrOf(dtype.D)
	if err != nil {
		return err
	}
	wrapper := struct {
		V interface{} `json:"val"`
	}{v}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	vv := reflect.ValueOf(v)
	sdv := make([]SimpleData, vv.Elem().Len())
	for i := 0; i < vv.Elem().Len(); i++ {
		sdv[i] = vv.Elem().Index(i).Interface()
	}
	c.v = &sdv
	return nil
}

func newDFColumn() dfColumn {
	v := make([]SimpleData, 0, 64)
	return dfColumn{&v}
}

// DataFrame is used to hold tabular data. This type
// can be easily marshaled into an R dataframe.
// DataFrames can only store data that meet the
// requirements set by SimpleData.
//
// DataFrames store data by column not by row. As a
// result, column-oriented operations can be completed
// with fewer copies than row-oriented ones.
//
// DataFrames (and associated types) are not thread-safe.
//
// This type is still experimental.
type DataFrame struct {
	cols         []dfColumn
	colNames     []string
	rowNames     []string
	rowNameIndex map[string]int
	namelessRows bool
}

// New creates a new DataFrame with the provided column names.
func New(colNames ...string) *DataFrame {
	var df DataFrame
	df.SetCols(colNames...)
	return &df
}

// Remember to update this, MarshalJSON, UnmarshalJSON,
// SendDF, and validate when updating DataFrame.
type dataFrameJSON struct {
	Cols         []dfColumn `json:"cols"`
	ColNames     []string   `json:"colNames"`
	RowNames     []string   `json:"rowNames"`
	NamelessRows bool       `json:"namelessRows"`
}

// ValidateColumns checks that columns are composed of
// identical types.
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
	for coli, scol := range df.cols {
		col := *scol.v
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

// ColNames returns a slice of the column names.
func (df *DataFrame) ColNames() []string {
	return df.colNames
}

// RowNames returns a slice of the row names and a bool
// indicating whether the rows are named.
func (df *DataFrame) RowNames() (rowNames []string, hasNamedRows bool) {
	return df.rowNames, !df.namelessRows
}

// Col gets the column of the provided name. This operation
// does not copy any data.
//
// Col will panic if the colName does not exist.
func (df *DataFrame) Col(colName string) Column {
	for i, name := range df.colNames {
		if name == colName {
			return df.ColIndex(i)
		}
	}
	panic("unable to find column")
}

// ColIndex gets the column of the provided index. This
// operation does not copy any data.
//
// ColIndex will panic if the index is out of bounds.
func (df *DataFrame) ColIndex(i int) Column {
	return Column{
		v:            df.cols[i].v,
		rowNames:     &df.rowNames,
		rowNameIndex: df.rowNameIndex,
	}
}

// Row gets the row of the provided name. This operation
// does copy data. Use Col() over Row() whenever possible.
//
// Row will panic if the rowName does not exist.
func (df *DataFrame) Row(rowName string) Row {
	if df.rowNameIndex != nil {
		if i, ok := df.rowNameIndex[rowName]; ok {
			return df.RowIndex(i)
		}
	}
	for i, name := range df.rowNames {
		if name == rowName {
			return df.RowIndex(i)
		}
	}
	panic("unable to find row")
}

// RowIndex gets the row of the provided index. This
// operation does copy data. Use ColIndex() over
// RowIndex() whenever possible.
//
// RowIndex will panic if the index is out of bounds.
func (df *DataFrame) RowIndex(i int) Row {
	vals := make([]SimpleData, len(df.cols))
	for j := range df.cols {
		vals[j] = df.cols[j].get(i)
	}
	return Row{
		v:        vals,
		colNames: df.colNames,
	}
}

// SetCols sets the columns of the DataFrame to have
// the given names. You may not call this method more
// than once.
//
// Do not call this method if you are using NewDataFrame.
func (df *DataFrame) SetCols(colNames ...string) {
	if df.colNames != nil {
		panic("already set columns on this dataframe")
	}
	df.colNames = colNames
	df.cols = make([]dfColumn, len(colNames))
	for i := range df.cols {
		df.cols[i] = newDFColumn()
	}
}

// FastRowLookups increases row name book-keeping to
// increase the performance of row name lookups.
//
// For maximum benefit, enable before adding any rows.
func (df *DataFrame) FastRowLookups(enable bool) {
	if !enable {
		df.rowNameIndex = nil
	}
	if df.rowNameIndex == nil {
		df.rowNameIndex = make(map[string]int)
	}
}

// AppendRow adds a new row of data to the DataFrame with the
// given name. If you do not want named rows, use AppendURow
// instead.
func (df *DataFrame) AppendRow(name string, vals ...SimpleData) {
	if df.namelessRows {
		panic("cannot add named row: rows are nameless")
	}
	if len(vals) != len(df.cols) {
		panic("incorrect number of values being appended")
	}
	df.rowNames = append(df.rowNames, name)
	if df.rowNameIndex != nil {
		df.rowNameIndex[name] = len(df.rowNames)
	}
	for i := range df.cols {
		df.cols[i] = df.cols[i].append(vals[i])
	}
}

// AppendURow adds a new row of data to the DataFrame without
// a name. If you do want named rows, use AppendRow instead.
func (df *DataFrame) AppendURow(vals ...SimpleData) {
	if df.rowNames != nil {
		panic("cannot add nameless row: rows are named")
	}
	if len(vals) != len(df.cols) {
		panic("incorrect number of values being appended")
	}
	for i := range df.cols {
		df.cols[i] = df.cols[i].append(vals[i])
	}
}
