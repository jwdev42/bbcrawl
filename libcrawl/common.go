package libcrawl

import (
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func WalkTree(n *html.Node, pre, post func(*html.Node) bool) bool {
	if pre != nil && !pre(n) {
		return false
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !WalkTree(c, pre, post) {
			return false
		}
	}
	if post != nil && !post(n) {
		return false
	}
	return true
}

func ElementsByTagAndAttrs(n *html.Node, id string, attrs []html.Attribute) []*html.Node {
	nodes := make([]*html.Node, 0, 10)
	pre := func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == id {
			for _, a := range attrs {
				found := false
				for _, na := range n.Attr {
					if a == na {
						found = true
						break
					}
				}
				if !found {
					return true
				}
			}
			nodes = append(nodes, n)
		}
		return true
	}
	WalkTree(n, pre, nil)
	return nodes
}

func DownloadURL(url *url.URL, path string) error {
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

func rel2absURL(domain *url.URL, url *url.URL) (*url.URL, error) {
	if url.IsAbs() {
		panic("url parameter is absolute")
	}
	if !domain.IsAbs() {
		panic("domain parameter is relative")
	}
	if newurl, err := url.Parse(domain.Hostname() + url.EscapedPath()); err != nil {
		return nil, err
	} else {
		return newurl, nil
	}
}
