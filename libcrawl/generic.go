package libcrawl

import (
	"net/url"
)

type PagerInterface interface {
	Next() (*url.URL, error)
}

type CrawlerInterface interface {
	Crawl(*url.URL) error
}

type CrawlContext struct {
	Pager   PagerInterface
	Crawler CrawlerInterface
}

func Crawl(cc *CrawlContext) error {
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
