package engine

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
)

type Output struct {
	sync.Mutex
	dests []io.Writer
	tasks sync.WaitGroup
	used  bool
}

// Tail returns the n last lines of a buffer
// stripped out of trailing white spaces, if any.
//
// if n <= 0, returns an empty string
func Tail(buffer *bytes.Buffer, n int) string {
	if n <= 0 {
		return ""
	}
	s := strings.TrimRightFunc(buffer.String(), unicode.IsSpace)
	i := len(s) - 1
	for ; i >= 0 && n > 0; i-- {
		if s[i] == '\n' {
			n--
			if n == 0 {
				break
			}
		}
	}
	// when i == -1, return the whole string which is s[0:]
	return s[i+1:]
}

// NewOutput returns a new Output object with no destinations attached.
// Writing to an empty Output will cause the written data to be discarded.
func NewOutput() *Output {
	return &Output{}
}

// Return true if something was written on this output
func (o *Output) Used() bool {
	o.Lock()
	defer o.Unlock()
	return o.used
}

// Add attaches a new destination to the Output. Any data subsequently written
// to the output will be written to the new destination in addition to all the others.
// This method is thread-safe.
func (o *Output) Add(dst io.Writer) {
	o.Lock()
	defer o.Unlock()
	o.dests = append(o.dests, dst)
}

// Set closes and remove existing destination and then attaches a new destination to
// the Output. Any data subsequently written to the output will be written to the new
// destination in addition to all the others. This method is thread-safe.
func (o *Output) Set(dst io.Writer) {
	o.Close()
	o.Lock()
	defer o.Unlock()
	o.dests = []io.Writer{dst}
}

// AddPipe creates an in-memory pipe with io.Pipe(), adds its writing end as a destination,
// and returns its reading end for consumption by the caller.
// This is a rough equivalent similar to Cmd.StdoutPipe() in the standard os/exec package.
// This method is thread-safe.
func (o *Output) AddPipe() (io.Reader, error) {
	r, w := io.Pipe()
	o.Add(w)
	return r, nil
}

// Write writes the same data to all registered destinations.
// This method is thread-safe.
func (o *Output) Write(p []byte) (n int, err error) {
	o.Lock()
	defer o.Unlock()
	o.used = true
	var firstErr error
	for _, dst := range o.dests {
		_, err := dst.Write(p)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return len(p), firstErr
}

// Close unregisters all destinations and waits for all background
// AddTail and AddString tasks to complete.
// The Close method of each destination is called if it exists.
func (o *Output) Close() error {
	o.Lock()
	defer o.Unlock()
	var firstErr error
	for _, dst := range o.dests {
		if closer, ok := dst.(io.Closer); ok {
			err := closer.Close()
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	o.tasks.Wait()
	o.dests = nil
	return firstErr
}

type Input struct {
	src io.Reader
	sync.Mutex
}

// NewInput returns a new Input object with no source attached.
// Reading to an empty Input will return io.EOF.
func NewInput() *Input {
	return &Input{}
}

// Read reads from the input in a thread-safe way.
func (i *Input) Read(p []byte) (n int, err error) {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()
	if i.src == nil {
		return 0, io.EOF
	}
	return i.src.Read(p)
}

// Closes the src
// Not thread safe on purpose
func (i *Input) Close() error {
	if i.src != nil {
		if closer, ok := i.src.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// Add attaches a new source to the input.
// Add can only be called once per input. Subsequent calls will
// return an error.
func (i *Input) Add(src io.Reader) error {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()
	if i.src != nil {
		return fmt.Errorf("Maximum number of sources reached: 1")
	}
	i.src = src
	return nil
}

// AddEnv starts a new goroutine which will decode all subsequent data
// as a stream of json-encoded objects, and point `dst` to the last
// decoded object.
// The result `env` can be queried using the type-neutral Env interface.
// It is not safe to query `env` until the Output is closed.
func (o *Output) AddEnv() (dst *Env, err error) {
	src, err := o.AddPipe()
	if err != nil {
		return nil, err
	}
	dst = &Env{}
	o.tasks.Add(1)
	go func() {
		defer o.tasks.Done()
		decoder := NewDecoder(src)
		for {
			env, err := decoder.Decode()
			if err != nil {
				return
			}
			*dst = *env
		}
	}()
	return dst, nil
}
