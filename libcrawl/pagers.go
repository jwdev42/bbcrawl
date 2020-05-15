package libcrawl

import (
	"fmt"
	"net/url"
	"strings"
)

type UnknownBBPager struct {
	Start  int
	End    int
	Posts  int
	Thread *url.URL
	page   int
}

func (r *UnknownBBPager) Next() (*url.URL, error) {
	const threadsuffix string = ".html"
	var thread string
	if r.End < r.Start {
		return nil, fmt.Errorf("Start page is bigger than end page!")
	}
	if r.page > r.End {
		return nil, nil
	}
	if r.page < r.Start {
		r.page = r.Start
	}
	counter := (r.page - 1) * r.Posts
	if counter > 0 {
		li := strings.LastIndex(r.Thread.String(), threadsuffix)
		if li == -1 {
			return nil, fmt.Errorf("cannot split thread string %q", thread)
		}
		thread = r.Thread.String()[:li] + fmt.Sprintf("-%d", counter) + threadsuffix
	} else {
		thread = r.Thread.String()
	}
	r.page++
	if newurl, err := url.Parse(thread); err != nil {
		return nil, err
	} else {
		return newurl, nil
	}
}
