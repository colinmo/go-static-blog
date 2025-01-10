package cmd

import (
	"testing"
)

func TestMakeBlueskyPost(t *testing.T) {
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
			"Syndication failed expected:'%s', got: '%s'",
			expected,
			synd)
	}
	if len(facets) > 0 {
		t.Fatalf("Darn I got facets of %v", facets)
	}

	fm = FrontMatter{
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
	synd, facets = makeBlueskyPost(&fm)
	expected = "Damnit, America, Halloween was _LAST_ week\n\nhttps://vonexplaino.com/blog/posts/toot/2024/11/06/thoughts-of-horror.html\n#code"
	if synd != expected {
		t.Fatalf(
			"Syndication failed expected:'%s', got: '%s'",
			expected,
			synd)
	}
	if len(facets) > 0 {
		t.Fatalf("Darn I got facets of %v", facets)
	}
}
