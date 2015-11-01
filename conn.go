package rgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Conn struct {
	cmd     *exec.Cmd
	inPipe  io.WriteCloser
	counter uint64
	server  *server
}

func (c *Conn) start() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	return err
}

type connConfig struct {
	debug bool
}

type ConnOption func(*connConfig)

func WithDebug() ConnOption {
	return func(c *connConfig) {
		c.debug = true
	}
}

const checkDepsCmd = "cat(is.element(\"jsonlite\", installed.packages()[,1]) & is.element(\"RCurl\", installed.packages()[,1]))\n"

func Connection(opts ...ConnOption) (*Conn, error) {
	var cfg connConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	var c Conn
	out, err := exec.Command("R", "--no-save", "-s", "-e", checkDepsCmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %v", err)
	}
	if string(out) != "TRUE" {
		fmt.Printf("got: %s", out)
		return nil, errors.New("need to install 'jsonlite' and 'RCurl'")
	}
	c.cmd = exec.Command("R", "--no-save")
	c.inPipe, err = c.cmd.StdinPipe()
	if cfg.debug {
		c.cmd.Stdout = os.Stdout
		c.cmd.Stderr = os.Stderr
	}
	if err != nil {
		return nil, err
	}
	err = c.start()
	if err != nil {
		return nil, err
	}
	if err != nil {
		c.Close()
		return nil, err
	}
	c.server, err = newServer()
	if err != nil {
		goto ErrCleanup
	}
	err = c.directR("library(jsonlite)\n")
	if err != nil {
		goto ErrCleanup
	}
	err = c.directR("library(RCurl)\n")
	if err != nil {
		goto ErrCleanup
	}
	return &c, nil

ErrCleanup:
	c.Close()
	return nil, err
}

func (c *Conn) Close() error {
	c.directR("q()")
	c.inPipe.Close()
	err1 := c.cmd.Wait()
	err2 := c.server.s.Stop()
	if err1 != nil {
		return err1
	}
	return err2
}

// directR sends the command to R without trapping errors.
// directR should only be used for registring some internal functions.
func (c *Conn) directR(cmd string) error {
	_, err := io.WriteString(c.inPipe, cmd)
	return err
}

const cmdStr = `..rgo.ret = c()
tryCatch({
	%s
}, warning = function(w) {
	..rgo.ret["Warning"] <<- conditionMessage(w)
}, error = function(e) {
	..rgo.ret["Error"] <<- conditionMessage(e)
})
print(..rgo.ret)
httpPUT("http://localhost:%d/%s", toJSON(..rgo.ret))
`

type res struct {
	Error   string
	Warning string
}

type rError string

func (e rError) Error() string { return string(e) }
func (e rError) IsError()      {}

type rWarning string

func (w rWarning) Error() string { return string(w) }
func (w rWarning) IsWarning()    {}

func (r res) toError() error {
	if r.Error != "" {
		return rError(r.Error)
	} else if r.Warning != "" {
		return rWarning(r.Warning)
	}
	return nil
}

func (c *Conn) R(cmd string) error {
	key := "r.result"
	rch := make(chan readerDone)
	c.server.putFwd(key, rch)
	defer c.server.rmFwd(key)
	fmt.Fprintf(c.inPipe, cmdStr, cmd, c.server.port, key)
	rd := <-rch
	defer close(rd.done)
	dec := json.NewDecoder(rd.r)
	var result res
	err := dec.Decode(&result)
	if err != nil {
		return fmt.Errorf("error while decoding: %v", err)
	}
	return result.toError()
}

func (c *Conn) Rf(format string, args ...interface{}) error {
	return c.R(fmt.Sprintf(format, args...))
}

func (c *Conn) getuid() uint64 {
	x := c.counter
	c.counter++
	return x
}

func (c *Conn) write(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("go.data.%d", c.getuid())
	c.server.putData(key, b)
	return key, nil
}

func (c *Conn) Send(data interface{}, name string) error {
	key, err := c.write(data)
	if key != "" {
		defer c.server.rmData(key)
	}
	if err != nil {
		return err
	}
	return c.Rf("%s = fromJSON(getURL(\"http://localhost:%d/%s\"))", name, c.server.port, key)
}

func (c *Conn) Get(data interface{}, name string) error {
	key := fmt.Sprintf("r.data.%d", c.getuid())
	rch := make(chan readerDone)
	c.server.putFwd(key, rch)
	defer c.server.rmFwd(key)

	errCh := make(chan error)
	go func() {
		errCh <- c.Rf("httpPUT(\"http://localhost:%d/%s\", toJSON(%s))", c.server.port, key, name)
	}()

	rd := <-rch
	dec := json.NewDecoder(rd.r)
	err := dec.Decode(data)
	close(rd.done)
	errExec := <-errCh
	if errExec != nil {
		return errExec
	}
	return err
}

type RError interface {
	error
	IsError()
}

type RWarning interface {
	error
	IsWarning()
}

func IsError(e error) bool {
	_, ok := e.(RError)
	return ok
}

func IsWarning(e error) bool {
	_, ok := e.(RWarning)
	return ok
}
