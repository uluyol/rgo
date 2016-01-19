package dataframe

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func dfEqual(a, b *CDataFrame) error {
	if len(a.cols) != len(b.cols) {
		return errors.New("don't have the same number of cols")
	}
	if len(a.colNames) != len(b.colNames) {
		return errors.New("don't have the same number of col names")
	}
	if a.namelessRows != b.namelessRows {
		return errors.New("one has named rows and the other doesn't")
	}
	if len(a.rowNames) != len(b.rowNames) {
		return errors.New("don't have the same number of row names")
	}
	for i := range a.colNames {
		if a.colNames[i] != b.colNames[i] {
			return errors.New("don't have the same col names")
		}
	}
	for i := range a.cols {
		if a.cols[i].len() != b.cols[i].len() {
			return fmt.Errorf("different number of rows in col %d", i)
		}
		//if !reflect.DeepEqual(a.cols[i], b.cols[i]) {
		//	return fmt.Errorf("col %d entries do not match", i)
		//}
		for j := 0; j < a.cols[i].len(); j++ {
			t1 := reflect.ValueOf(a.cols[i].get(j)).Type()
			t2 := reflect.ValueOf(b.cols[i].get(j)).Type()
			if t1 != t2 {
				return fmt.Errorf("col %d row %d types differ, %s and %s", i, j, t1, t2)
			}
			if !reflect.DeepEqual(a.cols[i].get(j), b.cols[i].get(j)) {
				return fmt.Errorf("col %d row %d entries do not match", i, j)
			}
		}
	}
	for i := range a.rowNames {
		if a.rowNames[i] != b.rowNames[i] {
			return fmt.Errorf("row names for row %d do not match, got %q and %q",
				i, a.rowNames[i], b.rowNames[i])
		}
	}
	return nil
}

func TestDataFrameGet(t *testing.T) {
	testCases := []struct {
		ColNames    []string
		Rows        [][]SimpleData
		GetFunc     func(DataFrame) bool
		HasRowNames bool
	}{
		{
			ColNames:    []string{"a", "B", "cc"},
			Rows:        [][]SimpleData{{1, "x", false}, {65, "asdfasdfasdf", false}, {1, "aa", true}},
			HasRowNames: false,
			GetFunc: func(df DataFrame) bool {
				var (
					a  int
					b  string
					cc bool
				)
				df.Col("a").GetIndex(1, &a)
				df.Col("B").GetIndex(0, &b)
				df.Col("cc").GetIndex(2, &cc)

				var (
					a2  int
					b2  string
					cc2 bool
				)
				df.RowIndex(0).Get("a", &a2)
				df.RowIndex(2).Get("B", &b2)
				df.RowIndex(1).Get("cc", &cc2)

				return (a == 65 && b == "x" && cc == true &&
					a2 == 1 && b2 == "aa" && cc2 == false)
			},
		},
		{
			ColNames:    []string{"234234", "123123f"},
			Rows:        [][]SimpleData{{"asdf", 2.5, 4}, {"222", float64(2), 1}},
			HasRowNames: true,
			GetFunc: func(df DataFrame) bool {
				var first float64
				var second int
				df.Col("234234").Get("222", &first)
				df.Col("123123f").Get("asdf", &second)

				var first2 float64
				var second2 int
				df.Row("asdf").GetIndex(0, &first2)
				df.Row("222").GetIndex(1, &second2)

				return (first == 2 && second == 4 &&
					first2 == 2.5 && second2 == 1)
			},
		},
	}
	for caseN, c := range testCases {
		var df CDataFrame
		df.SetCols(c.ColNames...)
		if c.HasRowNames {
			for _, r := range c.Rows {
				df.AppendRow(r[0].(string), r[1:]...)
			}
		} else {
			for _, r := range c.Rows {
				df.AppendURow(r...)
			}
		}
		if err := df.ValidateColumns(); err != nil {
			t.Errorf("case %d: failed to validate data frame: %v", caseN, err)
		}
		if !c.GetFunc(&df) {
			t.Errorf("case %d: get returned incorrect values", caseN)
		}
	}
}

func TestDataFrameJSON(t *testing.T) {
	testCases := []struct {
		ColNames    []string
		Rows        [][]SimpleData
		HasRowNames bool
	}{
		{
			ColNames: []string{"anjsadkjn", "2341234123$", "	2341"},
			Rows:        [][]SimpleData{{123, 123, 123.0}, {90, 23, 11.1}},
			HasRowNames: false,
		},
		{
			ColNames:    []string{"123123", "asdfadsf123123", "___"},
			Rows:        [][]SimpleData{{"asdf", "asdf", 123, false}, {"00000", "wert", 12111, true}, {"-", "+", -1, true}},
			HasRowNames: true,
		},
	}
	for caseN, c := range testCases {
		var df CDataFrame
		df.SetCols(c.ColNames...)
		if c.HasRowNames {
			for _, r := range c.Rows {
				df.AppendRow(r[0].(string), r[1:]...)
			}
		} else {
			for _, r := range c.Rows {
				df.AppendURow(r...)
			}
		}
		if err := df.ValidateColumns(); err != nil {
			t.Errorf("case %d: failed to validate data frame: %v", caseN, err)
		}
		b, err := json.Marshal(&df)
		if err != nil {
			t.Errorf("case %d: error while marshalling: %v", caseN, err)
		}
		var df2 CDataFrame
		if err := json.Unmarshal(b, &df2); err != nil {
			t.Errorf("case %d: error while unmarshaling: %v", caseN, err)
		}
		if err := dfEqual(&df, &df2); err != nil {
			t.Errorf("case %d: data frames do not equal: %v", caseN, err)
			t.Logf("case %d: %+v %+v", caseN, df.cols, df2.cols)
		}
	}
}
