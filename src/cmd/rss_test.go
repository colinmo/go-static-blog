package cmd

import (
	"fmt"
	"os"
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
	if mek.Version != "2.0" {
		t.Fatalf(`Version didn't parse %s`, mek.Version)
	}
	if mek.Channel.Title != "Professor von Explaino" {
		t.Fatalf(`Channel Title didn't parse [%s]`, mek.Channel.Title)
	}
	if len(mek.Channel.Items[0].Tags) != 1 {
		t.Fatalf(`Failed to parse tag count %d`, len(mek.Channel.Items[0].Tags))
	}
}

func TestReadWriteRSS(t *testing.T) {
	ConfigData.Metadata.Title = `Professor von Explaino`
	ConfigData.BaseURL = `https://vonexplaino.com/blog/`
	ConfigData.Metadata.Description = `Steampunk, PHP coding, Brisbane`
	ConfigData.Metadata.Language = `en-au`
	ConfigData.Metadata.Webmaster = `professor@vonexplaino.com (Colin Morris)`
	ConfigData.Metadata.Ttl = 40
	ConfigData.BaseDir = `f:\Dropbox\swap\golang\vonblog\features\tests\rss\`

	mek, err := ReadRSS(ConfigData.BaseDir + `rss1.xml`)
	if err != nil {
		t.Fatalf(`Failed to parse %s`, err)
	}

	err = WriteRSS(mek, `rss1_out.xml`, 10)
	if err != nil {
		t.Fatalf(`Failed to save the file %s`, err)
	}
	if len(mek.Channel.Items) == 0 {
		t.Fatalf(`Could not load tags`)
	}

	f1, _ := os.ReadFile(ConfigData.BaseDir + `rss1.xml`)
	f2, _ := os.ReadFile(ConfigData.BaseDir + `rss1_out.xml`)

	rep := regexp.MustCompile(`\n\s*`)

	rep2 := regexp.MustCompile(`<lastBuildDate>.*</lastBuildDate>`)
	f1s := rep.ReplaceAllString(rep2.ReplaceAllString(string(f1), ""), "")
	f2s := rep.ReplaceAllString(rep2.ReplaceAllString(string(f2), ""), "")
	if f1s != f2s {
		for i := range f1s {
			if f1s[i:i+1] != f2s[i:i+1] {
				fmt.Printf("Index %d is different %s:%s\n", i, f1s[i:i+1], f2s[i:i+1])
				fmt.Printf("\n%s\n%s\n", f1s[(i-10):(i+10)], f2s[(i-10):(i+10)])

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
