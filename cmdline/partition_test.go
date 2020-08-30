/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package cmdline

import (
	"fmt"
	"strings"
	"testing"
)

func testPartitionErrors(t *testing.T) {
	lines := make([]string, 0, 10)
	lines = append(lines, "")
	lines = append(lines, "bbcrawl")
	lines = append(lines, "bbcrawl -pager testpager http://www.example.net")
	lines = append(lines, "bbcrawl -crawler testcrawler -pager testpager http://www.example.net")
	lines = append(lines, "bbcrawl -pager testpager -crawler testcrawler")
	for _, line := range lines {
		_, err := Partition(strings.Split(line, " "))
		if err == nil {
			t.Errorf("%s: Input should have triggered an error: \"%s\"", t.Name(), line)
		} else {
			t.Logf("%s: Logging expected error for input: (error: \"%s\", input: \"%s\")", t.Name(), err, line)
		}
	}
}

func testPartitionPositive(t *testing.T) {
	lines := make([]string, 0, 10)
	lines = append(lines, "bbcrawl -arg1 yes -arg2 no -pager testpager -arg3 hello -arg4 there -crawler testcrawler -depth deep -height high http://example.net")
	lines = append(lines, "bbcrawl -pager testpager 1 2 3 -crawler testcrawler 4 5 6 http://example.net")
	lines = append(lines, "bbcrawl -pager testpager -crawler testcrawler 4 5 6 http://example.net")
	lines = append(lines, "bbcrawl -pager testpager 1 2 3 -crawler testcrawler http://example.net")
	lines = append(lines, "bbcrawl -pager testpager -crawler testcrawler http://example.net")
	lines = append(lines, "bbcrawl -pager testpager -crawler testcrawler http://example.net http://example.net/2")
	for _, line := range lines {
		res, err := Partition(strings.Split(line, " "))
		if err != nil {
			t.Logf("%s: Parser error: %s", t.Name(), err)
			t.FailNow()
		}
		result := fmt.Sprintf("bbcrawl %s", res.String())
		if line != result {
			t.Errorf("Expected: \"%s\", result: \"%s\"", line, result)
		}
	}
}

func TestPartition(t *testing.T) {
	testPartitionPositive(t)
	testPartitionErrors(t)
}
