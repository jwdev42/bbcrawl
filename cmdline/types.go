package cmdline

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type URLCollection struct {
	/*used to convert a comma seperated string of raw urls into a slice of pointers to URL types*/
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
	return nil
}

func (v *FSDirectory) String() string {
	return v.Path
}
