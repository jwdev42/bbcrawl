/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package cmdline

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	TOKEN_INVALID = iota
	TOKEN_BAREARG
	TOKEN_SWITCH
	TOKEN_SWITCH_PAGER
	TOKEN_SWITCH_CRAWLER
	TOKEN_URL
	TOKEN_ARG
	TOKEN_EOS
)

const (
	LIT_MINUS = "-"
)

type token struct {
	t   int //Token type, one of the TOKEN_* constants
	val string
}

func tokenType2String(t int) string {
	switch t {
	case TOKEN_INVALID:
		return "INVALID"
	case TOKEN_BAREARG:
		return "BAREARG"
	case TOKEN_SWITCH:
		return "SWITCH"
	case TOKEN_SWITCH_PAGER:
		return "SWITCH_PAGER"
	case TOKEN_SWITCH_CRAWLER:
		return "SWITCH_CRAWLER"
	case TOKEN_URL:
		return "URL"
	case TOKEN_ARG:
		return "ARG"
	case TOKEN_EOS:
		return "EOS"
	}
	return "UNDEFINED"
}

type Lexer struct {
	tokens []token
}

func (l *Lexer) Analyze(args []string) error {
	var t int
	for i, arg := range args {
		if l.isBareArg(arg) {
			t = TOKEN_BAREARG
		} else if l.isSwitch(arg) {
			t = TOKEN_SWITCH
		} else if l.isSwitchPager(arg) {
			t = TOKEN_SWITCH_PAGER
		} else if l.isSwitchCrawler(arg) {
			t = TOKEN_SWITCH_CRAWLER
		} else if l.isUrl(arg) {
			t = TOKEN_URL
		} else {
			return fmt.Errorf("Position %d: Invalid token %q", i, arg)
		}
		l.tokens = append(l.tokens, token{t, arg})
	}
	l.tokens = append(l.tokens, token{TOKEN_EOS, ""})
	return nil
}

func (l *Lexer) isBareArg(s string) bool {
	if len(s) == 0 {
		return false
	}
	if strings.HasPrefix(s, LIT_MINUS) {
		return false
	}
	if l.isUrl(s) {
		return false
	}
	return true
}

func (l *Lexer) isSwitch(s string) bool {
	if len(s) < 2 || !strings.HasPrefix(s, LIT_MINUS) || string(s[1]) == LIT_MINUS || l.isSwitchPager(s) || l.isSwitchCrawler(s) {
		return false
	}
	return true
}

func (l *Lexer) isSwitchPager(s string) bool {
	if s == "-pager" {
		return true
	}
	return false
}

func (l *Lexer) isSwitchCrawler(s string) bool {
	if s == "-crawler" {
		return true
	}
	return false
}

func (l *Lexer) isUrl(s string) bool {
	u, err := url.Parse(s)
	if err != nil || !u.IsAbs() || u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return true
}

func parsingError(t token, expected int, pos int) error {
	return fmt.Errorf("Token %q, position %d: Expected %q, got %q", t.val, pos, tokenType2String(expected), tokenType2String(t.t))
}

type Parser struct {
	tokens []token
	pos    int
}

type Product struct {
	GlobalFlags  []string
	Crawler      string
	CrawlerFlags []string
	Pager        string
	PagerFlags   []string
	Url          string
}

func NewParser(l *Lexer) *Parser {
	return &Parser{l.tokens, 0}
}

func (p *Parser) getToken() token {
	if len(p.tokens) > p.pos {
		return p.tokens[p.pos]
	}
	panic("One does not simply receive tokens beyond their last element!")
}

func (p *Parser) advanceTokens() {
	if len(p.tokens) > p.pos+1 {
		p.pos++
		return
	}
	panic("One does not simply advance tokens beyond their last element!")
}

func (p *Parser) peek() token {
	if len(p.tokens) > p.pos+1 {
		return p.tokens[p.pos+1]
	}
	panic("One does not simply peek tokens beyond their last element!")
}

func (p *Parser) parseSwitch() (string, error) {
	tk := p.getToken()
	if tk.t != TOKEN_SWITCH {
		return "", parsingError(tk, TOKEN_SWITCH, p.pos)
	}
	p.advanceTokens()
	return tk.val, nil
}

