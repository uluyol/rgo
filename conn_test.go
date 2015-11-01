package rgo

import "testing"

func newTestConn(t *testing.T) *Conn {
	c, err := Connection()
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
