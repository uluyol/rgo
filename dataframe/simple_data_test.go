package dataframe

import "testing"

type simpleComposite struct {
	a int
	b bool
}

func TestIsNumeric(t *testing.T) {
	good := []SimpleData{
		int(0),
		int8(1),
		int32(0xff),
		int64(-44),
		uint(99),
		uint8(255),
		uint16(11),
		uint32(8245),
		uint64(1231231299),
		uintptr(123),
		float32(0.523),
		float64(-1123.1231),
	}
	bad := []SimpleData{
		"asdfasdf",
		simpleComposite{},
		struct{}{},
	}
	for _, sd := range good {
		if !IsNumeric(sd) {
			t.Errorf("expected numeric but claimed otherwise: %v", sd)
		}
	}
	for _, sd := range bad {
		if IsNumeric(sd) {
			t.Errorf("expected non-numeric but claimed otherwise: %v", sd)
		}
	}
}
