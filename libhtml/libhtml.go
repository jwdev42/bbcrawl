package libhtml

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

func AttrVal(node *html.Node, attribute string) string {
	for _, attr := range node.Attr {
		if attr.Key == attribute {
			return attr.Val
		}
	}
	return ""
}

func ElementByID(n *html.Node, id string) *html.Node {
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

func ElementsByAttrMatch(n *html.Node, key string, val *regexp.Regexp) []*html.Node {
	nc := &nodecollection{nodes: make([]*html.Node, 0, 25)}
	walkTree(n, matchAttrVal(nc, key, val), nil)
	return nc.nodes
}

func ElementsByTag(n *html.Node, tag ...atom.Atom) []*html.Node {
	nodes := make([]*html.Node, 0, 10)
	pre := func(n *html.Node) bool {
		for _, t := range tag {
			if n.DataAtom == t {
				nodes = append(nodes, n)
				break
			}
		}
		return true
	}
	walkTree(n, pre, nil)
	return nodes
}

func ElementsByTagAndAttrs(n *html.Node, id string, attrs []html.Attribute) []*html.Node {
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

func HasAttr(node *html.Node, attribute string) bool {
	for _, attr := range node.Attr {
		if attr.Key == attribute {
			return true
		}
	}
	return false
}

func MatchAttrs(node *html.Node, attr ...html.Attribute) bool {
	attrs_to_match := make(map[html.Attribute]bool)
	for _, a := range attr {
		attrs_to_match[a] = false
	}

	for _, a := range node.Attr {
		if _, ok := attrs_to_match[a]; ok {
			attrs_to_match[a] = true
		}
	}

	for _, v := range attrs_to_match {
		if !v {
			return false
		}
	}
	return true
}
