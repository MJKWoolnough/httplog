package httplog

import (
	"net/http"
	"time"
)

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

type wrapPusher struct {
	wrapRW
	http.Pusher
}

type logMux struct {
	http.Handler
	fn func(Details)
}

func NewLogMux(m http.Handler, fn func(Details)) *logMux {
	if m == nil {
		m = http.DefaultServeMux
	}
	return &logMux{Handler: m, fn: fn}
}

func (l *logMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := Details{
		Request: r,
		Status:  200,
	}

	if pusher, ok := w.(http.Pusher); ok {
		w = &wrapPusher{
			wrapRW{
				w,
				&d.Status,
				&d.ResponseLength,
			},
			pusher,
		}
	} else {
		w = &wrapRW{
			w,
			&d.Status,
			&d.ResponseLength,
		}
	}

	d.StartTime = time.Now()
	l.Handler.ServeHTTP(w, r)
	d.EndTime = time.Now()

	go l.fn(d)
}
