package httplog

import (
	"io"
	"net/http"
)

type LogMux struct {
	http.Handler
	io.Writer
}

func NewLogMux(m http.Handler, w io.Writer) *LogMux {
	if m == nil {
		m = http.DefaultServeMux
	}
	return &LogMux{Handler: m, Writer: w}
}

func (l *LogMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.Handler.ServeHTTP(w, r)
}
