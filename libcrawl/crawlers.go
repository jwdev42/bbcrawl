/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"github.com/jwdev42/bbcrawl/libcrawl/download"
	"github.com/jwdev42/bbcrawl/libhtml"
	"github.com/jwdev42/logger"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CRAWLER_VB_ATTACHMENTS = "vb-attachments"
	CRAWLER_SRC            = "src"
	CRAWLER_FILE           = "file"
)

var vb4_regex_postid *regexp.Regexp = regexp.MustCompile("^post_?[0-9]+$")
var vb4_regex_attachmentid *regexp.Regexp = regexp.MustCompile("^attachment[0-9]+$")

type baseCrawler struct {
	client        *http.Client
	cc            *CrawlContext
	cookie_setup  bool
	debug         bool
	debug_counter int
	dispatcher    *download.DownloadDispatcher
	yield         chan int
	excluded      []*url.URL
	redirect      func(*http.Request, []*http.Request) error
}

func newBaseCrawler(cc *CrawlContext) *baseCrawler {
	return &baseCrawler{
		cc: cc, client: new(http.Client),
		excluded: make([]*url.URL, 0, 1),
		redirect: logRedirect,
	}
}

func (c *baseCrawler) debug_DumpHeader(dir, name string, header http.Header) {
	filename := fmt.Sprintf("%d - %s.txt", c.debug_counter, name)
	c.debug_counter++
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error(fmt.Errorf("DumpHeader failed: %w", err))
		return
	}
	f, err := os.Create(path)
	if err != nil {
		log.Error(fmt.Errorf("DumpHeader failed: %w", err))
		return
	}
	defer f.Close()
	for k, v := range header {
		var b strings.Builder
		for _, vv := range v {
			b.WriteString(fmt.Sprintf("%s:\t", k))
			b.WriteString(vv)
			b.WriteByte('\n')
		}
		_, err := f.WriteString(b.String())
		if err != nil {
			log.Error(fmt.Errorf("DumpHeader failed: %w", err))
			return
		}
	}
}

//getPage receives an http response by issuing a "GET" request on url "page". This function has 3 side effects.
//At first the http client's CheckRedirect function is set to baseCrawler's "redirect" member. Secondly a new cookie jar
//is deployed to the http client if there isn't already one. Thirdly the cookie jar is filled with the CrawlContext's cookie slice,
//but only if the cookie jar did not exist before (i.e. on the first call).
func (c *baseCrawler) getPage(page *url.URL) (*http.Response, error) {
	c.client.CheckRedirect = c.redirect
	req, err := http.NewRequest("GET", page.String(), nil)
	if err != nil {
		return nil, err
	}

	//setup cookie jar if none exists
	if c.client.Jar == nil {
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}

		//load passed cookies into the cookie jar
		if len(c.cc.Cookies) > 0 {
			cookieurl, err := baseURLOnly(page)
			if err != nil {
				return nil, err
			}
			jar.SetCookies(cookieurl, c.cc.Cookies)
		}
		c.client.Jar = jar
	}
	cookies := c.client.Jar.Cookies(page)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	if c.debug {
		c.debug_DumpHeader(filepath.Join(c.cc.output, "debug"), "Request Header", req.Header)
	}
	resp, err := c.client.Do(req)
	if c.debug {
		c.debug_DumpHeader(filepath.Join(c.cc.output, "debug"), "Response Header", resp.Header)
	}
	return resp, err
}

//redirection sets the optional redirection handler function for the crawler's http.Client
func (c *baseCrawler) redirection(redirect func(*http.Request, []*http.Request) error) {
	c.client.CheckRedirect = redirect
}

