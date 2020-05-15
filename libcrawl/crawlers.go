package libcrawl

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ThumbCrawler struct {
	Out      string
	Page     int
	Excluded []*url.URL
}

func (r *ThumbCrawler) Crawl(url *url.URL) error {
	const imgtag string = "img"
	picid := 1
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	body, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
				} else {
					fmt.Fprintf(os.Stderr, "image suffix missing: %s\n", a.Val)
					break
				}
				dlpath := fmt.Sprintf("%s/%d-%d.%s", r.Out, r.Page, picid, suffix)
				dl, err := url.Parse(a.Val)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					break
				}
				if !dl.IsAbs() {
					dl, err = rel2absURL(url, dl)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v\n", err)
						break
					}
				}
				if r.isExcluded(dl) {
					break
				}
				if err := DownloadURL(dl, dlpath); err != nil {
					fmt.Fprintf(os.Stderr, "download failed: %s: %v\n", dlpath, err)
					break
				}
				picid++
				break
			}
		}
	}
	r.Page++
	return nil
}

func (r *ThumbCrawler) isExcluded(url *url.URL) bool {
	for _, exurl := range r.Excluded {
		if exurl.String() == url.String() {
			return true
		}
	}
	return false
}
