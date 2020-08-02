/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package attrs

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	split := rune('/')
	escape := rune('\\')
	tests := make(map[string][]int)
	tests[""] = nil
	tests["test/split/end/"] = []int{token_text, token_split, token_text, token_split, token_text, token_split}
	tests["/test/split/end"] = []int{token_split, token_text, token_split, token_text, token_split, token_text}
	tests["\\\\test\\/123/456"] = []int{token_escape, token_text, token_escape, token_text, token_split, token_text}

	tk := NewTokenizer(split, escape)
	for k, v := range tests {
		ts, err := tk.Tokenize(k)
		if err != nil {
			t.Error(err)
		}
		for i, vv := range v {
			if vv != ts[i].t {
				t.Errorf("Test input \"%s\", index %d: Expected: %d, got %d", k, i, vv, ts[i].t)
			}
		}
	}
}

func TestParser(t *testing.T) {
	split := rune('/')
	escape := rune('\\')
	tests := make(map[string][]string)
	tests["test/split/end/"] = []string{"test", "split", "end", ""}
	tests["/test/split/end"] = []string{"", "test", "split", "end"}
	tests["/test/split/end/"] = []string{"", "test", "split", "end", ""}
	tests["///s///"] = []string{"", "", "", "s", "", "", ""}
	tests["///////"] = []string{"", "", "", "", "", "", "", ""}
	tests["\\\\/test\\/test/2"] = []string{"\\", "test/test", "2"}
	tests["\\/\\/\\/"] = []string{"///"}
	tk := NewTokenizer(split, escape)
	for k, exp := range tests {
		ts, err := tk.Tokenize(k)
		if err != nil {
			t.Errorf("%s: %v", t.Name(), err)
		}
		parser := NewParser(ts)
		res := parser.Parse()
		for i, expv := range exp {
			if expv != res[i] {
				t.Errorf("%s: Expected \"%v\", got \"%v\"", t.Name(), exp, res)
				break
			}
		}
	}
}