func (p *Parser) parseBareArg() (string, error) {
	tk := p.getToken()
	if tk.t != TOKEN_BAREARG {
		return "", parsingError(tk, TOKEN_BAREARG, p.pos)
	}
	p.advanceTokens()
	return tk.val, nil
}

func (p *Parser) parseUrl() (string, error) {
	tk := p.getToken()
	if tk.t != TOKEN_URL {
		return "", parsingError(tk, TOKEN_URL, p.pos)
	}
	nt := p.peek()
	if nt.t == TOKEN_EOS {
		return "", fmt.Errorf("Token %q, position %d: Did not expect %q at position %d", tokenType2String(TOKEN_URL), p.pos, tokenType2String(TOKEN_EOS), p.pos+1)
	}
	p.advanceTokens()
	return tk.val, nil
}

func (p *Parser) parseArg() (string, error) {
	tk := p.getToken()
	switch tk.t {
	case TOKEN_BAREARG:
		return p.parseBareArg()
	case TOKEN_URL:
		return p.parseUrl()
	}
	return "", parsingError(tk, TOKEN_ARG, p.pos)
}

func (p *Parser) parseFlag() ([]string, error) {
	start := p.pos
	sw, err := p.parseSwitch()
	if err != nil {
		return nil, err
	}
	arg, err := p.parseArg()
	if err != nil {
		p.pos = start
		return nil, err
	}
	return []string{sw, arg}, nil
}

func (p *Parser) parseFlags() ([]string, error) {
	tk := p.getToken()
	args := make([]string, 0, 10)
	for tk.t == TOKEN_SWITCH {
		fs, err := p.parseFlag()
		if err != nil {
			return args, err
		}
		for _, f := range fs {
			args = append(args, f)
		}
		tk = p.getToken()
	}
	return args, nil
}

func (p *Parser) parsePager() (string, error) {
	start := p.pos
	tk := p.getToken()
	if tk.t != TOKEN_SWITCH_PAGER {
		return "", parsingError(tk, TOKEN_SWITCH_PAGER, p.pos)
	}
	p.advanceTokens()
	arg, err := p.parseArg()
	if err != nil {
		p.pos = start
		return "", err
	}
	return arg, nil
}

func (p *Parser) parseCrawler() (string, error) {
	start := p.pos
	tk := p.getToken()
	if tk.t != TOKEN_SWITCH_CRAWLER {
		return "", parsingError(tk, TOKEN_SWITCH_CRAWLER, p.pos)
	}
	p.advanceTokens()
	arg, err := p.parseArg()
	if err != nil {
		p.pos = start
		return "", err
	}
	return arg, nil
}

func (p *Parser) parseGlobalSet() (string, []string, error) {
	args, err := p.parseFlags()
	if err != nil {
		return "", nil, err
	}
	pager, err := p.parsePager()
	if err != nil {
		return "", nil, err
	}
	return pager, args, nil
}

func (p *Parser) parsePagerSet() (string, []string, error) {
	args, err := p.parseFlags()
	if err != nil {
		return "", nil, err
	}
	crawler, err := p.parseCrawler()
	if err != nil {
		return "", nil, err
	}
	return crawler, args, nil
}

func (p *Parser) parseCrawlerSet() ([]string, error) {
	args, err := p.parseFlags()
	if err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseThreadSet() (string, error) {
	tk := p.getToken()
	if tk.t != TOKEN_URL {
		return "", parsingError(tk, TOKEN_URL, p.pos)
	}
	nt := p.peek()
	if nt.t != TOKEN_EOS {
		return "", parsingError(nt, TOKEN_EOS, p.pos+1)
	}
	p.advanceTokens()
	return tk.val, nil
}

func (p *Parser) Parse() (*Product, error) {
	result := Product{}
	pager, globalArgs, err := p.parseGlobalSet()
	if err != nil {
		return nil, err
	}
	result.Pager = pager
	result.GlobalFlags = globalArgs
	crawler, pagerArgs, err := p.parsePagerSet()
	if err != nil {
		return nil, err
	}
	result.Crawler = crawler
	result.PagerFlags = pagerArgs
	crawlerArgs, err := p.parseCrawlerSet()
	if err != nil {
		return nil, err
	}
	result.CrawlerFlags = crawlerArgs
	thread, err := p.parseThreadSet()
	if err != nil {
		return nil, err
	}
	result.Url = thread
	return &result, nil
}
