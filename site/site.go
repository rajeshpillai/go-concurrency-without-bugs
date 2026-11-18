// Package site is a small fake website for the crawler to visit, so every run is the same.
// 300 pages; page N links to four others. Each page takes 50 ms to answer.
// With Trap set, one page (/page/137) never answers: like a broken server out on the internet.
package site

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"time"
)

const Pages = 300

type Site struct {
	URL     string
	server  *httptest.Server
	now     atomic.Int64 // requests being answered right now
	Peak    atomic.Int64 // the most at once
	Served  atomic.Int64 // pages answered
	Trap    bool
	Latency time.Duration
}

// Links returns the four pages page n links to.
func Links(n int) []int {
	return []int{(n*7 + 1) % Pages, (n*13 + 5) % Pages, (n*31 + 11) % Pages, (n*3 + 2) % Pages}
}

func Start(trap bool) *Site {
	s := &Site{Trap: trap, Latency: 50 * time.Millisecond}
	s.server = httptest.NewServer(http.HandlerFunc(s.page))
	s.URL = s.server.URL
	return s
}

func (s *Site) Close() { s.server.Close() }

func (s *Site) page(w http.ResponseWriter, r *http.Request) {
	n := s.now.Add(1)
	defer s.now.Add(-1)
	for p := s.Peak.Load(); n > p && !s.Peak.CompareAndSwap(p, n); p = s.Peak.Load() {
	}
	id, err := strconv.Atoi(r.URL.Path[len("/page/"):])
	if err != nil || id < 0 || id >= Pages {
		http.NotFound(w, r)
		return
	}
	if s.Trap && id == 137 {
		<-r.Context().Done() // never answers; gives up only when the client does
		return
	}
	time.Sleep(s.Latency)
	s.Served.Add(1)
	fmt.Fprintf(w, "<h1>Page %d</h1>\n", id)
	for _, l := range Links(id) {
		fmt.Fprintf(w, "<a href=\"/page/%d\">page %d</a>\n", l, l)
	}
}
