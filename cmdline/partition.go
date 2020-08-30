/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package cmdline

import (
	"errors"
	"fmt"
	"strings"
)

type Product struct {
	GlobalFlags  []string
	Crawler      string
	CrawlerFlags []string
	Pager        string
	PagerFlags   []string
	Url          string
}

func (r *Product) String() string {
	const sep = " "
	builder := make([]string, 0, len(r.GlobalFlags)+len(r.CrawlerFlags)+len(r.PagerFlags)+5)
	appendAll := func(s []string, a []string) []string {
		for _, v := range a {
			s = append(s, v)
		}
		return s
	}
	builder = appendAll(builder, r.GlobalFlags)
	if r.Pager != "" {
		builder = append(builder, "-pager")
		builder = append(builder, r.Pager)
		builder = appendAll(builder, r.PagerFlags)
	}
	if r.Crawler != "" {
		builder = append(builder, "-crawler")
		builder = append(builder, r.Crawler)
		builder = appendAll(builder, r.CrawlerFlags)
	}
	builder = append(builder, r.Url)
	return strings.Join(builder, sep)
}

func Partition(cmdln []string) (*Product, error) {
	findItem := func(item string, items []string) int {
		for i, v := range items {
			if v == item {
				return i
			}
		}
		return -1
	}
	//oor checks if index is out of range
	oor := func(index int, values []string) bool {
		if len(values) > index {
			return false
		}
		return true
	}
	if len(cmdln) < 2 {
		return nil, errors.New("Empty command line")
	}
	product := new(Product)
	args := cmdln[1:]
	var index int

	if index = findItem("-pager", args); index < 0 || oor(index+1, args) {
		return nil, errors.New("No pager found")
	}
	product.GlobalFlags = args[0:index]
	product.Pager = args[index+1]

	if index += 2; oor(index, args) {
		return nil, fmt.Errorf("Unexpected EOS after \"%s\"", product.Pager)
	}

	args = args[index:]
	if index = findItem("-crawler", args); index < 0 || oor(index+1, args) {
		return nil, errors.New("No crawler found")
	}
	product.PagerFlags = args[0:index]
	product.Crawler = args[index+1]

	if index += 2; oor(index, args) {
		return nil, fmt.Errorf("Unexpected EOS after \"%s\"", product.Crawler)
	}

	args = args[index:]

	switch len(args) {
	case 0:
		return nil, errors.New("No URL found")
	case 1:
		product.Url = args[0]
	default:
		product.CrawlerFlags = args[0 : len(args)-1]
		product.Url = args[len(args)-1]
	}
	return product, nil
}
