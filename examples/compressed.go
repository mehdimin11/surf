package main

import (
	"github.com/enetx/surf"
)

func main() {
	surf.NewClient().Get("https://httpbin.org/gzip").Do().Ok().Body.String().Print()
	surf.NewClient().Get("https://httpbin.org/deflate").Do().Ok().Body.String().Print()
	surf.NewClient().Get("https://httpbin.org/brotli").Do().Ok().Body.String().Print()
}
