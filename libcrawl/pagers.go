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
	PAGER_QUERY     = "query"
	PAGER_URLFMT = "format"
)

type QueryPager struct {
	counter struct {
		id    string
		val   int
		limit int
	}
	thread *url.URL
	query  url.Values
	cc     *CrawlContext
}

func NewQueryPager(cc *CrawlContext) PagerInterface {
	return &QueryPager{cc: cc}
}

func (r *QueryPager) Next() (*url.URL, error) {
	if r.counter.val > r.counter.limit {
		return nil, nil
	}
	r.query.Set(r.counter.id, strconv.Itoa(r.counter.val))
	u, err := url.Parse(r.thread.String())
	if err != nil {
		return nil, err
	}
	u.RawQuery = r.query.Encode()
	r.counter.val++
	return u, nil
}

func (r *QueryPager) PageNum() int {
	return r.counter.val - 1
}

func (r *QueryPager) SetOptions(args []string) error {
	start := cmdline.StartPage(0)
	end := cmdline.NewEndPage(&start)
	set := flag.NewFlagSet("QueryPager", flag.ContinueOnError)
	set.Var(&start, "start", "start page")
	set.Var(end, "end", "end page")
	namep := set.String("name", "page", "identifier for the page variable in the query string")
	if err := set.Parse(args); err != nil {
		return err
	}
	if start < 1 {
		return fmt.Errorf("Start page not set")
	}
	r.counter.val = int(start)
	if end.End < r.counter.val {
		return fmt.Errorf("End page not set")
	}
	r.counter.limit = end.End
	if len(*namep) == 0 {
		return fmt.Errorf("Page identifier not set")
	}
	r.counter.id = *namep
	return nil
}

func (r *QueryPager) SetUrl(addr string) error {
	var u_str, q_str string
	s := strings.SplitN(addr, "?", 2)
	switch len(s) {
	case 2:
		q_str = s[1]
		fallthrough
	case 1:
		u_str = s[0]
	default:
		panic("You are not supposed to be here!")
	}
	u, err := url_for_pager(u_str)
	if err != nil {
		return err
	}
	r.query, err = url.ParseQuery(q_str)
	if err != nil {
		return err
	}
	r.thread = u
	return nil
}

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

//URLFormatPager browses through the pages by having the url as a format string.
type URLFormatPager struct {
	end int
	page int
	step int
	adjust int
	startpage *url.URL
	usestartpage bool
	fmtstr string
}

func NewURLFormatPager(cc *CrawlContext) PagerInterface {
	return new(URLFormatPager)
}

func (r *URLFormatPager) Next() (*url.URL, error) {
	if r.usestartpage {
		r.usestartpage = false
		log.Debug(fmt.Sprintf("QueryPager: Sending url %q", r.startpage))
		return r.startpage, nil
	}
	if r.page > r.end {
		return nil, nil
	}
	u, err := url.Parse(fmt.Sprintf(r.fmtstr, r.page * r.step))
	if err != nil {
		return nil, err
	}
	r.page++
	log.Debug(fmt.Sprintf("QueryPager: Sending url %q", u))
	return u, nil
}

func (r *URLFormatPager) PageNum() int {
	return r.page - 1 + r.adjust
}

func (r *URLFormatPager) SetOptions(args []string) error {
	//setup
	set := flag.NewFlagSet("URLFormatPager", flag.ContinueOnError)
	adjp := set.Int("adjust", 0, "adjust the page reported to the crawler")
	startp, endp := set.Int("start", -1, "first page"), set.Int("end", -1, "last page")
	stepp := set.Int("step", 1, "number of pages to advance with every page load")
	fmtstrp := set.String("format" ,"", "url format string")
	startpagep := set.Bool("startpage", false, "if true, the url at the end of the command line will be used as the start page before using the format string. If false (default), that url will be ignored.")
	if err := set.Parse(args); err != nil {
		return err
	}
	//validation
	if *startp < 0 {
		return fmt.Errorf("start not set or set to an illegal value")
	}
	if *stepp < 1 {
		return fmt.Errorf("step set to an illegal value")
	}
	if len(*fmtstrp) == 0 {
		return fmt.Errorf("format not set")
	}
	if u, err := url.Parse(fmt.Sprintf(*fmtstrp, 1)); err != nil {
		return fmt.Errorf("format: format string does not produce a valid url: %w", err)
	} else if !u.IsAbs() {
		return fmt.Errorf("format: url is not absolute")
	}
	//set pager vars
	r.adjust = *adjp
	r.page, r.end, r.step, r.fmtstr, r.usestartpage = *startp, *endp, *stepp, *fmtstrp, *startpagep
	return nil
}

func (r *URLFormatPager) SetUrl(addr string) error {
	u, err := url_for_pager(addr)
	if err != nil {
		return err
	}
	r.startpage = u
	return nil
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
