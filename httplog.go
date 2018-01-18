// Package httplog is used to create wrappers around http.Handler's to gather
// information about a request and its response.
package httplog

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/MJKWoolnough/httpwrap"
)

// DefaultLog is a simple template to output log data in something reminiscent
// on the Apache default format
const DefaultFormat = "{{.RemoteAddr}} - {{.URL.User.Username}} - [{{.StartTime.Format \"02/01/2006:15:04:05 +0700\"}}] \"{{.Method}} {{.URL.RequestURI}} {{.Proto}}\" {{.Status}} {{.RequestLength}} {{.StartTime.Sub .EndTime}}"

// Details is a collection of data about the request and response
type Details struct {
	*http.Request
	Status, ResponseLength int
	StartTime, EndTime     time.Time
}

type wrapRW struct {
	http.ResponseWriter
	status, contentLength *int
}

func (w *wrapRW) WriteHeader(n int) {
	*w.status = n
	w.ResponseWriter.WriteHeader(n)
}

type logMux struct {
	http.Handler
	Logger
}

// Logger allows clients to specifiy how collected data is handled
type Logger interface {
	Log(d Details)
}

// Wrap wraps an existing http.Handler and collects data about the request
// and response and passes it to a logger.
func Wrap(m http.Handler, l Logger) http.Handler {
	if m == nil {
		m = http.DefaultServeMux
	}
	return &logMux{Handler: m, Logger: l}
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return new(wrapRW)
	},
}

// ServeHTTP satisfies the http.Handler interface
func (l *logMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := Details{
		Request: r,
		Status:  200,
	}

	rw := responsePool.Get().(*wrapRW)

	*rw = wrapRW{
		w,
		&d.Status,
		&d.ResponseLength,
	}
	d.StartTime = time.Now()
	l.Handler.ServeHTTP(
		httpwrap.Wrap(w, httpwrap.OverrideWriter(rw), httpwrap.OverrideHeaderWriter(rw)),
		r,
	)
	d.EndTime = time.Now()

	*rw = wrapRW{}
	responsePool.Put(rw)

	go l.Logger.Log(d)
}

// WriteLogger is a Logger which formats log data to a given template and
// writes it to a given io.Writer
type WriteLogger struct {
	mu       sync.Mutex
	w        io.Writer
	template *template.Template
}

// NewWriteLogger uses the given format as a template to write log data to the
// given io.Writer
func NewWriteLogger(w io.Writer, format string) (Logger, error) {
	if format == "" {
		return nil, errors.New("invalid format")
	}
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	t, err := template.New("").Parse(format)
	if err != nil {
		return nil, err
	}
	return &WriteLogger{
		w:        w,
		template: t,
	}, nil
}

// Log satisfies the Logger interface
func (w *WriteLogger) Log(d Details) {
	w.mu.Lock()
	w.template.Execute(w.w, d)
	w.mu.Unlock()
}
