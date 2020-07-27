/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"fmt"
	"net/http"
	"net/url"
)

func noRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		return fmt.Errorf("Attempted Redirection: %q → %q", lastReq.URL.String(), req.URL.String())
	}
	return nil
}

func logRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		log.Notice(fmt.Sprintf("Redirection: %q → %q\n", lastReq.URL.String(), req.URL.String()))
	}
	if len(via) > DEFAULT_REDIRECTS {
		return fmt.Errorf("Too many redirects")
	}
	return nil
}

func printFetchError(url *url.URL) {
	//BUG(jw): Password needs to be filtered out of the url before printing it.
	log.Error(fmt.Sprintf("File %q could not be downloaded.", url.String()))
}
