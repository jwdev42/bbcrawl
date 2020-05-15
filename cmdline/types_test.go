package cmdline

import (
	"strings"
	"testing"
)

func TestURLCollection(t *testing.T) {
	input := "https://www.google.com,ftp://example.com,relative/url/example.html"
	input_split := strings.Split(input, ",")
	urls := &URLCollection{}
	if urls.String() != "" {
		t.Logf("%s: Expected empty string.", t.Name())
		t.Fail()
	}
	if err := urls.Set(input); err != nil {
		t.Logf("%s: error in method \"Set\": %v", t.Name(), err)
		t.FailNow()
	}
	for i, url := range urls.URLs {
		if input_split[i] != url.String() {
			t.Logf("%s: expected %s==%s", t.Name(), input_split[i], url.String())
			t.FailNow()
		}
	}
	if input != urls.String() {
		t.Logf("%s: expected %s==%s", t.Name(), input, urls.String())
		t.FailNow()
	}
}

func TestSingleURL(t *testing.T) {
	input := "https://www.google.com"
	url := &SingleURL{}
	if url.String() != "" {
		t.Logf("%s: Expected empty string.", t.Name())
		t.Fail()
	}
	if err := url.Set(input); err != nil {
		t.Logf("%s: error in method \"Set\": %v", t.Name(), err)
		t.FailNow()
	}
	if input != url.String() {
		t.Logf("%s: expected %s==%s", t.Name(), input, url.String())
		t.FailNow()
	}
}

func TestFSDirectory(t *testing.T) {
	dir := "/var"
	nodir := "allyourbasearebelongtous"
	fsdir := &FSDirectory{}
	nofsdir := &FSDirectory{}
	if err := fsdir.Set(dir); err != nil {
		t.Logf("%s: %v.", t.Name(), err)
		t.Fail()
	}
	if err := nofsdir.Set(nodir); err == nil {
		t.Logf("%s: Expected an error.", t.Name())
		t.Fail()
	}
}
