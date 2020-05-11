package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/libcrawl"
	"os"
)

type cmdflags struct {
	thread string
	start  int
	end    int
	posts  int
	out    string
}

func cmdline() (*cmdflags, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fs := flag.NewFlagSet("main", flag.ExitOnError)
	flags := new(cmdflags)
	fs.IntVar(&flags.start, "s", -1, "page you want to start with")
	fs.IntVar(&flags.end, "e", -1, "page you want to end with")
	fs.IntVar(&flags.posts, "p", -1, "amount of posts on a page")
	fs.StringVar(&flags.out, "o", cwd, "output directory")
	fs.StringVar(&flags.thread, "t", "", "output directory")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	//validate input
	if flags.start < 1 {
		return nil, errors.New("start value not set or less then 1")
	}
	if flags.end < 1 {
		return nil, errors.New("end value not set or less then 1")
	}
	if flags.end < flags.start {
		return nil, errors.New("end cannot be greater than start")
	}
	if flags.posts < 1 {
		return nil, errors.New("amount of posts not set or less then 1")
	}
	if b, err := isdir(flags.out); !b {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("\"%s\" is not a directory", flags.out)
	}
	if flags.thread == "" {
		return nil, errors.New("no thread specified")
	}

	return flags, nil
}

func isdir(f string) (bool, error) {
	fp, err := os.Open(f)
	if err != nil {
		return false, err
	}
	defer fp.Close()
	fpi, err := fp.Stat()
	if err != nil {
		return false, err
	}
	return fpi.IsDir(), nil
}

func main() {
	flags, err := cmdline()
	if err != nil {
		fmt.Fprintf(os.Stderr, "command line failed: %v\n", err)
		os.Exit(2)
	}
	err = libcrawl.ImgCrawl(flags.out, flags.thread, flags.start, flags.end, flags.posts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crawler failed: %v\n", err)
		os.Exit(2)
	}
}
