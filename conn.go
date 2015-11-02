package rgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Conn is used to start and communicate with an R process. Conn
// is NOT thread-safe. However, you may run multiple Conn's in
// the same process safely. It is safe to run successive
// operations on Conn successively and retrieve any errors
// using Error(). Do not that although R warnings will be
// returned in method calls, they will not be captured in Error().
type Conn struct {
	cmd     *exec.Cmd
	inPipe  io.WriteCloser
	counter uint64
	server  *server

	err error
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

const cmdStr = `..rgo.ret = c("", "")
tryCatch({
	%s
}, warning = function(w) {
	..rgo.ret[1] <<- conditionMessage(w)
}, error = function(e) {
	..rgo.ret[2] <<- conditionMessage(e)
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

// R sends a command to R. An Error or Warning generated by the
// command will be returned as an RError or RWarning.
func (c *Conn) R(cmd string) error {
	if c.err != nil {
		return c.err
	}
	key := "r.result"
	rch := make(chan readerDone)
	c.server.putFwd(key, rch)
	defer c.server.rmFwd(key)
	fmt.Fprintf(c.inPipe, cmdStr, cmd, c.server.port, key)
	rd := <-rch
	defer close(rd.done)
	dec := json.NewDecoder(rd.r)
	var resultPair []string
	if err := dec.Decode(&resultPair); err != nil {
		c.err = fmt.Errorf("error while decoding result: %v", err)
		return c.err
	}
	if len(resultPair) != 2 {
		c.err = fmt.Errorf("invalid result pair: %v has length %d", resultPair, len(resultPair))
		return c.err
	}
	result := res{resultPair[0], resultPair[1]}
	c.err = result.toError()
	if IsWarning(c.err) {
		err := c.err
		c.err = nil
		return err
	}
	return c.err
}

// Rf is like R but takes a format string and arguments.
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

// Send sends data into R. data must be json-serializable.
func (c *Conn) Send(data interface{}, name string) error {
	if c.err != nil {
		return c.err
	}
	key, err := c.write(data)
	if key != "" {
		defer c.server.rmData(key)
	}
	if err != nil {
		c.err = err
		return err
	}
	return c.Rf("%s = fromJSON(getURL(\"http://localhost:%d/%s\"))", name, c.server.port, key)
}

// Get gets data from R. data will be deserialized from json.
func (c *Conn) Get(data interface{}, name string) error {
	if c.err != nil {
		return c.err
	}
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
	c.err = <-errCh
	if c.err != nil {
		return c.err
	}
	c.err = err
	return err
}

// Error returns the first error that occured in the sequence of
// operations. R warnings are ignored.
func (c *Conn) Error() error {
	return c.err
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
