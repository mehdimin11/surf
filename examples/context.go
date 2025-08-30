package main

import (
	"context"
	"log"
	"time"

	"github.com/enetx/surf"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	cli := surf.NewClient().Builder().WithContext(ctx).Build()
	req := cli.Get("https://httpbingo.org/get")

	resp := req.Do()
	if resp.IsErr() {
		log.Fatal(resp.Err())
	}

	log.Println(resp.Ok().Body.String())
}
