/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"github.com/jwdev42/bbcrawl/global"
	"github.com/jwdev42/cookiefile"
	"github.com/jwdev42/logger"
	"net/http"
	"net/url"
)

const DEFAULT_DL_JOBS = 5

var log = global.GetLogger()

var pagers = map[string]func(*CrawlContext) PagerInterface{
	PAGER_VB4:    NewVB4Pager,
	PAGER_QUERY:  NewQueryPager,
	PAGER_URLCUT: NewURLCuttingPager,
}

var crawlers = map[string]func(*CrawlContext) (CrawlerInterface, error){
	CRAWLER_VB_ATTACHMENTS: NewVBAttachmentCrawler,
	CRAWLER_SRC:            NewSrcCrawler,
	CRAWLER_FILE:           NewFileCrawler,
}

type PagerInterface interface {
	Next() (*url.URL, error)
	PageNum() int
	SetOptions([]string) error
	SetUrl(string) error
}

type CrawlerInterface interface {
	Crawl(*url.URL) error
	Finish()
	SetOptions([]string) error
	Setup()
}

type CrawlContext struct {
	output  string
	Cookies []*http.Cookie
	Pager   PagerInterface
	Crawler CrawlerInterface
}

//Parse global options and attach them to the CrawlContext
func (cc *CrawlContext) SetOptions(args []string) error {
	flagSet := flag.NewFlagSet("GlobalOptions", flag.ContinueOnError)
	outputDir := &cmdline.FSDirectory{}
	flagSet.Var(outputDir, "o", "set the output directory")
	cf := flagSet.String("cookie-file", "", "load cookies from file")
	loglevel := logger.LevelFlag(global.Default_Loglevel)
	flagSet.Var(&loglevel, "loglevel", "set the least severe loglevel that will have its messages printed")
	if err := flagSet.Parse(args); err != nil {
		return err
	}
	if len(outputDir.Path) > 0 {
		cc.output = outputDir.Path
	}
	if len(*cf) > 0 {
		cookies, err := cookiefile.Load(*cf)
		if err != nil {
			return err
		}
		cc.Cookies = cookies
	}
	log.SetLevel(int(loglevel))
	return nil
}

func NewCrawlContext(pager string, crawler string, defaultDir string) (*CrawlContext, error) {
	var err error
	cc := &CrawlContext{
		output: defaultDir,
	}
	newPager := pagers[pager]
	if newPager == nil {
		return nil, fmt.Errorf("Pager not found: %q", pager)
	}
	cc.Pager = newPager(cc)
	newCrawler := crawlers[crawler]
	if newCrawler == nil {
		return nil, fmt.Errorf("Crawler not found: %q", crawler)
	}
	cc.Crawler, err = newCrawler(cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func Crawl(cc *CrawlContext) error {
	cc.Crawler.Setup()
	defer cc.Crawler.Finish()
	for url, err := cc.Pager.Next(); url != nil; {
		if err != nil {
			return err
		}
		if err := cc.Crawler.Crawl(url); err != nil {
			return err
		}
		url, err = cc.Pager.Next()
	}
	return nil
}
