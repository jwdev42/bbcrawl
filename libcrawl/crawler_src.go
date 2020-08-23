/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"flag"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline"
	"github.com/jwdev42/bbcrawl/libcrawl/download"
	"github.com/jwdev42/bbcrawl/libhtml"
	"github.com/jwdev42/bbcrawl/libhttp"
	"github.com/jwdev42/bbcrawl/libhttp/redirect"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type SrcCrawler struct {
	*baseCrawler
	attrs  []html.Attribute
	atoms  []atom.Atom
	fileid int
}

func NewSrcCrawler(cc *CrawlContext) (CrawlerInterface, error) {
	crawler := &SrcCrawler{
		baseCrawler: newBaseCrawler(cc),
	}
	return crawler, nil
}

func (r *SrcCrawler) Crawl(u *url.URL) error {
	const attr_src = "src"
	r.fileid = 1
	resp, err := r.getPage(u)
	if err != nil {
		return err
	}
	body, err := libhttp.BodyUTF8(resp)
	if err != nil {
		return err
	}
	document, err := html.Parse(body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	nodes := libhtml.ElementsByTag(document, r.atoms...)
	for _, n := range nodes {
		switch n.DataAtom {
		case atom.Audio:
			if r.hasAtom(n.DataAtom) && libhtml.MatchAttrs(n, r.attrs...) {
				if err := r.scrapeAV(u, n); err != nil {
					log.Error(fmt.Errorf("Download error: %v", err))
				}
			}
		case atom.Img:
			if r.hasAtom(n.DataAtom) && libhtml.MatchAttrs(n, r.attrs...) {
				link := libhtml.AttrVal(n, attr_src)
				if len(link) > 0 {
					name, err := r.uniqueName(link)
					if err != nil {
						log.Error(fmt.Errorf("Download error: %v", err))
						break
					}
					if err := r.download(u, link, r.cc.output, name); err != nil {
						log.Error(fmt.Errorf("Download error: %v", err))
					}
				}
			}
		case atom.Video:
			if r.hasAtom(n.DataAtom) && libhtml.MatchAttrs(n, r.attrs...) {
				if err := r.scrapeAV(u, n); err != nil {
					log.Error(fmt.Errorf("Download error: %v", err))
				}
			}
		default:
			panic("You're not supposed to be here!")
		}
	}
	return nil
}

func (r *SrcCrawler) SetOptions(args []string) error {
	set := flag.NewFlagSet("SrcCrawler", flag.ContinueOnError)
	common := addCommonCrawlerFlags(set)
	cmdattrs := make(cmdline.Attrs)
	set.Var(cmdattrs, "attrs", "Download only images that match the declared node attributes")
	taglist := cmdline.NewStringWhitelist(",", "audio", "img", "video")
	set.Var(taglist, "tags", "Download sources contained within the given tags")
	if err := set.Parse(args); err != nil {
		return err
	}
	r.excluded = common.excludedURLs.URLs
	if *common.allowRedirect {
		r.redirect = redirect.Log
	} else {
		r.redirect = redirect.Deny
	}
	r.debug = bool(*common.debugMode)
	r.attrs = cmdAttrs2htmlAttrs(cmdattrs)
	if len(taglist.Result()) == 0 {
		return fmt.Errorf("No html tag specified with \"-tags\"")
	}
	r.atoms = r.tags2atoms(taglist.Result())
	return nil
}

func (r *SrcCrawler) download(page *url.URL, link, dir, name string) error {
	if link == "" {
		panic("link must not be empty")
	}
	u, err := url.Parse(link)
	if err != nil {
		return err
	}
	if !u.IsAbs() {
		u, err = rel2absURL(page, u)
		if err != nil {
			return err
		}
	}
	dl := &download.Download{
		Client: r.client,
		Addr:   u,
	}
	if err := dl.SetDir(dir); err != nil {
		return err
	}
	dl.SetFile(name)
	r.dispatcher.Dispatch(dl)
	return nil
}

//scrapeAV searches for sources inside an audio or video tag and its children and downloads them
func (r *SrcCrawler) scrapeAV(page *url.URL, node *html.Node) error {
	const attr_src = "src"
	downloads := make([]string, 0, 5)
	root := libhtml.AttrVal(node, attr_src)
	if len(root) > 0 {
		downloads = append(downloads, root)
	}
	children := libhtml.ElementsByTag(node, atom.Source, atom.Track)
	for _, child := range children {
		link := libhtml.AttrVal(child, attr_src)
		if len(link) > 0 {
			downloads = append(downloads, link)
		}
	}
	switch len(downloads) {
	case 0:
		return nil
	case 1:
		name, err := r.uniqueName(downloads[0])
		if err != nil {
			log.Error(fmt.Errorf("Download error: %v", err))
			break
		}
		if err := r.download(page, downloads[0], r.cc.output, name); err != nil {
			log.Error(fmt.Errorf("Download error: %v", err))
		}
	default:
		dir := filepath.Join(r.cc.output, fmt.Sprintf("%d-%d", r.cc.Pager.PageNum(), r.fileid))
		r.fileid++
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}

		sources := make(avTag)
		for _, link := range downloads {
			if err := sources.addSrc(link); err != nil {
				log.Error(fmt.Errorf("Download error: %v", err))
			}
		}
		for link, name := range sources {
			if err := r.download(page, link, dir, name); err != nil {
				log.Error(fmt.Errorf("Download error: %v", err))
			}
		}
	}
	return nil
}

//uniqueName constructs a unique file name by extracting the input url's file extension and combining it with a unique string
func (r *SrcCrawler) uniqueName(s string) (string, error) {
	var suffix string
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	li := strings.LastIndex(u.Path, ".")
	if li+1 < len(u.Path) {
		suffix = u.Path[li+1:]
	} else {
		return "", fmt.Errorf("No suffix available in url path \"%s\"", u.Path)
	}
	fid := r.fileid
	r.fileid++
	return fmt.Sprintf("%d-%d.%s", r.cc.Pager.PageNum(), fid, suffix), nil
}

func (r *SrcCrawler) hasAtom(atom atom.Atom) bool {
	for _, a := range r.atoms {
		if a == atom {
			return true
		}
	}
	return false
}

func (r *SrcCrawler) isExcluded(url *url.URL) bool {
	for _, exurl := range r.excluded {
		if exurl.String() == url.String() {
			return true
		}
	}
	return false
}

func (r *SrcCrawler) tags2atoms(tags []string) []atom.Atom {
	atoms := make([]atom.Atom, 0, len(tags))
	for _, tag := range tags {
		switch tag {
		case "audio":
			atoms = append(atoms, atom.Audio)
		case "img":
			atoms = append(atoms, atom.Img)
		case "video":
			atoms = append(atoms, atom.Video)
		default:
			panic(fmt.Errorf("Unknown or invalid tag: \"%s\"", tag))
		}
	}
	return atoms
}