func (c *baseCrawler) setup(jobs int) {
	c.dispatcher = download.NewDownloadDispatcher(jobs)
	c.yield = make(chan int)
	f := func() {
		for dl := c.dispatcher.Collect(); dl != nil; dl = c.dispatcher.Collect() {
			if dl.Err != nil {
				if err, ok := dl.Err.(download.RenameError); ok {
					log.Error(fmt.Errorf("%s: %s", err, err.Unwrap()))
				} else {
					log.Error(fmt.Errorf("Download failed %q: %w", dl.Addr.String(), dl.Err))
				}
			} else {
				log.Info(fmt.Sprintf("Download complete: %s → %s", dl.Addr.String(), dl.File()))
			}
		}
		c.yield <- 1
	}
	go f()
}

func (c *baseCrawler) SetOptions(args []string) error {
	set := flag.NewFlagSet("baseCrawler", flag.ContinueOnError)
	common := addCommonCrawlerFlags(set)
	if err := set.Parse(args); err != nil {
		return err
	}
	c.excluded = common.excludedURLs.URLs
	if *common.allowRedirect {
		c.redirect = logRedirect
	} else {
		c.redirect = noRedirect
	}
	c.debug = bool(*common.debugMode)
	return nil
}

func (c *baseCrawler) Setup() {
	c.setup(DEFAULT_DL_JOBS)
}

//Finish() is a default cleanup function for crawlers, If baseCrawler's Setup() or setup() method was used
//Finish() closes baseCrawler's DownloadDispatcher and yields until all Downloads have been finished.
//Otherwise it does nothing.
func (c *baseCrawler) Finish() {
	if c.yield != nil {
		c.dispatcher.Close()
		<-c.yield
	}
}

//FileCrawler is a crawler that treats every input from the pager as a file that needs to be downloaded.
type FileCrawler struct {
	*baseCrawler
}

func NewFileCrawler(cc *CrawlContext) (CrawlerInterface, error) {
	crawler := &FileCrawler{baseCrawler: newBaseCrawler(cc)}
	return crawler, nil
}

func (r *FileCrawler) Crawl(u *url.URL) error {
	var filename string
	page := r.cc.Pager.PageNum()
	name := fileNameFromURL(u)

	//determine filename for download
	if len(name) > 0 {
		filename = fmt.Sprintf("%d - %s", page, name)
	}

	//setup Download struct
	dl := &download.Download{Client: r.client, Addr: u}
	if err := dl.SetDir(r.cc.output); err != nil {
		return err
	}
	if len(filename) > 0 {
		dl.SetFile(filename)
	}
	//run download
	r.dispatcher.Dispatch(dl)
	return nil
}

type VBAttachmentCrawler struct {
	*baseCrawler
	headernames bool
	page        *url.URL
}

type vbpost html.Node
type vbattachment html.Node

func NewVBAttachmentCrawler(cc *CrawlContext) (CrawlerInterface, error) {
	crawler := &VBAttachmentCrawler{baseCrawler: newBaseCrawler(cc)}
	return crawler, nil
}

func (r *VBAttachmentCrawler) SetOptions(args []string) error {
	set := flag.NewFlagSet("VBAttachmentCrawler", flag.ContinueOnError)
	headernames := new(cmdline.Boolean)
	common := addCommonCrawlerFlags(set)
	set.Var(headernames, "names-from-header", "if true, the downloader will use the file names sent via the http header")
	if err := set.Parse(args); err != nil {
		return err
	}
	r.excluded = common.excludedURLs.URLs
	if *common.allowRedirect {
		r.redirect = logRedirect
	} else {
		r.redirect = noRedirect
	}
	r.debug = bool(*common.debugMode)
	r.headernames = bool(*headernames)
	return nil
}

