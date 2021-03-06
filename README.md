Gzip Handler
============

This is a tiny Go package which wraps HTTP handlers to transparently gzip the
response body, for clients which support it. Although it's usually simpler to
leave that to a reverse proxy (like nginx or Varnish), this package is useful
when that's undesirable.

This is a fork of the [original][nytimes] version, heavily optimised for
performance and low latency.

## Usage

Call `Wrap` with any handler (an object which implements the
`http.Handler` interface), and it'll return a new handler which gzips the
response. For example:

```go
package main

import (
	"io"
	"net/http"
	"github.com/bsm/gziphandler"
)

func main() {
	withoutGz := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Hello, World")
	})

	withGz := gziphandler.Wrap(withoutGz)

	http.Handle("/", withGz)
	http.ListenAndServe("0.0.0.0:8000", nil)
}
```


## Documentation

The docs can be found at [godoc.org][docs], as usual.


## License

[Apache 2.0][license].


[nytimes]:  https://github.com/NYTimes/gziphandler
[docs]:     https://godoc.org/github.com/bsm/gziphandler
[license]:  https://github.com/bsm/gziphandler/blob/master/LICENSE.md
