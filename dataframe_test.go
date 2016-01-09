package rgo

import "testing"

func TestDataFrameGet(t *testing.T) {
	testCases := []struct {
		ColNames    []string
		Rows        [][]SimpleData
		GetFunc     func(*DataFrame) bool
		HasRowNames bool
	}{
		{
			ColNames:    []string{"a", "B", "cc"},
			Rows:        [][]SimpleData{{1, "x", false}, {65, "asdfasdfasdf", false}, {1, "aa", true}},
			HasRowNames: false,
			GetFunc: func(df *DataFrame) bool {
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
			GetFunc: func(df *DataFrame) bool {
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
		var df DataFrame
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
