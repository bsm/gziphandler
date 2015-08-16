package gziphandler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	scheme          = "gzip"
)

// gzipResponseWriter provides an http.ResponseWriter interface, which gzips
// bytes before writing them to the underlying response. This doesn't set the
// Content-Encoding header, nor close the writers, so don't forget to do that.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write appends data to the gzip writer.
func (gzw gzipResponseWriter) Write(b []byte) (int, error) {
	return gzw.Writer.Write(b)
}

// WrapLevel behaves like GzipHandler but allows a custom GZIP
// compression level. Invalid compression level inputs are reset to default.
func WrapLevel(h http.Handler, level int) http.Handler {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	pool := sync.Pool{
		New: func() interface{} { w, _ := gzip.NewWriterLevel(nil, level); return w },
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(vary, acceptEncoding)

		if !strings.Contains(r.Header.Get(acceptEncoding), scheme) {
			h.ServeHTTP(w, r)
			return
		}

		// Bytes written during ServeHTTP are redirected to this gzip writer
		// before being written to the underlying response.
		gzw := pool.Get().(*gzip.Writer)
		gzw.Reset(w)
		defer gzw.Close()

		w.Header().Set(contentEncoding, scheme)
		h.ServeHTTP(gzipResponseWriter{gzw, w}, r)
		pool.Put(gzw)
	})
}

// Wrap wraps an HTTP handler, to transparently gzip the response body if
// the client supports it (via the Accept-Encoding header).
func Wrap(h http.Handler) http.Handler {
	return WrapLevel(h, gzip.DefaultCompression)
}
