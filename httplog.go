package httplog

import (
	"net/http"
	"net/url"
	"time"
)

type Details struct {
	Status, ContentLength, RequestLength                       int
	URL                                                        *url.URL
	Method, Proto, Host, RemoteAddr, UserAgent, User, Referrer string
	StartTime, EndTime                                         time.Time
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

type LogMux struct {
	http.Handler
	fn func(Details)
}

func NewLogMux(m http.Handler, fn func(Details)) *LogMux {
	if m == nil {
		m = http.DefaultServeMux
	}
	return &LogMux{Handler: m, fn: fn}
}

func (l *LogMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := Details{
		RemoteAddr:    r.RemoteAddr,
		Proto:         r.Proto,
		Method:        r.Method,
		URL:           r.URL,
		UserAgent:     r.UserAgent(),
		Host:          r.Host,
		Referer:       r.Referer(),
		RequestLength: r.ContentLength,
		Status:        200,
	}

	user, _, ok := r.BasicAuth()
	if ok {
		d.User = user
	}

	if pusher, ok := w.(http.Pusher); ok {
		w = &wrapPusher{
			wrapRW{
				w,
				&d.Status,
				&d.ContentLength,
			},
			pusher,
		}
	} else {
		w = &wrapRW{
			w,
			&d.Status,
			&d.ContentLength,
		}
	}

	d.StartTime = time.Now()
	l.Handler.ServeHTTP(w, r)
	d.EndTime = time.Now()

	go l.fn(d)
}
