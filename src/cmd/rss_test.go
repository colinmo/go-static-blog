package cmd

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"testing"
)

// TestReadRSS test parses some RSS Feeds
func TestReadRSS(t *testing.T) {
	mek, err := ReadRSS(`f:\Dropbox\swap\golang\vonblog\features\tests\rss\rss1.xml`)
	if err != nil {
		t.Fatalf(`Failed to parse %s`, err)
	}
	if mek.Version != "1.0" {
		t.Fatalf(`Version didn't parse %s`, mek.Version)
	}
	if mek.Channel.Title != "Professor von Explaino" {
		t.Fatalf(`Channel Title didn't parse [%s]`, mek.Channel.Title)
	}
}

func TestReadWriteRSS(t *testing.T) {
	ConfigData.Metadata.Title = `Professor von Explaino`
	ConfigData.BaseURL = `https://vonexplaino.com/blog/`
	ConfigData.Metadata.Description = `Steampunk, PHP coding, Brisbane`
	ConfigData.Metadata.Language = `en-au`
	ConfigData.Metadata.Webmaster = `professor@vonexplaino.com (Colin Morris)`
	ConfigData.Metadata.Ttl = 40

	mek, err := ReadRSS(`f:\Dropbox\swap\golang\vonblog\features\tests\rss\rss1.xml`)
	if err != nil {
		t.Fatalf(`Failed to parse %s`, err)
	}

	err = WriteRSS(mek, `f:/Dropbox/swap/golang/vonblog/features/tests/rss/rss1_out.xml`)
	if err != nil {
		t.Fatalf(`Failed to save the file %s`, err)
	}
	if len(mek.Channel.Items) == 0 {
		t.Fatalf(`Could not load tags`)
	}

	f1, _ := ioutil.ReadFile(`f:\Dropbox\swap\golang\vonblog\features\tests\rss\rss1.xml`)
	f2, _ := ioutil.ReadFile(`f:\Dropbox\swap\golang\vonblog\features\tests\rss\rss1_out.xml`)

	rep := regexp.MustCompile(`\n\s*`)

	rep2 := regexp.MustCompile(`<lastBuildDate>[a-zA-Z]{3}, \d{1,2} [a-zA-Z]{3} \d{4} \d{2}:\d{2}:\d{2} \+1000</lastBuildDate>`)
	f1s := rep.ReplaceAllString(rep2.ReplaceAllString(string(f1), ""), "")
	f2s := rep.ReplaceAllString(rep2.ReplaceAllString(string(f2), ""), "")
	fmt.Printf("Basd %s\n", f1s[600:700])

	if f1s != f2s {
		for i := range f1s {
			if f1s[i:i+1] != f2s[i:i+1] {
				fmt.Printf("Index %d is different %s:%s\n", i, f1s[i:i+1], f2s[i:i+1])
				t.Fatalf(`Nuts`)
			}
		}
		t.Fatalf(`Files didn't match`)
	}
}

func TestSortRSS(t *testing.T) {
	mek, err := ReadRSS(`f:\Dropbox\swap\golang\vonblog\features\tests\rss\rss1.xml`)
	if err != nil {
		t.Fatalf(`Failed to parse %s`, err)
	}

	sort.Slice(mek.Channel.Items, func(p, q int) bool {
		return mek.Channel.Items[p].PubDateAsDate.After(mek.Channel.Items[q].PubDateAsDate)
	})

	if mek.Channel.Items[0].Title != "Bookmark: Scaling the Practice of Architecture, Conversationally" {
		t.Fatalf(`Did not sort the right way %s`, mek.Channel.Items[0].Title)
	}
}
