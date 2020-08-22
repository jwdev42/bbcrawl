/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"fmt"
	"math/rand"
	"net/url"
	"path"
	"strings"
)

type avTag map[string]string

func (r avTag) addSrc(src string) error {
	u, err := url.Parse(src)
	if err != nil {
		return err
	}
	name := fileNameFromURL(u)
	if name == "" {
		return fmt.Errorf("Could not derive a filename from input path \"%s\"", u.Path)
	}
	for k, v := range r {
		if v == name {
			if k == src {
				return nil //identical entry already exists
			}
			//filename exists for another entry
			name = r.randomName(name)
			break
		}
	}
	r[src] = name
	return nil
}

//internal use only
func (r avTag) randomName(name string) string {
	ext := path.Ext(name)
	b := new(strings.Builder)
	for i := 0; i < 64; i++ {
		num := rune(rand.Int31n(25) + 0x61)
		b.WriteRune(num)
	}
	if ext != "" {
		b.WriteByte('.')
		b.WriteString(ext)
	}
	return b.String()
}
