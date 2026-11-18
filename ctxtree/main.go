// Contexts form a tree: cancel a parent, and every child is cancelled too.
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	crawl, cancelCrawl := context.WithCancel(context.Background())
	page, cancelPage := context.WithTimeout(crawl, time.Hour) // a child, with a long deadline
	defer cancelPage()
	retry, cancelRetry := context.WithCancel(page) // a grandchild
	defer cancelRetry()

	cancelCrawl() // cancel only the top
	for _, c := range []struct {
		name string
		ctx  context.Context
	}{{"crawl", crawl}, {"page", page}, {"retry", retry}} {
		fmt.Printf("%-6s %v\n", c.name, c.ctx.Err())
	}
}
