/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

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
	PAGER_VB4    = "vb4"
	PAGER_QUERY  = "query"
	PAGER_URLCUT = "cutter"
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

// URLCuttingPager browses through the pages by cutting out a part of itself and replacing that with an increasing number.
type URLCuttingPager struct {
	end, page, step, adjust       int
	cut                           []int
	startpage                     *url.URL
	leftpart, rightpart, digitfmt string
}

func NewURLCuttingPager(cc *CrawlContext) PagerInterface {
	return new(URLCuttingPager)
}

func (r *URLCuttingPager) Next() (*url.URL, error) {
	if r.startpage != nil {
		ret := r.startpage
		r.startpage = nil
		return ret, nil
	}
	if r.page > r.end {
		return nil, nil
	}
	fmtstr := fmt.Sprintf("%%s%s%%s", r.digitfmt)
	u, err := url.Parse(fmt.Sprintf(fmtstr, r.leftpart, r.page*r.step, r.rightpart))
	if err != nil {
		return nil, err
	}
	r.page++
	return u, nil
}

func (r *URLCuttingPager) PageNum() int {
	return r.page - 1 + r.adjust
}

func (r *URLCuttingPager) SetOptions(args []string) error {
	var cut = new(cmdline.IntTuple)
	//setup
	set := flag.NewFlagSet("URLCuttingPager", flag.ContinueOnError)
	adjp := set.Int("adjust", 0, "adjust the page reported to the crawler")
	startp, endp := set.Int("start", -1, "first page"), set.Int("end", -1, "last page")
	stepp := set.Int("step", 1, "number of pages to advance with every page load")
	digitsp := set.Int("digits", 0, "number of digits to fill, do not set for auto mode")
	startpagep := set.String("startpage", "", "if set, the given url will be used as the start page before using the regular url.")
	set.Var(cut, "cut", "range in the url you want to cut out and replace with the page number")
	if err := set.Parse(args); err != nil {
		return err
	}
	//validation
	if *startp < 0 {
		return fmt.Errorf("start not set or set to an illegal value")
	}
	if *startp > *endp {
		return fmt.Errorf("end must not be smaller than start")
	}
	if *stepp < 1 {
		return fmt.Errorf("step set to an illegal value")
	}
	if cut.Numbers[0] == 0 {
		return fmt.Errorf("cut: first argument cannot be 0")
	}
	if len(cut.Numbers) != 2 {
		return fmt.Errorf("cut needs 2 integers")
	}
	if cut.Numbers[1] < 0 {
		return fmt.Errorf("cut: cannot cut out a negative amount of characters")
	}
	if *startpagep != "" {
		if u, err := url.Parse(*startpagep); err != nil {
			return fmt.Errorf("startpage: %v", err)
		} else {
			r.startpage = u
		}
	}
	if *digitsp > 0 && *digitsp < len(strconv.Itoa(*endp)) {
		return fmt.Errorf("digits: not enough space to hold the desired page numbers")
	}
	//set pager vars
	r.adjust = *adjp
	r.page, r.end, r.step = *startp, *endp, *stepp
	r.cut = cut.Numbers
	if *digitsp > 0 {
		r.digitfmt = fmt.Sprintf("%%0%dd", *digitsp)
	} else {
		r.digitfmt = "%d"
	}
	return nil
}

func (r *URLCuttingPager) SetUrl(addr string) error {
	//test if url is valid
	if _, err := url_for_pager(addr); err != nil {
		return err
	}
	cutindex := r.cut[0]
	if cutindex < 0 {
		cutindex = len(addr) + cutindex + 1
	}
	//split address in a left and a right part
	if len(addr) <= cutindex-1 || cutindex < 1 {
		return fmt.Errorf("cutoff index out of range")
	}
	r.leftpart = addr[:cutindex-1]
	if len(addr) > cutindex-1+r.cut[1] {
		r.rightpart = addr[cutindex-1+r.cut[1]:]
	}
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
