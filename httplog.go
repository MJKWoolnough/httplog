package httplog

import (
	"net/http"
	"net/url"
	"time"
)

type Details struct {
	Status, ContentLength                                      int
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
	var d Details

	d.URL = r.URL
	d.Method = r.Method
	d.UserAgent = r.UserAgent()
	d.Proto = r.Proto
	d.Host = r.Host
	d.RemoteAddr = r.RemoteAddr
	d.Referer = r.Referer()
	user, _, ok := r.BasicAuth()
	if ok {
		d.User = user
	}
	d.Status = 200

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
