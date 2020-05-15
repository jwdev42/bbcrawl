package main

import (
	"errors"
	"flag"
	"github.com/jwdev42/bbcrawl/cmdline"
	"os"
)

type cmdflags struct {
	thread   *cmdline.SingleURL
	start    int
	end      int
	posts    int
	out      *cmdline.FSDirectory
	excluded *cmdline.URLCollection
}

func newCmdflags() *cmdflags {
	flags := new(cmdflags)
	flags.thread = new(cmdline.SingleURL)
	flags.out = new(cmdline.FSDirectory)
	flags.excluded = new(cmdline.URLCollection)
	return flags
}

func parseCmdline() (*cmdflags, error) {
	flags := newCmdflags()
	fs := flag.NewFlagSet("main", flag.ExitOnError)
	fs.IntVar(&flags.start, "s", -1, "page you want to start with")
	fs.IntVar(&flags.end, "e", -1, "page you want to end with")
	fs.IntVar(&flags.posts, "p", -1, "amount of posts on a page")
	fs.Var(flags.out, "o", "output directory")
	fs.Var(flags.thread, "t", "thread to dump")
	fs.Var(flags.excluded, "-exclude", "comma-seperated list of urls to ignore")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	return flags, nil
}

func validateCmdline(flags *cmdflags) error {
	if flags.start < 1 {
		errors.New("start value not set or less then 1")
	}
	if flags.end < 1 {
		errors.New("end value not set or less then 1")
	}
	if flags.end < flags.start {
		errors.New("end cannot be greater than start")
	}
	if flags.posts < 1 {
		errors.New("amount of posts not set or less then 1")
	}
	return nil
}
