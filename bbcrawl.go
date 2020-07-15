package main

import (
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"github.com/jwdev42/bbcrawl/global"
	"github.com/jwdev42/bbcrawl/libcrawl"
	"os"
	"time"
)

var log = global.GetLogger()

func eexit(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(2)
}

func parseCmdline() (*cmdline.Product, error) {
	l := &cmdline.Lexer{}
	if err := l.Analyze(os.Args[1:]); err != nil {
		return nil, err
	}
	p := cmdline.NewParser(l)
	res, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	log.SetTimeFormat(time.RFC1123)
	cmd, err := parseCmdline()
	if err != nil {
		eexit(fmt.Errorf("Command line parser error: %v", err))
	}
	workDir, err := os.Getwd()
	if err != nil {
		eexit(err)
	}
	cc, err := libcrawl.NewCrawlContext(cmd.Pager, cmd.Crawler, workDir)
	if err != nil {
		eexit(fmt.Errorf("Error creating CrawlContext: %v", err))
	}
	err = cc.SetOptions(cmd.GlobalFlags)
	if err != nil {
		eexit(fmt.Errorf("Error setting global flags: %v", err))
	}
	err = cc.Pager.SetOptions(cmd.PagerFlags)
	if err != nil {
		eexit(fmt.Errorf("Error setting pager flags: %v", err))
	}
	err = cc.Pager.SetUrl(cmd.Url)
	if err != nil {
		eexit(fmt.Errorf("Error setting Url: %v", err))
	}
	err = cc.Crawler.SetOptions(cmd.CrawlerFlags)
	if err != nil {
		eexit(fmt.Errorf("Error setting crawler flags: %v", err))
	}
	err = libcrawl.Crawl(cc)
	if err != nil {
		eexit(fmt.Errorf("Error: %v", err))
	}
}
