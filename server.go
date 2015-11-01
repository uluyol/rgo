package rgo

import (
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/facebookgo/httpdown"
)

type readerDone struct {
	r    io.Reader
	done chan<- struct{}
}

type server struct {
	s    httpdown.Server
	port int

	mu   sync.Mutex
	data map[string][]byte

	fmu sync.Mutex
	fwd map[string]chan<- readerDone
}

func (s *server) putData(key string, val []byte) {
	defer s.mu.Unlock()
	s.mu.Lock()
	s.data[key] = val
}

func (s *server) rmData(key string) {
	defer s.mu.Unlock()
	s.mu.Lock()
	delete(s.data, key)
}

func (s *server) putFwd(key string, c chan<- readerDone) {
	defer s.fmu.Unlock()
	s.fmu.Lock()
	s.fwd[key] = c
}

func (s *server) rmFwd(key string) {
	defer s.fmu.Unlock()
	s.fmu.Lock()
	delete(s.fwd, key)
}

func (s *server) httpHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, "/")
	if r.Method == "GET" {
		s.mu.Lock()
		data, ok := s.data[path]
		s.mu.Unlock()
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Write(data)
		return
	} else if r.Method == "PUT" {
		s.fmu.Lock()
		c, ok := s.fwd[path]
		s.fmu.Unlock()
		if !ok {
			r.Body.Close()
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		done := make(chan struct{})
		rd := readerDone{r.Body, done}
		c <- rd
		<-done
		r.Body.Close()
		return
	}
	http.Error(w, "", http.StatusMethodNotAllowed)
}

func newServer() (*server, error) {
	var s server
	s.data = make(map[string][]byte)
	s.fwd = make(map[string]chan<- readerDone)
	hs := http.Server{
		Addr:    ":0",
		Handler: http.HandlerFunc(s.httpHandler),
	}
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	var h httpdown.HTTP
	s.s = h.Serve(&hs, ln)
	s.port = ln.Addr().(*net.TCPAddr).Port
	return &s, nil
}
