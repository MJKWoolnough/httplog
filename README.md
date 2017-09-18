# httplog
--
    import "github.com/MJKWoolnough/httplog"

Package httplog is used to create wrappers around http.Handler's to gather
information about a request and its response.

## Usage

#### func  NewLogMux

```go
func NewLogMux(m http.Handler, l Logger) http.Handler
```
NewLogMux wraps an existing http.Handler and collects data about the request and
response and passes it to a logger.

#### type Details

```go
type Details struct {
	*http.Request
	Status, ResponseLength int
	StartTime, EndTime     time.Time
}
```

Details is a collection of data about the request and response

#### type Logger

```go
type Logger interface {
	Log(d Details)
}
```

Logger allows clients to specifiy how collected data is handled

#### func  NewWriteLogger

```go
func NewWriteLogger(w io.Writer, format string) (Logger, error)
```
NewWriteLogger uses the given format as a template to write log data to the
given io.Writer

#### type WriteLogger

```go
type WriteLogger struct {
}
```

WriteLogger is a Logger which formats log data to a given template and writes it
to a given io.Writer

#### func (*WriteLogger) Log

```go
func (w *WriteLogger) Log(d Details)
```
Log satisfies the Logger interface
