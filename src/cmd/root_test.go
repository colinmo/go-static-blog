package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestFullRebuild(t *testing.T) {
	testroot := `F:\Dropbox\swap\golang\vonblog\features\tests\update\tags\`
	type testexpect struct {
		filename string
		expected map[string][]FrontMatter
	}
	t2 := []FrontMatter{{Title: "MySQL Learnings"}}
	t3 := []FrontMatter{{Title: "asdf#$324#@$"}}
	for _, thing := range []testexpect{
		{filename: "testfile1.md", expected: map[string][]FrontMatter{}},
		{filename: "testfile2.md", expected: map[string][]FrontMatter{"tagone": t2}},
		{filename: "testfile3.md", expected: map[string][]FrontMatter{"tagtwo": t3, "tag3": t3}},
		{filename: "testfile4.md", expected: map[string][]FrontMatter{}},
	} {
		ConfigData.RepositoryDir = testroot
		ConfigData.BaseURL = "https://vonexplaino.com/blog/"
		ConfigData.TemplateDir = "f:/dropbox/swap/golang/vonblog/templates/"
		var tags map[string][]FrontMatter
		tags, _, _ = getTagsFromPost(thing.filename, tags)

		if len(tags) != len(thing.expected) {
			t.Fatalf(
				"Got the wrong count of tags for %s. Was %d, should be %d.\n",
				thing.filename,
				len(tags),
				len(thing.expected))
		}
		// Ensure we have tags we need
		for x, y := range thing.expected {
			val, ok := tags[x]
			if !ok {
				t.Fatalf("Welp, didn't find %s", x)
			}
			if len(val) != len(y) {
				fmt.Printf("Expected: %v\nActual: %v", y, val)
				t.Fatalf("Didn't match tag sizes %d,%d", len(y), len(val))
			}
			for x1, y1 := range y {
				if y1.Title != tags[x][x1].Title {
					t.Fatalf("Welp, title %s didn't match %s", y1.Title, tags[x][x1].Title)
				}
			}
		}
		// Ensure we don't have tags we don't need
		for x := range tags {
			_, ok := thing.expected[x]
			if !ok {
				t.Fatalf("Welp, shouldn't have found %s", x)
			}
		}
	}
}

func TestPrintIfNotSilent(t *testing.T) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Silent = false
	PrintIfNotSilent("hi")

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	if string(out) != "hi" {
		t.Errorf("Didn't print %s\n", string(out))
	}

	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	Silent = true
	PrintIfNotSilent("again")

	w.Close()
	out, _ = ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	if string(out) == "again" {
		t.Errorf("Was not silent")
	}
}
