package cmdline

import (
	"strings"
	"testing"
)

func Test_Lexer_isBareArg(t *testing.T) {
	tests := make(map[string]bool)
	tests["-test"] = false
	tests[""] = false
	tests["http://www.google.com"] = false
	tests["https://www.google.com"] = false
	tests["ftp://google.com"] = true
	tests["google.com"] = true
	tests["val"] = true
	tests["--gnu-style"] = false
	tests["林間"] = true
	l := &Lexer{}
	for k, v := range tests {
		if l.isBareArg(k) != v {
			t.Logf("%s: Expected %t on input %q", t.Name(), v, k)
			t.Fail()
		}
	}
}

func Test_Lexer_isSwitch(t *testing.T) {
	tests := make(map[string]bool)
	tests["-test"] = true
	tests["-23"] = true
	tests["-林間"] = true
	tests["http://www.google.com"] = false
	tests["val"] = false
	tests["--gnu-style"] = false
	tests["-pager"] = false
	tests["-crawler"] = false
	tests[""] = false
	l := &Lexer{}
	for k, v := range tests {
		if l.isSwitch(k) != v {
			t.Logf("%s: Expected %t on input %q", t.Name(), v, k)
			t.Fail()
		}
	}
}

func Test_Parser_Parse(t *testing.T) {
	validinput := strings.Split("-arg1 yes -arg2 no -pager testpager -arg3 hello -arg4 there -crawler testcrawler -depth deep -height high http://google.com", " ")
	lexx := &Lexer{}
	if err := lexx.Analyze(validinput); err != nil {
		t.Logf("%s: Error during lexical analysis: %s", t.Name(), err)
		t.FailNow()
	}
	p := NewParser(lexx)
	res, err := p.Parse()
	if err != nil {
		t.Logf("%s: Parser error: %s", t.Name(), err)
		t.FailNow()
	}
	if res.Pager != validinput[5] {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[5], res.Pager)
		t.Fail()
	}
	if res.Crawler != validinput[11] {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[11], res.Crawler)
		t.Fail()
	}
	if strings.Join(res.GlobalFlags, " ") != strings.Join(validinput[0:4], " ") {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[0:4], res.GlobalFlags)
		t.Fail()
	}
	if strings.Join(res.PagerFlags, " ") != strings.Join(validinput[6:10], " ") {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[6:10], res.PagerFlags)
		t.Fail()
	}
	if strings.Join(res.CrawlerFlags, " ") != strings.Join(validinput[12:16], " ") {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[12:16], res.CrawlerFlags)
		t.Fail()
	}
	if res.Url != validinput[16] {
		t.Logf("%s: Parsed Field: expected %q, got %q", t.Name(), validinput[16], res.Url)
		t.Fail()
	}
}