func (r *VBAttachmentCrawler) Crawl(u *url.URL) error {
	r.page = u
	resp, err := r.getPage(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	posts := r.vb4PostList(body)
	if posts == nil {
		log.Error(fmt.Sprintf("No posts found at page %q", u.String()))
	}
	for _, post := range posts {
		atts := post.attachments()
		for _, att := range atts {
			attUrl, err := att.href()
			if err != nil {
				printFetchError(attUrl)
				continue
			}
			if !attUrl.IsAbs() {
				attUrl, err = rel2absURL(u, attUrl)
				if err != nil {
					printFetchError(attUrl)
					continue
				}
			}
			dl := &download.Download{Client: r.client, Addr: attUrl}

			//determine download filename
			postid := post.id()
			if r.headernames {
				dl.AfterDownload = download.ADNameFromHeader(postid)
			} else {
				name := fileNameFromURL(attUrl)
				if name == "" {
					printFetchError(attUrl)
					continue
				}
				dl.SetFile(fmt.Sprintf("%s - %s", postid, name))
			}
			//set download directory
			if err := dl.SetDir(r.cc.output); err != nil {
				log.Error(err)
				continue
			}

			//run download
			r.dispatcher.Dispatch(dl)
		}
	}
	return nil
}

func (r *VBAttachmentCrawler) vb4PostList(node *html.Node) []*vbpost {
	const searchForID string = "posts"
	posts := libhtml.ElementByID(node, searchForID)
	if posts == nil {
		return nil
	}
	nodes := libhtml.ElementsByAttrMatch(posts, "id", vb4_regex_postid)
	if len(nodes) == 0 {
		return nil
	}
	vbposts := make([]*vbpost, len(nodes))
	for i := range nodes {
		vbposts[i] = (*vbpost)(nodes[i])
		if log.Level() == logger.Level_Debug {
			log.Debug(fmt.Sprintf("VBAttachmentCrawler: found post %q", vbposts[i].id()))
		}
	}
	return vbposts
}

func (r *vbpost) id() string {
	for _, a := range r.Attr {
		if a.Key == "id" && vb4_regex_postid.MatchString(a.Val) {
			re := regexp.MustCompile(`[0-9]+`)
			return re.FindString(a.Val)
		}
	}
	panic("vbpost.id() did not find a post id, that should not have happened")
}

func (r *vbpost) attachments() []*vbattachment {
	nodes := libhtml.ElementsByAttrMatch((*html.Node)(r), "id", vb4_regex_attachmentid)
	vb4att := make([]*vbattachment, len(nodes))
	for i := range nodes {
		vb4att[i] = (*vbattachment)(nodes[i])
		if log.Level() == logger.Level_Debug {
			var id string
			for _, attr := range nodes[i].Attr {
				if attr.Key == "id" {
					id = attr.Val
					break
				}
			}
			log.Debug(fmt.Sprintf("VBAttachmentCrawler: Found attachment %q", id))
		}
	}
	return vb4att
}

func (r *vbattachment) href() (*url.URL, error) {
	for _, a := range r.Attr {
		if a.Key == "href" {
			url, err := url.Parse(a.Val)
			if err != nil {
				return nil, err
			}
			return url, nil
		}
	}
	return nil, nil
}

/* functions and types that can be used by all crawlers: */

type commonCrawlerFlags struct {
	excludedURLs  cmdline.URLCollection
	allowRedirect *cmdline.Boolean
	debugMode     *cmdline.Boolean
}

func addCommonCrawlerFlags(set *flag.FlagSet) *commonCrawlerFlags {
	res := commonCrawlerFlags{debugMode: new(cmdline.Boolean), allowRedirect: new(cmdline.Boolean)}
	*res.allowRedirect = cmdline.Boolean(true)
	set.Var(&res.excludedURLs, "exclude", "Comma-separated list of URLs that won't be downloaded")
	set.Var(res.allowRedirect, "redirect", "Allow or deny redirects")
	set.Var(res.debugMode, "debug", "Enable extra debugging code for the crawler")
	return &res
}

func cmdAttrs2htmlAttrs(attrs_cmd cmdline.Attrs) []html.Attribute {
	attrs_html := make([]html.Attribute, 0, 10)
	for key, vals := range attrs_cmd {
		for _, val := range vals {
			attr := html.Attribute{
				Key: key,
				Val: val,
			}
			attrs_html = append(attrs_html, attr)
		}
	}
	return attrs_html
}
