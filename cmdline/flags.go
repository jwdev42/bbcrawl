/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package cmdline

import (
	"bytes"
	"fmt"
	"github.com/jwdev42/bbcrawl/cmdline/attrs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Boolean bool

func (b *Boolean) Set(s string) error {
	lower := strings.ToLower(s)
	switch lower {
	case "true":
		*b = true
	case "false":
		*b = false
	default:
		return fmt.Errorf("Invalid input for Boolean flag: %q", s)
	}
	return nil
}

func (b *Boolean) String() string {
	if *b {
		return "true"
	}
	return "false"
}

type StartPage int

func (i *StartPage) Set(s string) error {
	num, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if num < 1 {
		return fmt.Errorf("%d is an invalid start page.", num)
	}
	*i = StartPage(num)
	return nil
}

func (i *StartPage) String() string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(int(*i))
}

type EndPage struct {
	start *StartPage
	End   int
}

func NewEndPage(start *StartPage) *EndPage {
	return &EndPage{start: start, End: 0}
}

func (p *EndPage) Set(s string) error {
	num, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if num < 1 {
		return fmt.Errorf("%d is an invalid start page.", num)
	}
	if num < int(*p.start) {
		return fmt.Errorf("End (%d) is greater than start (%d).", num, int(*p.start))
	}
	p.End = num
	return nil
}

func (p *EndPage) String() string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(p.End)
}

//URLCollection is used to convert a comma seperated string of raw urls into a slice of pointers to URL types.
type URLCollection struct {
	URLs []*url.URL
}

func (v *URLCollection) Set(s string) error {
	input := strings.Split(s, ",")
	v.URLs = make([]*url.URL, len(input), len(input))
	for i, rawurl := range input {
		if url, err := url.Parse(rawurl); err != nil {
			return err
		} else {
			v.URLs[i] = url
		}
	}
	return nil
}

func (v *URLCollection) String() string {
	if len(v.URLs) == 0 {
		return ""
	}
	var con bytes.Buffer
	for i, url := range v.URLs {
		if _, err := con.WriteString(url.String()); err != nil {
			return ""
		}
		if i+1 < len(v.URLs) {
			if err := con.WriteByte(','); err != nil {
				return ""
			}
		}
	}
	return con.String()
}

type SingleURL struct {
	URL *url.URL
}

func (v *SingleURL) Set(s string) error {
	if url, err := url.Parse(s); err != nil {
		return err
	} else {
		v.URL = url
	}
	return nil
}

func (v *SingleURL) String() string {
	if v.URL == nil {
		return ""
	}
	return v.URL.String()
}

type IntRange struct {
	Range [2]int
}

func (v *IntRange) Set(s string) error {
	splitted := strings.Split(s, ",")
	if len(splitted) != 2 {
		return fmt.Errorf("IntRange needs 2 values")
	}
	for i, str := range splitted {
		var err error
		v.Range[i], err = strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return err
		}
	}
	if v.Range[1] < v.Range[0] {
		return fmt.Errorf("the second integer must be greater or equal than the first")
	}
	return nil
}

func (v *IntRange) String() string {
	return fmt.Sprintf("%d,%d", v.Range[0], v.Range[1])
}

type IntTuple struct {
	Numbers []int
}

func (v *IntTuple) Set(s string) error {
	splitted := strings.Split(s, ",")
	v.Numbers = make([]int, 0, len(splitted))
	for _, str := range splitted {
		n, err := strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return err
		}
		v.Numbers = append(v.Numbers, n)
	}
	return nil
}

func (v *IntTuple) String() string {
	if v.Numbers == nil {
		return ""
	}
	return fmt.Sprintf("%v", v.Numbers)
}

type FSDirectory struct {
	Path string
}

func (v *FSDirectory) Set(s string) error {
	p, err := filepath.Abs(s)
	if err != nil {
		return err
	}
	fp, err := os.Open(p)
	if err != nil {
		return err
	}
	defer fp.Close()
	fpi, err := fp.Stat()
	if err != nil {
		return err
	}
	if !fpi.IsDir() {
		return fmt.Errorf("File \"%s\" is not a directory!", p)
	}
	v.Path = p
	return nil
}

func (v *FSDirectory) String() string {
	if v == nil {
		return ""
	}
	return v.Path
}

type Attrs map[string][]string

const (
	attrs_token_pair_separator      = rune('/')
	attrs_token_escape_character    = rune('\\')
	attrs_token_key_value_separator = '='
	attrs_token_value_separator     = ','
)

func (v Attrs) Set(s string) error {
	tok := attrs.NewTokenizer(attrs_token_pair_separator, attrs_token_escape_character)
	ts, err := tok.Tokenize(s)
	if err != nil {
		return err
	}
	parser := attrs.NewParser(ts)
	pairs := parser.Parse()
	for _, pair := range pairs {
		kv := strings.SplitN(pair, string(attrs_token_key_value_separator), 2)
		if len(kv) != 2 {
			return fmt.Errorf("Substring %q: Less or more than one equal sign", pair)
		}
		key := kv[0]
		vals := strings.Split(kv[1], string(attrs_token_value_separator))
		if _, ok := v[key]; ok {
			return fmt.Errorf("Key %q used twice", key)
		}
		v[key] = vals
	}
	return nil
}

func (v Attrs) String() string {
	if v == nil {
		return ""
	}
	builder := new(strings.Builder)
	elements := len(v)
	element := 1
	for key, vals := range v {
		builder.WriteString(key)
		builder.WriteByte(attrs_token_key_value_separator)
		for i, val := range vals {
			builder.WriteString(val)
			if i+1 < len(v[key]) {
				builder.WriteByte(attrs_token_value_separator)
			}
		}
		if element < elements {
			builder.WriteRune(attrs_token_pair_separator)
		}
		element++
	}
	return builder.String()
}

type StringWhitelist struct {
	delim     string
	elems     []string
	whitelist []string
}

func NewStringWhitelist(delim string, whitelisted ...string) *StringWhitelist {
	return &StringWhitelist{delim: delim, whitelist: whitelisted, elems: make([]string, 0, 5)}
}

func (v *StringWhitelist) Result() []string {
	return v.elems
}

func (v *StringWhitelist) Set(s string) error {
	splitted := strings.Split(s, v.delim)
	for _, name := range splitted {
		found := false
		for _, a := range v.whitelist {
			if name == a {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("String \"%s\" not in whitelist. Valid values are %s", name, strings.Join(v.whitelist, ", "))
		}
	}
	v.elems = splitted
	return nil
}

func (v *StringWhitelist) String() string {
	if v == nil || v.elems == nil {
		return ""
	}
	return strings.Join(v.elems, v.delim)
}
