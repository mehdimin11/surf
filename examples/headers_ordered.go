package main

import (
	"log"

	"github.com/enetx/g"
	"github.com/enetx/surf"
)

func main() {
	// const url = "https://localhost"
	const url = "https://tls.peet.ws/api/all"

	// oh := g.NewMapOrd[string, string]()
	oh := g.NewMapOrd[g.String, g.String]()
	oh.Set(":method", "")
	oh.Set(":authority", "")
	oh.Set(":scheme", "")
	oh.Set(":path", "")
	oh.Set("1", "1")
	oh.Set("2", "2")
	oh.Set("User-Agent", "")
	oh.Set("3", "3")
	oh.Set("4", "4")
	oh.Set("Accept-Encoding", "gzip")

	r := surf.NewClient().
		Builder().
		UserAgent("root").
		SetHeaders(oh).
		Build().
		Get(url).
		Do()

	if r.IsErr() {
		log.Fatal(r.Err())
	}

	r.Ok().Debug().Request(true).Response(true).Print()
}
