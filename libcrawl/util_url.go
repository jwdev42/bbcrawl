/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"fmt"
	"net/url"
)

func fileNameFromURL(url *url.URL) string {
	return smallestSubstrRight(url.Path, "/")
}

//baseURLOnly returns a new url that has the same host and the same scheme as the src argument, but no path and no query string.
//Panics if the src argument is a relative url or nil.
func baseURLOnly(src *url.URL) (*url.URL, error) {
	if len(src.Hostname()) == 0 {
		panic("relative URLs are not supported!")
	}
	return url.Parse(fmt.Sprintf("%s://%s", src.Scheme, src.Hostname()))
}

func rel2absURL(domain *url.URL, url *url.URL) (*url.URL, error) {
	if url.IsAbs() {
		panic("url parameter is absolute")
	}
	if !domain.IsAbs() {
		panic("domain parameter is relative")
	}
	var fmtstr string
	requrl := url.RequestURI()
	if requrl[0] == '/' {
		fmtstr = "%s://%s%s"
	} else {
		fmtstr = "%s://%s/%s"
	}
	if newurl, err := url.Parse(fmt.Sprintf(fmtstr, domain.Scheme, domain.Hostname(), url.RequestURI())); err != nil {
		return nil, err
	} else {
		return newurl, nil
	}
}

//Standard url validation function a pager's SetUrl-method can call
func url_for_pager(addr string) (*url.URL, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, fmt.Errorf("%q is not an absolute URL", addr)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%q is an unsupported url scheme", addr)
	}
	return u, nil
}
