package libcrawl

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

func DownloadURL(url string, path string) error {
	dl, err := http.Get(url)
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

func ImgCrawl(dir string, thread string, start int, end int, step int) error {
	const threadsuffix string = ".html"
	const imgtag string = "img"
	var threadd string //derivative of thread
	picid := 1
	if end < start {
		return nil
	}
	count := (start - 1) * step
	if count > 0 {
		li := strings.LastIndex(thread, threadsuffix)
		if li == -1 {
			return fmt.Errorf("cannot split thread string \"%s\"", thread)
		}
		threadd = thread[:li] + fmt.Sprintf("-%d", count) + threadsuffix
	} else {
		threadd = thread
	}
	resp, err := http.Get(threadd)
	if err != nil {
		return err
	}
	body, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	attrs := make([]html.Attribute, 1)
	attrs[0].Key = "class"
	attrs[0].Val = "img-responsive"
	nodes := ElementsByTagAndAttrs(body, imgtag, attrs)
	for _, n := range nodes {
		for _, a := range n.Attr {
			if a.Key == "src" {
				li := strings.LastIndex(a.Val, ".")
				var suffix string
				if li+1 < len(a.Val) {
					suffix = a.Val[li+1:]
				}
				dlpath := fmt.Sprintf("%s/%d-%d.%s", dir, start, picid, suffix)
				if err := DownloadURL(a.Val, dlpath); err != nil {
					fmt.Fprintf(os.Stderr, "download failed: %s: %v\n", dlpath, err)
				}
				picid++
				break
			}
		}
	}
	resp.Body.Close()
	return ImgCrawl(dir, thread, start+1, end, step)
}
