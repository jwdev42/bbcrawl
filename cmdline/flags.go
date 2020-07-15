package cmdline

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

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

type FSDirectory struct {
	Path string
}

func (v *FSDirectory) Set(s string) error {
	fp, err := os.Open(s)
	if err != nil {
		return err
	}
	defer fp.Close()
	fpi, err := fp.Stat()
	if err != nil {
		return err
	}
	if !fpi.IsDir() {
		return fmt.Errorf("File \"%s\" is not a directory!", s)
	}
	v.Path = s
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
	attrs_token_pair_separator      = '/'
	attrs_token_key_value_separator = '='
	attrs_token_value_separator     = ','
)

func (v Attrs) Set(s string) error {
	pairs := strings.Split(s, string(attrs_token_pair_separator))
	for _, pair := range pairs {
		kv := strings.SplitN(pair, string(attrs_token_key_value_separator), 2)
		if len(kv) != 2 {
			return fmt.Errorf("Substring %q: Too many equal signs", pair)
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
			builder.WriteByte(attrs_token_pair_separator)
		}
		element++
	}
	return builder.String()
}