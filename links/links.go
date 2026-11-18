// Package links reads one page and returns the links on it.
package links

import (
	"context"
	"io"
	"net/http"
	"regexp"
)

var href = regexp.MustCompile(`href="([^"]+)"`)

// Get downloads url and returns the links it points to (as full URLs on the same site).
func Get(client *http.Client, base, url string) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range href.FindAllStringSubmatch(string(body), -1) {
		out = append(out, base+m[1])
	}
	return out, nil
}

// GetContext is Get, but it gives up when ctx is cancelled or its deadline passes.
func GetContext(ctx context.Context, client *http.Client, base, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range href.FindAllStringSubmatch(string(body), -1) {
		out = append(out, base+m[1])
	}
	return out, nil
}
