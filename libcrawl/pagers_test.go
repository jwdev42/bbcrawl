/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"fmt"
	"strings"
	"testing"
)

func genericURLCuttingPagertest(t *testing.T, addr, addrfmt, cmdline string) {
	var startpage bool
	var startpage_str string
	options := strings.Split(cmdline, " ")
	pager := new(URLCuttingPager)
	if err := pager.SetOptions(options); err != nil {
		t.Logf("%v", err)
		t.FailNow()
	}
	if pager.startpage != nil {
		startpage = true
		startpage_str = pager.startpage.String()
	}
	if err := pager.SetUrl(addr); err != nil {
		t.Logf("%v", err)
		t.FailNow()
	}
	i := pager.page
	for u, err := pager.Next(); u != nil; u, err = pager.Next() {
		if err != nil {
			t.Logf("%v", err)
			t.FailNow()
		}
		if startpage {
			if u.String() != startpage_str {
				t.Errorf("Expected %q, got %q", startpage_str, u.String())
			}
			startpage = false
		} else {
			if u.String() != fmt.Sprintf(addrfmt, i) {
				t.Errorf("Expected %q, got %q", fmt.Sprintf(addrfmt, i), u.String())
			}
			i++
		}
	}
}

func TestURLCuttingPager(t *testing.T) {
	genericURLCuttingPagertest(t, "http://www.example.net/1/test", "http://www.example.net/%d/test", "-start 1 -end 100 -cut 24,24")
	genericURLCuttingPagertest(t, "http://www.example.net/1/test", "http://www.example.net/%05d/test", "-start 1 -end 100 -cut 24,24 -digits 5")
	genericURLCuttingPagertest(t, "http://www.example.net/1", "http://www.example.net/%d", "-start 1 -end 100 -cut 24,24")
	genericURLCuttingPagertest(t, "http://www.example.net/1", "http://www.example.net/%05d", "-start 1 -end 100 -cut 24,24 -digits 5")
	genericURLCuttingPagertest(t, "http://www.example.net/1", "http://www.example.net/%d", "-start 1 -end 100 -cut 24,25")
	genericURLCuttingPagertest(t, "http://www.example.net/1/", "http://www.example.net/%d/", "-start 1 -end 100 -cut 24,24")
	genericURLCuttingPagertest(t, "http://www.example.net/1/", "http://www.example.net/%d/", "-startpage http://www.example.net -start 1 -end 100 -cut 24,24")
}
