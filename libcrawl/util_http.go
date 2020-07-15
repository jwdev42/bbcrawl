package libcrawl

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func noRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		return fmt.Errorf("Redirection: %q → %q", lastReq.URL.String(), req.URL.String())
	}
	return nil
}

func logRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		fmt.Printf("Redirection: %q → %q\n", lastReq.URL.String(), req.URL.String())
	}
	if len(via) > DEFAULT_REDIRECTS {
		return fmt.Errorf("Too many redirects")
	}
	return nil
}

func printFetchError(url *url.URL) {
	fmt.Fprintf(os.Stderr, "File %q could not be downloaded.", url.String())
}
