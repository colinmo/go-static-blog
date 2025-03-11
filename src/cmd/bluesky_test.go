package cmd

import (
	"strings"
	"testing"
)

func TestMakeBlueskyPost1(t *testing.T) {
	fm := FrontMatter{
		Title:        `thoughts-of-horror`,
		Created:      `2024-11-06T20:22:32+1000`,
		Tags:         []string{"code"},
		Type:         `toot`,
		Status:       `live`,
		Slug:         `thoughts-of-horror.html`,
		Synopsis:     `Damnit, America, Halloween was _LAST_ week`,
		FeatureImage: ``,
		Link:         `https://vonexplaino.com/blog/posts/toot/2024/11/06/thoughts-of-horror.html`,
		Updated:      `2024-11-06T20:22:32+1000`,
		Author:       `Colin Morris`,
	}
	synd, facets := makeBlueskyPost(&fm)
	expected := "Damnit, America, Halloween was _LAST_ week\n#code"
	if synd != expected {
		t.Fatalf(
			"Syndication failed expected:\n%s\ngot\n%s\n",
			expected,
			synd)
	}
	if len(facets) > 0 {
		t.Fatalf("Darn I got facets of %v", facets)
	}
}
func TestMakeBlueskyPost2(t *testing.T) {
	fm := FrontMatter{
		Title:        `thoughts-of-horror`,
		Created:      `2024-11-06T20:22:32+1000`,
		Tags:         []string{"code"},
		Type:         `article`,
		Status:       `live`,
		Slug:         `thoughts-of-horror.html`,
		Synopsis:     `Damnit, America, Halloween was _LAST_ week`,
		FeatureImage: ``,
		Link:         `https://vonexplaino.com/blog/posts/toot/2024/11/06/thoughts-of-horror.html`,
		Updated:      `2024-11-06T20:22:32+1000`,
		Author:       `Colin Morris`,
	}
	synd, facets := makeBlueskyPost(&fm)
	expected := "Damnit, America, Halloween was _LAST_ week\r\n\r\nhttps://vonexplaino.com/blog/posts/toot/2024/11/06/thoughts-of-horror.html\n#code"
	if strings.Compare(synd, expected) != 0 {
		t.Fatalf(
			"Syndication 2 failed expected:\n[%s]\ngot\n[%s]\n%d",
			expected,
			synd,
			strings.Compare(synd, expected))
	}
	if len(facets) != 1 {
		t.Fatalf("Darn I got the wrong number of facets: %v", facets)
	}
}
