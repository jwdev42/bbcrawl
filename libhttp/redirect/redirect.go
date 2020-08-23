/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package redirect

import (
	"fmt"
	"github.com/jwdev42/bbcrawl/global"
	"net/http"
)

const default_redirects = 10

var log = global.GetLogger()

func Deny(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		return fmt.Errorf("Attempted Redirection: %q → %q", lastReq.URL.String(), req.URL.String())
	}
	return nil
}

func Log(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		lastReq := via[len(via)-1]
		log.Notice(fmt.Sprintf("Redirection: %q → %q\n", lastReq.URL.String(), req.URL.String()))
	}
	if len(via) > default_redirects {
		return fmt.Errorf("Too many redirects")
	}
	return nil
}
