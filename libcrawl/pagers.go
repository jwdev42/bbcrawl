package libcrawl

import (
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"net/url"
	"strconv"
	"strings"
)

const (
	PAGER_UNKNOWNBB = "unknownbb"
	PAGER_VB4       = "vb4"
)

type UnknownBBPager struct {
	Start  int
	End    int
	Posts  int
	Thread *url.URL
	page   int
	cc     *CrawlContext
}

func NewUnknownBBPager(cc *CrawlContext) PagerInterface {
	return &UnknownBBPager{cc: cc}
}

func (r *UnknownBBPager) PageNum() int {
	return r.page - 1
}

func (r *UnknownBBPager) SetOptions(args []string) error {
	start := cmdline.StartPage(0)
	end := cmdline.NewEndPage(&start)
	set := flag.NewFlagSet("UnknownBBPager", flag.ContinueOnError)
	set.Var(&start, "start", "start page")
	set.Var(end, "end", "end page")
	posts := set.Int("posts", 0, "posts per page")
	if err := set.Parse(args); err != nil {
		return err
	}
	if start < 1 {
		return fmt.Errorf("Start page not set")
	}
	r.Start = int(start)
	if end.End < r.Start {
		return fmt.Errorf("End page not set")
	}
	r.End = end.End
	if *posts < 1 {
		return fmt.Errorf("Amount of posts per page not set or < 1")
	}
	r.Posts = *posts
	return nil
}

func (r *UnknownBBPager) SetUrl(addr string) error {
	u, err := url_for_pager(addr)
	if err != nil {
		return err
	}
	r.Thread = u
	return nil
}

func (r *UnknownBBPager) Next() (*url.URL, error) {
	const threadsuffix string = ".html"
	var thread string

	if r.page < r.Start {
		r.page = r.Start
	}
	if r.page > r.End {
		return nil, nil
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

type VB4Pager struct {
	Start  int
	End    int
	Thread *url.URL
	cc     *CrawlContext
	page   int
}

func NewVB4Pager(cc *CrawlContext) PagerInterface {
	return &VB4Pager{cc: cc}
}

func (r *VB4Pager) Next() (*url.URL, error) {
	if r.page < r.Start {
		r.page = r.Start
	}
	if r.page > r.End {
		return nil, nil
	}
	if r.page == 1 {
		r.page++
		return r.Thread, nil
	}
	var newpage string
	thread := r.Thread.String()
	page_str := strconv.Itoa(r.page)
	if strings.LastIndexByte(thread, '/') == len(thread)-1 {
		newpage = thread + "page" + page_str
	} else {
		newpage = thread + "/page" + page_str
	}
	r.page++
	if newurl, err := url.Parse(newpage); err != nil {
		return nil, err
	} else {
		return newurl, nil
	}
}

func (r *VB4Pager) PageNum() int {
	return r.page - 1
}

func (r *VB4Pager) SetOptions(args []string) error {
	start := cmdline.StartPage(0)
	end := cmdline.NewEndPage(&start)
	set := flag.NewFlagSet("VB4Pager", flag.ContinueOnError)
	set.Var(&start, "start", "")
	set.Var(end, "end", "")
	if err := set.Parse(args); err != nil {
		return err
	}
	if start < 1 {
		return fmt.Errorf("Start page not set")
	}
	r.Start = int(start)
	if end.End < r.Start {
		return fmt.Errorf("End page not set")
	}
	r.End = end.End
	return nil
}

func (r *VB4Pager) SetUrl(addr string) error {
	u, err := url_for_pager(addr)
	if err != nil {
		return err
	}
	r.Thread = u
	return nil
}
