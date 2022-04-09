package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func TestGetTagsFromPost(t *testing.T) {
	testroot := "f:/dropbox/swap/golang/vonblog/features/tests/update/tags/"
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

func TestGetTargetFilenameFromPost(t *testing.T) {
	testroot := "f:/dropbox/swap/golang/vonblog/features/tests/update/tags/"
	type testexpect struct {
		filename string
		expected string
	}
	for _, thing := range []testexpect{
		{filename: "testfile1.md", expected: "/article/2021/03/mysql-learnings.html"},
		{filename: "testfile2.md", expected: "/article/2021/03/mepmep.html"},
		{filename: "testfile3.md", expected: "/page/2019/01/asdf-324-.html"},
		{filename: "testfile4.md", expected: "/article/2021/03/mysql-learnings.html"},
	} {
		ConfigData.RepositoryDir = testroot
		ConfigData.BaseURL = "https://vonexplaino.com/blog/"
		ConfigData.TemplateDir = "f:/dropbox/swap/golang/vonblog/templates/"
		filename, _ := getTargetFilenameFromPost(thing.filename, make(map[string]struct{}))

		_, ok := filename[thing.expected]
		// if !reflect.DeepEqual(tags, thing.expected) {
		if !ok {
			t.Fatalf(
				"Got the wrong tags for %s. Was %v, should be %v.\n",
				thing.filename,
				filename,
				thing.expected)
		}
	}
}

func TestCreatePageAndRSSForTags(t *testing.T) {
	testroot := "f:/dropbox/swap/golang/vonblog/features/tests/update/rss/"
	ConfigData.RepositoryDir = testroot
	ConfigData.BaseURL = "https://vonexplaino.com/blog/"
	ConfigData.BaseDir = testroot + "rundir/"
	ConfigData.TemplateDir = testroot + "../../../../templates/"
	ConfigData.PerPage = 20
	type testexpect struct {
		scenariodir   string
		tags          map[string][]FrontMatter
		filesToDelete map[string]struct{}
		expected      string
	}
	for _, thing := range []testexpect{
		{
			scenariodir: "1",
			tags: map[string][]FrontMatter{
				"tagone": {justParseFrontMatter("Title: MySQL Learnings\nTags: [tagone]\nCreated: 2021-03-31T17:11:15+1000\nUpdated: 2021-03-31T17:11:15+1000\nType: article\nStatus: live\nSynopsis: Upskilling my MySQL because I need a 6Gig database in my life created in Golang\nFeatureImage: /blog/media/2021/03/mysql-logo.svg")},
			},
			filesToDelete: map[string]struct{}{},
			expected:      "",
		},
	} {
		// Init the scenario
		ClearDir(testroot + "rundir/")
		copy(testroot+"scenarios/"+thing.scenariodir, testroot+"rundir/")
		createPageAndRSSForTags(thing.tags, thing.filesToDelete)
		// Check the file and rss are OK
		// failed := true
		// if failed {
		// 	t.Fatalf(
		// 		"Got the wrong tags for %s. Was %v, should be %v.\n",
		// 		thing.tags,
		// 		thing.filesToDelete,
		// 		thing.expected)
		// }
	}
}

func TestPopulateAllGitFiles(t *testing.T) {
	a, b := PopulateAllGitFiles("f:/dropbox/swap/golang/vonblog/features/tests/gits/")
	if b != nil {
		t.Fatalf("Failed to run populate all git files %v\n", b)
	}
	if len(a.Added) != 4 {
		t.Fatalf("Didn't get the expected list of files %v\n", a)
	}
	if a.Added[0] != `posts\testfile1.md` {
		t.Fatalf("Didn't find the first file [%v]\n", a.Added[0])
	}
	if a.Added[3] != `posts\testfile4.md` {
		t.Fatalf("Didn't find the last file %v\n", a.Added[3])
	}
}

func TestEmbeddedMarkdownInHtml(t *testing.T) {
	testroot := "f:/dropbox/swap/golang/vonblog/features/tests/make_page/embed-md/"
	type testexpect struct {
		filename string
		expected string
		fail     string
	}
	for _, thing := range []testexpect{
		{filename: "embed.md", expected: "embed.html", fail: "embedfail.html"},
	} {
		ConfigData.RepositoryDir = testroot
		ConfigData.BaseURL = "https://vonexplaino.com/blog/"
		ConfigData.BaseDir = testroot + "rundir/"
		ConfigData.TemplateDir = "f:/dropbox/swap/golang/vonblog/templates/"
		html, _, err := parseFile(testroot + thing.filename)

		// if !reflect.DeepEqual(tags, thing.expected) {
		if err != nil {
			t.Fatalf(
				"Failed to parse %s|%v\n",
				thing.filename,
				err)
		}
		if len(html) == 0 {
			t.Fatalf(
				"Didn't even make a single HTML file\n",
			)
		}
		txt, err := ioutil.ReadFile(testroot + thing.expected)
		if err != nil {
			t.Fatalf("Failed to parse the expected file\n")
		}
		// Ensure compare is close
		re := regexp.MustCompile(`\n`)
		html = re.ReplaceAllString(html, " ")
		text2 := re.ReplaceAllString(string(txt), " ")
		re = regexp.MustCompile(`\s\s+`)
		html = re.ReplaceAllString(html, " ")
		text2 = re.ReplaceAllString(text2, " ")

		if text2 != html {
			ioutil.WriteFile(testroot+thing.fail, []byte(html), 0777)
			fmt.Printf("Look in %s\n", testroot+thing.fail)
			t.Fatalf(
				"Files didn't match",
			)
		}
	}
}

func TestWriteLatestPost(t *testing.T) {
	ConfigData.TemplateDir = `f:\Dropbox\swap\golang\vonblog\templates\`
	ConfigData.BaseDir = `f:\Dropbox\swap\golang\vonblog\features\tests\update\latest\`
	testTime, _ := time.Parse("2006-01-02 15:04:05", "2021-12-30 19:00:23")
	entry := FrontMatter{
		ID:          "noogienoogie",
		Link:        "/here-goes.html",
		Title:       "Well let's check",
		CreatedDate: testTime,
		UpdatedDate: testTime,
		Synopsis:    "Fourscore and twenty vodkas ago",
	}
	err := WriteLatestPost(entry)
	if err != nil {
		t.Fatalf("Fek %v\n", err)
	}
}

func TestUpdateFullRegenerate(t *testing.T) {
	ConfigData.BaseDir = `f:\Dropbox\swap\golang\vonblog\features\tests\update\fullregenbase\`
	ConfigData.RepositoryDir = `f:\Dropbox\swap\golang\vonblog\features\tests\update\fullregenrep\`
	// Empty Repo
	gitCommand = `f:\Dropbox\swap\golang\vonblog\features\tests\update\scripts\empty.bat`
	allPosts, tags, postsById, filesToDelete, changes, err := updateFullRegenerate()
	if len(allPosts.Channel.Items) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(tags) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(postsById) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(filesToDelete) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Modified) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.CopyEdit) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.RenameEdit) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Added) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Deleted) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Unmerged) != 0 {
		t.Fatalf("How did that happen?")
	}
	if err != nil {
		t.Fatalf("How did that happen?")
	}
	// Repo has stuff
	gitCommand = `f:\Dropbox\swap\golang\vonblog\features\tests\update\scripts\fill.bat`
	allPosts, tags, postsById, filesToDelete, changes, err = updateFullRegenerate()
	if err != nil {
		t.Fatalf("Regenerate failed %v\n", err)
	}
	if len(allPosts.Channel.Items) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(tags) != 2 {
		t.Fatalf("How did that happen?")
	}
	if len(postsById) != 1 {
		t.Fatalf("How did that happen?")
	}
	if len(filesToDelete) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Modified) != 0 {
		t.Fatalf("How did that happen? %v", changes)
	}
	if len(changes.CopyEdit) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.RenameEdit) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Added) != 2 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Deleted) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(changes.Unmerged) != 0 {
		t.Fatalf("How did that happen?")
	}
}

func justParseFrontMatter(front string) FrontMatter {
	x, _ := parseFrontMatter(front, "")
	return x
}

func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}

func anUpdateIsRequestedWithVerbose() error {
	return godog.ErrPending
}

func iRunCreatePageGallerymd() error {
	return godog.ErrPending
}

func iRunCreatePageTodaymd() error {
	return godog.ErrPending
}

func iRunCreatePageTomorrowmd() error {
	return godog.ErrPending
}

func iShouldReceiveCreateNewPage(arg1 string) error {
	return godog.ErrPending
}

func iShouldReceiveCreateNewPageUpdatePageDeletePage(arg1, arg2, arg3 string) error {
	return godog.ErrPending
}

func iShouldReceiveDeletePage(arg1 string) error {
	return godog.ErrPending
}

func iShouldReceiveNewImage(arg1 string) error {
	return godog.ErrPending
}

func iShouldReceiveNoChangesFound() error {
	return godog.ErrPending
}

func iShouldReceiveUpdatePage(arg1 string) error {
	return godog.ErrPending
}

func iShouldSeeAFileFundayhtmlWithContentsRename() error {
	return godog.ErrPending
}

func iShouldSeeAFileGalleryhtmlWithContentsGallery() error {
	return godog.ErrPending
}

func iShouldSeeAFileTodayhtmlWithContentsBasic() error {
	return godog.ErrPending
}

func theBlogAndRemoteAreInScenario(arg1 int) error {
	return godog.ErrPending
}

func thePageBasicExists() error {
	return godog.ErrPending
}

func thePageGalleryExists() error {
	return godog.ErrPending
}

func thePageRenameExists() error {
	return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the page (.+) exists$`, pageExists)
	ctx.Step(`^run create page (.+)$`, runCreatePage)
	ctx.Step(`^should see a file (.+) with contents (.+)$`, shouldSeeAFileWithContents)
	ctx.Step(`^an update is requested with verbose$`, anUpdateIsRequestedWithVerbose)
	ctx.Step(`^I run create page gallery\.md$`, iRunCreatePageGallerymd)
	ctx.Step(`^I run create page today\.md$`, iRunCreatePageTodaymd)
	ctx.Step(`^I run create page tomorrow\.md$`, iRunCreatePageTomorrowmd)
	ctx.Step(`^I should receive Create new page "([^"]*)"$`, iShouldReceiveCreateNewPage)
	ctx.Step(`^I should receive Create new page "([^"]*)"\nUpdate page "([^"]*)"\nDelete page "([^"]*)"$`, iShouldReceiveCreateNewPageUpdatePageDeletePage)
	ctx.Step(`^I should receive Delete page "([^"]*)"$`, iShouldReceiveDeletePage)
	ctx.Step(`^I should receive New image "([^"]*)"$`, iShouldReceiveNewImage)
	ctx.Step(`^I should receive No changes found$`, iShouldReceiveNoChangesFound)
	ctx.Step(`^I should receive Update page "([^"]*)"$`, iShouldReceiveUpdatePage)
	ctx.Step(`^I should see a file funday\.html with contents rename$`, iShouldSeeAFileFundayhtmlWithContentsRename)
	ctx.Step(`^I should see a file gallery\.html with contents gallery$`, iShouldSeeAFileGalleryhtmlWithContentsGallery)
	ctx.Step(`^I should see a file today\.html with contents basic$`, iShouldSeeAFileTodayhtmlWithContentsBasic)
	ctx.Step(`^the blog and remote are in scenario (\d+)$`, theBlogAndRemoteAreInScenario)
	ctx.Step(`^the page basic exists$`, thePageBasicExists)
	ctx.Step(`^the page gallery exists$`, thePageGalleryExists)
	ctx.Step(`^the page rename exists$`, thePageRenameExists)
}
