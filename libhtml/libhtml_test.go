package libhtml

import (
	"golang.org/x/net/html"
	"testing"
)

func TestMatchAttrs(t *testing.T) {

	nodeattr := make([]html.Attribute, 0, 2)
	nodeattr = append(nodeattr, html.Attribute{"test", "id", "1337"})
	nodeattr = append(nodeattr, html.Attribute{"", "src", "https://example.net/image.jpg"})
	n := &html.Node{Data: "test", Attr: nodeattr}

	musthave := make([]html.Attribute, len(nodeattr))
	copy(musthave, nodeattr)

	if !MatchAttrs(n, musthave...) {
		t.Errorf("%v and %v were expected to be equal", n.Attr, musthave)
	}

	musthave = append(musthave, html.Attribute{"", "alt", "test"})

	if MatchAttrs(n, musthave...) {
		t.Errorf("%v and %v were not expected to be equal", n.Attr, musthave)
	}

	empty := make([]html.Attribute, 0)
	if !MatchAttrs(n, empty...) {
		t.Error("A match against an empty slice should always return true")
	}
}
