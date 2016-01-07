package rgo

import "encoding/json"

// DataFrames are used to hold tabular float64 data. This
// type can be easily marshaled into an R dataframe. This
// type is still experimental.
type DataFrame struct {
	cols         [][]float64
	colNames     []string
	rowNames     []string
	namelessRows bool
}

// Remember to update this, MarshalJSON, UnmarshalJSON,
// and SendDF, when updating DataFrame.
type dataFrameJSON struct {
	Cols         [][]float64 `json:"cols"`
	ColNames     []string    `json:"colNames"`
	RowNames     []string    `json:"rowNames"`
	NamelessRows bool        `json:"namelessRows"`
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
	df.cols = make([][]float64, len(colNames))
}

func (df *DataFrame) AppendRow(name string, vals ...float64) {
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

func (df *DataFrame) AppendURow(vals ...float64) {
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
