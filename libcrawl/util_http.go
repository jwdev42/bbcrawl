package libcrawl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func httpGet(c *http.Client, url *url.URL) (*http.Response, error) {
	resp, err := c.Get(url.String())
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		return resp, nil
	case 201:
		return resp, nil
	case 202:
		return resp, nil
	case 203:
		return resp, nil
	case 205:
		return resp, nil
	default:
		defer resp.Body.Close()
		return nil, fmt.Errorf("%s GET: %q: %q", resp.Proto, url.String(), resp.Status)
	}
}

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

func downloadURL(url *url.URL, path string) error {
	if !url.IsAbs() {
		panic("url parameter is relative")
	}
	dl, err := http.Get(url.String())
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadAll(dl.Body)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	f.Close()
	dl.Body.Close()
	return nil
}

func printFetchError(url *url.URL) {
	fmt.Fprintf(os.Stderr, "File %q could not be downloaded.", url.String())
}
