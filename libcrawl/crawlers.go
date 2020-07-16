package libcrawl

import (
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"github.com/jwdev42/bbcrawl/libcrawl/download"
	"github.com/jwdev42/logger"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"path/filepath"
	"os"
)

const (
	CRAWLER_VB4_ATTACHMENTS = "vb4-attachments"
	CRAWLER_IMAGE           = "img"
	CRAWLER_FILE            = "file"
)

var vb4_regex_postid *regexp.Regexp = regexp.MustCompile("^post_?[0-9]+$")
var vb4_regex_attachmentid *regexp.Regexp = regexp.MustCompile("^attachment[0-9]+$")

type baseCrawler struct {
	client        *http.Client
	cc            *CrawlContext
	cookie_setup  bool
	debug         bool
	debug_counter int
	download_jobs int
	excluded      []*url.URL
	redirect      func(*http.Request, []*http.Request) error
}

func newBaseCrawler(cc *CrawlContext) *baseCrawler {
	return &baseCrawler{
		cc: cc, client: new(http.Client),
		download_jobs: DEFAULT_DL_JOBS,
		excluded:      make([]*url.URL, 0, 1),
		redirect:      logRedirect,
	}
}

//checkDownloads loops through finished downloads and prints to the logger whether the download was successful or not.
func (c *baseCrawler) checkDownloads(downloads []*download.Download) {
	for _, dl := range downloads {
		if dl.Err != nil {
			log.Error(fmt.Errorf("Download failed: %w: %s", dl.Err, dl.Addr.String()))
		} else {
			log.Info(fmt.Sprintf("Download finished: %s", dl.Addr.String()))
		}
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
			b.WriteString(fmt.Sprintf("%s:\t",k))
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

//getPage receives a http response by issuing a "GET" request on url "page". This function has 3 side effects.
//Firstly the http client's CheckRedirect function is set to the crawler's "redirect" member. Secondly a new cookie jar
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
	if len(name) > 0 {
		filename = fmt.Sprintf("%d - %s", page, name)
	}
	disp := download.NewDownloadDispatcher(1)
	dl := &download.Download{Client: r.client, Addr: u, File: fmt.Sprintf("%s/%s", r.cc.output, filename)}
	disp.Dispatch(dl)
	disp.Close()
	r.checkDownloads(disp.Collect())
	return nil
}

type ImageCrawler struct {
	*baseCrawler
	attrs []html.Attribute
}

func NewImageCrawler(cc *CrawlContext) (CrawlerInterface, error) {
	crawler := &ImageCrawler{
		baseCrawler: newBaseCrawler(cc),
	}
	return crawler, nil
}

func (r *ImageCrawler) SetOptions(args []string) error {
	set := flag.NewFlagSet("ImageCrawler", flag.ContinueOnError)
	common := addCommonCrawlerFlags(set)
	cmd_attrs := make(cmdline.Attrs)
	set.Var(cmd_attrs, "attrs", "Download only images that match the declared node attributes")
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
	r.attrs = cmdAttrs2htmlAttrs(cmd_attrs)
	return nil
}

func (r *ImageCrawler) Crawl(url *url.URL) error {
	const imgtag string = "img"
	picid := 1
	page := r.cc.Pager.PageNum()
	resp, err := r.getPage(url)
	if err != nil {
		return err
	}
	body, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var nodes []*html.Node
	if len(r.attrs) == 0 {
		nodes = elementsByTag(body, atom.Img)
	} else {
		nodes = elementsByTagAndAttrs(body, imgtag, r.attrs)
	}
	dispatcher := download.NewDownloadDispatcher(r.download_jobs)
	for _, n := range nodes {
		for _, a := range n.Attr {
			if a.Key == "src" {
				li := strings.LastIndex(a.Val, ".")
				var suffix string
				if li+1 < len(a.Val) {
					suffix = a.Val[li+1:]
				} else {
					log.Println(logger.Level_Error, fmt.Errorf("Download error (no image suffix): %s", a.Val))
					break
				}
				if len(r.cc.output) < 1 {
					panic("Output directory missing")
				}
				dl := &download.Download{
					Client: r.client,
				}
				dl.File = fmt.Sprintf("%s/%d-%d.%s", r.cc.output, page, picid, suffix)
				if dl.Addr, err = url.Parse(a.Val); err != nil {
					log.Println(logger.Level_Error, fmt.Errorf("Download error: %w", err))
					break
				}
				if !dl.Addr.IsAbs() {
					dl.Addr, err = rel2absURL(url, dl.Addr)
					if err != nil {
						log.Println(logger.Level_Error, fmt.Errorf("Download error: %w", err))
						break
					}
				}
				if r.isExcluded(dl.Addr) {
					break
				}
				dispatcher.Dispatch(dl)
				picid++
				break
			}
		}
	}
	dispatcher.Close()
	r.checkDownloads(dispatcher.Collect())
	return nil
}

func (r *ImageCrawler) isExcluded(url *url.URL) bool {
	for _, exurl := range r.excluded {
		if exurl.String() == url.String() {
			return true
		}
	}
	return false
}

type VB4AttachmentCrawler struct {
	*baseCrawler
}

type vb4post html.Node
type vb4attachment html.Node

func NewVB4AttachmentCrawler(cc *CrawlContext) (CrawlerInterface, error) {
	crawler := &VB4AttachmentCrawler{baseCrawler: newBaseCrawler(cc)}
	return crawler, nil
}

func (r *VB4AttachmentCrawler) SetOptions(args []string) error {
	set := flag.NewFlagSet("VB4AttachmentCrawler", flag.ContinueOnError)
	common := addCommonCrawlerFlags(set)
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
	return nil
}

func (r *VB4AttachmentCrawler) Crawl(url *url.URL) error {
	resp, err := r.getPage(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	dispatcher := download.NewDownloadDispatcher(r.download_jobs)
	posts := r.vb4PostList(body)
	for _, post := range posts {
		atts := post.attachments()
		for _, att := range atts {
			attUrl, err := att.href()
			if err != nil {
				printFetchError(attUrl)
				continue
			}
			if !attUrl.IsAbs() {
				attUrl, err = rel2absURL(url, attUrl)
				if err != nil {
					printFetchError(attUrl)
					continue
				}
			}
			dl := &download.Download{Client: r.client, Addr: attUrl}
			name := fileNameFromURL(attUrl)
			if name == "" {
				printFetchError(attUrl)
				continue
			}
			dl.File = fmt.Sprintf("%s/%s", r.cc.output, name)
			dispatcher.Dispatch(dl)
		}
	}
	dispatcher.Close()
	r.checkDownloads(dispatcher.Collect())
	return nil
}

func (r *VB4AttachmentCrawler) vb4PostList(node *html.Node) []*vb4post {
	const searchForID string = "posts"
	posts := elementByID(node, searchForID)
	if posts == nil {
		return nil
		//TODO: log info that no posts have been found
	}
	nc := elementsByAttrMatch(posts, "id", vb4_regex_postid)
	vb4posts := make([]*vb4post, len(nc.nodes))
	for i := range nc.nodes {
		vb4posts[i] = (*vb4post)(nc.nodes[i])
		if log.Level() == logger.Level_Debug {
			page := r.cc.Pager.PageNum()
			log.Debug(fmt.Sprintf("VB4AttachmentCrawler: Page %d, found post %q", page, vb4posts[i].id()))
		}
	}
	return vb4posts
}

func (r *vb4post) id() string {
	for _, a := range r.Attr {
		if a.Key == "id" && vb4_regex_postid.MatchString(a.Val) {
			re := regexp.MustCompile(`[0-9]+`)
			return re.FindString(a.Val)
		}
	}
	return ""
}

func (r *vb4post) attachments() []*vb4attachment {
	nc := elementsByAttrMatch((*html.Node)(r), "id", vb4_regex_attachmentid)
	vb4att := make([]*vb4attachment, len(nc.nodes))
	for i := range nc.nodes {
		vb4att[i] = (*vb4attachment)(nc.nodes[i])
		if log.Level() == logger.Level_Debug {
			var id string
			for _, attr := range nc.nodes[i].Attr {
				if attr.Key == "id" {
					id = attr.Val
					break
				}
			}
			log.Debug(fmt.Sprintf("VB4AttachmentCrawler: Found attachment %q", id))
		}
	}
	return vb4att
}

func (r *vb4attachment) href() (*url.URL, error) {
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
	res := commonCrawlerFlags{ debugMode: new(cmdline.Boolean), allowRedirect: new(cmdline.Boolean) }
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
