// Package logplex implements streaming of syslog messages
package logplex

import (
	"io"
	"runtime"
	"strconv"
	"time"
)

type Msg struct {
	Priority  int
	Timestamp []byte
	Host      []byte
	User      []byte
	Pid       []byte
	Id        []byte
	Msg       []byte
}

func (m *Msg) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(m.Timestamp))
}

type BytesReader interface {
	io.Reader
	ReadBytes(delim byte) (line []byte, err error)
}

// Reader reads syslog streams
type Reader struct {
	buf BytesReader
}

// NewReader returns a new Reader that reads from buf.
func NewReader(buf BytesReader) *Reader {
	return &Reader{buf: buf}
}

// ReadMsg returns a single Msg. If no data is available, returns an error.
func (r *Reader) ReadMsg() (m *Msg, err error) {
	defer errRecover(&err)

	b := r.next()

	m = new(Msg)
	m.Priority = b.priority()
	m.Timestamp = b.bytes()
	m.Host = b.bytes()
	m.User = b.bytes()
	m.Pid = b.bytes()
	m.Id = b.bytes()
	m.Msg = b

	return
}

func (r *Reader) next() readBuf {
	b, err := r.buf.ReadBytes(' ')
	if err != nil {
		panic(err)
	}
	b = b[:len(b)-1]

	n, err := strconv.Atoi(string(b))
	if err != nil {
		panic(err)
	}

	buf := make(readBuf, n)
	_, err = io.ReadFull(r.buf, buf)
	if err != nil {
		panic(err)
	}

	return buf
}

func errRecover(err *error) {
	e := recover()
	if e != nil {
		switch ee := e.(type) {
		case runtime.Error:
			panic(e)
		case error:
			*err = ee
		default:
			panic(e)
		}
	}
}
