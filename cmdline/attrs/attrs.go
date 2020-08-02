/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package attrs

import (
	"fmt"
	"strings"
)

const (
	token_text = iota
	token_split
	token_escape
)

type tokens struct {
	split  rune
	escape rune
}

type token struct {
	t int
	v []rune
}

func newtoken(t int) *token {
	return &token{t: t, v: make([]rune, 0, 10)}
}

func (r *token) append(v rune) {
	r.v = append(r.v, v)
}

type Parser struct {
	tokenStream []*token
	pos         int
}

func NewParser(tokenStream []*token) *Parser {
	return &Parser{tokenStream: tokenStream}
}

func (r *Parser) advance() {
	r.pos++
}

func (r *Parser) getToken(pos int) *token {
	if pos < 0 || pos >= len(r.tokenStream) {
		return nil
	}
	return r.tokenStream[pos]
}

func (r *Parser) prev() *token {
	return r.getToken(r.pos - 1)
}

func (r *Parser) cur() *token {
	return r.getToken(r.pos)
}

func (r *Parser) next() *token {
	return r.getToken(r.pos + 1)
}

func (r *Parser) Parse() []string {
	s := make([]string, 0, 10)
	return r.parse(s)
}

func (r *Parser) parse(s []string) []string {
	tk := r.cur()
	if tk == nil {
		return s
	}
	if tk.t == token_text || tk.t == token_escape {
		s = append(s, r.parseString(new(strings.Builder)))
	} else if tk.t == token_split {
		s = r.parseSplitter(s)
	} else {
		panic(fmt.Errorf("Unknown token type: %d", tk.t))
	}
	return r.parse(s)
}

func (r *Parser) parseString(b *strings.Builder) string {
	tk := r.cur()
	if tk == nil {
		return b.String()
	}
	if tk.t == token_text || tk.t == token_escape {
		for _, c := range tk.v {
			b.WriteRune(c)
		}
	} else {
		return b.String()
	}
	r.advance()
	return r.parseString(b)
}

func (r *Parser) parseSplitter(s []string) []string {
	tk := r.cur()
	if tk.t != token_split {
		panic(fmt.Errorf("Illegal call to parseSplitter at a token of type %d", tk.t))
	}
	if r.prev() == nil || r.prev().t == token_split {
		s = append(s, "")
	}
	if r.next() == nil {
		s = append(s, "")
	}
	r.advance()
	return s
}

type Tokenizer struct {
	input       []rune
	tk          *tokens
	pos         int
	tokenStream []*token
}

func NewTokenizer(split rune, escape rune) *Tokenizer {
	return &Tokenizer{tk: &tokens{split: split, escape: escape}}
}

func (r *Tokenizer) Tokenize(input string) ([]*token, error) {
	r.input = []rune(input)
	r.pos = 0
	r.tokenStream = make([]*token, 0, 20)
	if err := r.decide(); err != nil {
		return nil, err
	}
	return r.tokenStream, nil
}

func (r *Tokenizer) decide() error {
	if r.pos >= len(r.input) {
		return nil
	}
	switch r.input[r.pos] {
	case r.tk.split:
		return r.split()
	case r.tk.escape:
		return r.escape()
	default:
		return r.text()
	}
}

func (r *Tokenizer) escape() error {
	next := r.pos + 1
	if next >= len(r.input) {
		return fmt.Errorf("Index %d: Unexpected EOF after escape character", next)
	}
	tok := newtoken(token_escape)
	tok.append(r.input[next])
	r.tokenStream = append(r.tokenStream, tok)
	r.pos += 2
	return r.decide()
}

func (r *Tokenizer) split() error {
	tok := newtoken(token_split)
	tok.append(r.input[r.pos])
	r.tokenStream = append(r.tokenStream, tok)
	r.pos++
	return r.decide()
}

func (r *Tokenizer) text() error {
	tok := newtoken(token_text)
	pos := r.pos
	for ; pos < len(r.input); pos++ {
		switch r.input[pos] {
		case r.tk.split:
			r.tokenStream = append(r.tokenStream, tok)
			r.pos = pos
			return r.split()
		case r.tk.escape:
			r.tokenStream = append(r.tokenStream, tok)
			r.pos = pos
			return r.escape()
		default:
			tok.append(r.input[pos])
		}
	}
	r.tokenStream = append(r.tokenStream, tok)
	return nil
}
