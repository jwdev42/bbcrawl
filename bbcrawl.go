package main

import (
	"fmt"
	"github.com/jwdev42/bbcrawl/libcrawl"
	"net/url"
	"os"
)

func eexit(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(2)
}

func main() {
	flags, err := parseCmdline()
	if err != nil {
		eexit(fmt.Errorf("command line parser error: %v", err))
	}
	if err := validateCmdline(flags); err != nil {
		eexit(fmt.Errorf("command line validation error: %v", err))
	}
	err = crawlUnknownThumb(flags.out.Path, flags.thread.URL, flags.start, flags.end, flags.posts)
	if err != nil {
		eexit(fmt.Errorf("crawler failed: %v\n", err))
	}
}

func crawlUnknownThumb(outputDir string, thread *url.URL, firstPage int, lastPage int, posts int) error {
	if len(outputDir) == 0 {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	context := &libcrawl.CrawlContext{
		Pager: &libcrawl.UnknownBBPager{
			Start:  firstPage,
			End:    lastPage,
			Posts:  posts,
			Thread: thread,
		},
		Crawler: &libcrawl.ThumbCrawler{
			Out:  outputDir,
			Page: firstPage,
		},
	}
	return libcrawl.Crawl(context)
}
