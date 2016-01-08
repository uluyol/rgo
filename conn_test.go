package rgo

import "testing"

func newTestConn(t *testing.T) *Conn {
	var c *Conn
	var err error
	if testing.Verbose() {
		c, err = Connection(WithDebug())
	} else {
		c, err = Connection()
	}
	if err != nil {
		t.Errorf("failed to create connection: %v", err)
	}
	if c == nil {
		t.Error("got nil connection")
	}
	return c
}

func TestCreateCloseConn(t *testing.T) {
	c := newTestConn(t)
	err := c.Close()
	if err != nil {
		t.Errorf("failed to close connection")
	}
}

func TestConn(t *testing.T) {
	c := newTestConn(t)

	data := []float64{4, 5, 6, 7, 8, 9}

	err := c.Send(data, "mydata")
	if err != nil {
		t.Errorf("unexpected error sending data: %v", err)
	}
	err = c.R("mydata = mydata + 1")
	if err != nil {
		t.Errorf("unexpected error printing data: %v", err)
	}
	var newdata []float64
	err = c.Get(&newdata, "mydata")
	if err != nil {
		t.Errorf("couldn't get 'mydata': %v", err)
	}
	for i := range data {
		if data[i]+1 != newdata[i] {
			t.Errorf("expected %v, got %v", data, newdata)
			break
		}
	}

	c.Close()
}

func TestConnStrict(t *testing.T) {
	c := newTestConn(t)
	defer c.Close()

	if err := c.Strict(); err != nil {
		t.Errorf("unexpected error setting strict mode: %v", err)
	}
	err := c.R("warning('hi')")
	if err == nil {
		t.Errorf("expected error")
	}
	if IsWarning(err) {
		t.Errorf("got warning instead of error")
	}
}

func TestSendDFTypes(t *testing.T) {
	testCases := []struct {
		ColNames    []string
		Rows        [][]SimpleData
		Types       []string
		HasRowNames bool
	}{
		{
			ColNames:    []string{"a", "B"},
			Rows:        [][]SimpleData{{1.0, "x"}, {65.0, "asdfasdfasdf"}, {1.0, "aa"}},
			Types:       []string{"double", "integer"},
			HasRowNames: false,
		},
		{
			ColNames:    []string{"234234", "123123f"},
			Rows:        [][]SimpleData{{"asdf", 2.5, 4}, {"222", float64(2), 1}},
			Types:       []string{"double", "double"},
			HasRowNames: true,
		},
	}
	rc := newTestConn(t)
	defer rc.Close()
	for caseN, c := range testCases {
		t.Logf("case %d: building data frame", caseN)
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
		t.Logf("case %d: validating data frame", caseN)
		if err := df.ValidateColumns(); err != nil {
			t.Errorf("case %d: failed to validate data frame: %v", caseN, err)
		}
		t.Logf("case %d: sending data frame", caseN)
		if err := rc.SendDF(&df, "data"); err != nil {
			t.Fatalf("case %d: error sending data frame: %v", caseN, err)
		}
		rc.R("print(data)")
		rc.R("print(typeof(..rgo.df.cols.0))")
		t.Logf("case %d: checking types of data frame columns", caseN)
		for i, cname := range c.ColNames {
			if err := rc.Rf("coltype <- typeof(data[[%q]])", cname); err != nil {
				t.Fatalf("case %d: error assigning type of %q column to R var: %v", caseN, cname, err)
			}
			var colType []string
			if err := rc.Get(&colType, "coltype"); err != nil {
				t.Errorf("case %d: error getting type of %q column from R: %v", caseN, cname, err)
			}
			if colType[0] != c.Types[i] {
				t.Errorf("case %d: %q column: expected type %q, got %q", caseN, cname, c.Types[i], colType)
			}
		}
	}
}
