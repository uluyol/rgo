package rgo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func startTestServer(t *testing.T) *server {
	s, err := newServer()
	if err != nil {
		t.Errorf("failed to start server: %v", err)
	}
	return s
}

func TestServerShutdown(t *testing.T) {
	s := startTestServer(t)
	err := s.s.Stop()
	if err != nil {
		t.Errorf("error after stopping server: %v", err)
	}
}

func TestServerData(t *testing.T) {
	s := startTestServer(t)
	s.putData("hello", []byte("world"))
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/hello", s.port))
	if err != nil {
		t.Errorf("failed to get 'hello': %v", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Errorf("error while reading data: %v", err)
	}
	if string(b) != "world" {
		t.Errorf("expected %q, got %q", "world", b)
	}

	s.rmData("hello")
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/hello", s.port))
	if err != nil {
		t.Errorf("unexpected error requesting 'hello': %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-OK status after removing 'hello'")
	}
	s.s.Stop()
}

func TestServerFwd(t *testing.T) {
	data := "this #@ THE_data"
	s := startTestServer(t)
	rch := make(chan readerDone)
	s.putFwd("abc.xyz", rch)
	done := make(chan struct{})
	go func() {
		rd := <-rch
		b, err := ioutil.ReadAll(rd.r)
		close(rd.done)
		if err != nil {
			t.Errorf("error while reading body: %v", err)
		}
		if string(b) != data {
			t.Errorf("expected %q, got %q", data, b)
		}
		close(done)
	}()
	var client http.Client
	var url = fmt.Sprintf("http://localhost:%d/abc.xyz", s.port)
	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
	if err != nil {
		t.Errorf("unexpected error building request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("unexpected error PUT-ing data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status PUT-ing data: %v", resp.Status)
	}
	<-done

	s.rmFwd("abc.xyz")
	req, err = http.NewRequest("PUT", url, bytes.NewBufferString(data))
	if err != nil {
		t.Errorf("unexpected error building request: %v", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("unexpected error PUT-ing data: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unexpected status PUT-ing data: %v", resp.Status)
	}

	s.s.Stop()
}
