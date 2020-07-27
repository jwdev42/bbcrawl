/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package libcrawl

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"regexp"
)

type nodecollection struct {
	nodes []*html.Node
}

func walkTree(n *html.Node, pre, post func(*html.Node) bool) bool {
	if pre != nil && !pre(n) {
		return false
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !walkTree(c, pre, post) {
			return false
		}
	}
	if post != nil && !post(n) {
		return false
	}
	return true
}

//BEGIN: Functions to be used as pre or post with walktree

//matchAttrVal add Node n and child nodes to the nodecollection nc if the regex val matches
//the value of the attribute specified with key
func matchAttrVal(nc *nodecollection, key string, val *regexp.Regexp) func(n *html.Node) bool {
	return func(node *html.Node) bool {
		for _, a := range node.Attr {
			if a.Key == key && val.MatchString(a.Val) {
				nc.nodes = append(nc.nodes, node)
				return true
			}
		}
		return true
	}
}

//END: Functions to be used as pre or post with walktree*

func elementByID(n *html.Node, id string) *html.Node {
	var elem *html.Node
	byID := func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == id {
				elem = n
				return false
			}
		}
		return true
	}
	walkTree(n, byID, nil)
	return elem
}

func elementsByAttrMatch(n *html.Node, key string, val *regexp.Regexp) *nodecollection {
	nc := &nodecollection{nodes: make([]*html.Node, 0, 25)}
	walkTree(n, matchAttrVal(nc, key, val), nil)
	return nc
}

func elementsByTag(n *html.Node, tag atom.Atom) []*html.Node {
	nodes := make([]*html.Node, 0, 10)
	pre := func(n *html.Node) bool {
		if n.DataAtom == tag {
			nodes = append(nodes, n)
		}
		return true
	}
	walkTree(n, pre, nil)
	return nodes
}

func elementsByTagAndAttrs(n *html.Node, id string, attrs []html.Attribute) []*html.Node {
	nodes := make([]*html.Node, 0, 10)
	pre := func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == id {
			for _, a := range attrs {
				found := false
				for _, na := range n.Attr {
					if a == na {
						found = true
						break
					}
				}
				if !found {
					return true
				}
			}
			nodes = append(nodes, n)
		}
		return true
	}
	walkTree(n, pre, nil)
	return nodes
}
