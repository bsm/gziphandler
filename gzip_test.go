package gziphandler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipHandler(t *testing.T) {
	testBody := "aaabbbccc"

	// This just exists to provide something for GzipHandler to wrap.
	handler := newTestHandler(testBody)

	// requests without accept-encoding are passed along as-is

	req1, _ := http.NewRequest("GET", "/whatever", nil)
	res1 := httptest.NewRecorder()
	handler.ServeHTTP(res1, req1)

	assert.Equal(t, 200, res1.Code)
	assert.Equal(t, "", res1.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", res1.Header().Get("Vary"))
	assert.Equal(t, testBody, res1.Body.String())

	// but requests with accept-encoding:gzip are compressed if possible

	req2, _ := http.NewRequest("GET", "/whatever", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	res2 := httptest.NewRecorder()
	handler.ServeHTTP(res2, req2)

	assert.Equal(t, 200, res2.Code)
	assert.Equal(t, "gzip", res2.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", res2.Header().Get("Vary"))
	assert.Equal(t, gzipStr(testBody), res2.Body.Bytes())
}

// --------------------------------------------------------------------

func BenchmarkGzipHandler_Serial(b *testing.B) {
	req, _ := http.NewRequest("GET", "/whatever", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler := newTestHandler(strings.Repeat("aaabbbccc", 500))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runBenchmark(b, req, handler)
	}
}

func BenchmarkGzipHandler_Parallel(b *testing.B) {
	req, _ := http.NewRequest("GET", "/whatever", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler := newTestHandler(strings.Repeat("aaabbbccc", 500))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBenchmark(b, req, handler)
		}
	})
}

// --------------------------------------------------------------------

func gzipStr(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	io.WriteString(w, s)
	w.Close()
	return b.Bytes()
}

func runBenchmark(b *testing.B, req *http.Request, handler http.Handler) {
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if code := res.Code; code != 200 {
		b.Fatalf("Expected 200 but got %d", code)
	} else if blen := res.Body.Len(); blen < 50 || blen > 75 {
		b.Fatalf("Expected complete response body, but got %d bytes", blen)
	}
}

func newTestHandler(body string) http.Handler {
	return New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, body)
	}))
}
