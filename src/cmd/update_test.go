package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	testdataloader "github.com/peteole/testdata-loader"
)

func TestGetTagsFromPost(t *testing.T) {
	testroot := filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/tags/`)
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
		ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates/`)
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
	testroot := filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/tags/`)
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
		ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates/`)
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
	testroot := filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/rss/`)
	ConfigData.RepositoryDir = testroot
	ConfigData.BaseURL = "https://vonexplaino.com/blog/"
	ConfigData.BaseDir = filepath.Join(testroot, "rundir/")
	ConfigData.TemplateDir = filepath.Clean(filepath.Join(testroot, "../../../../templates/"))
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
		createPageAndRSSForTags(thing.tags)
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
	a, b := PopulateAllGitFiles(filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/gits/`))
	if b != nil {
		t.Fatalf("Failed to run populate all git files %v\n", b)
	}
	if len(a.Added) != 4 {
		t.Fatalf("Didn't get the expected list of files %v\n", a)
	}
	if a.Added[0] != filepath.Join(`posts`, `testfile1.md`) {
		t.Fatalf("Didn't find the first file [%v][%v]\n", a.Added[0], filepath.Join(`posts`, `testfile1.md`))
	}
	if a.Added[3] != filepath.Join(`posts`, `testfile4.md`) {
		t.Fatalf("Didn't find the last file %v\n", a.Added[3])
	}
}

func TestEmbeddedMarkdownInHtml(t *testing.T) {
	testroot := filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/make_page/embed-md/`)
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
		ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates`)
		html, _, err := parseFile(filepath.Join(testroot, thing.filename))

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
		txt, err := os.ReadFile(filepath.Join(testroot, thing.expected))
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
			os.WriteFile(filepath.Join(testroot, thing.fail), []byte(html), 0777)
			fmt.Printf("Look in %s\n", filepath.Join(testroot, thing.fail))
			fmt.Printf("Look in %s\n", filepath.Join(testroot, thing.expected))
			t.Fatalf(
				"Files didn't match %s\n",
				filepath.Join(testroot, thing.fail),
			)
		}
	}
}

func TestWriteLatestPost(t *testing.T) {
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates/`)
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/latest/`)
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

func TestUpdateFullRegenerateBad(t *testing.T) {
	// Bad config
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + "/../statictest/B")
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + "/../statictest/A")
	ConfigData.BaseURL = `https://vonexplaino.com/`
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + "/../templates/")
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + "/../templates/features/tests/update/scripts/empty.sh")
	_, _, _, _, _, err := updateFullRegenerate()
	if err == nil {
		t.Fatalf("Didn't detect bad directory")
	}
}
func TestUpdateFullRegenerateEmpty(t *testing.T) {
	//
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/fullregenbase/`)
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/fullregenrep/`)
	ConfigData.BaseURL = `https://vonexplaino.com/`
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + "/../templates/")
	ConfigData.TempDir = `/tmp/`
	Silent = false
	// Empty Repo
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/scripts/empty.sh`)
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
		t.Fatalf("How did that happen? %s", err)
	}

}
func TestUpdateFullRegenerateFull(t *testing.T) {
	// Repo has stuff
	//
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/fullregenbase/`)
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/fullregenrep/`)
	ConfigData.BaseURL = `https://vonexplaino.com/`
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + "/../templates/")
	ConfigData.TempDir = `/tmp/`
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/scripts/fill.bat`)
	allPosts, tags, postsById, filesToDelete, changes, err := updateFullRegenerate()
	if err != nil {
		t.Fatalf("Regenerate failed %v\n", err)
	}
	if len(allPosts.Channel.Items) != 0 {
		t.Fatalf("How did that happen?")
	}
	if len(tags) != 2 {
		t.Fatalf("How did that happen? %d", len(tags))
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

//func TestDeleteAndRegenerate(t *testing.T) {
//	ConfigData.BaseDir = `c:/users/relap/dropbox\swap\golang\vonblog\features\tests\update\deleted\`
//
//	f, _ := os.Create(ConfigData.BaseDir + "testFile.txt")
//	f.Close()
//	filesToDelete := make(map[string]struct{})
//	filesToDelete["testFile.txt"] = struct{}{}
//
//	deleteAndRegenerate(
//		RSS{},
//		make(map[string][]FrontMatter),
//		make(map[string]Item),
//		filesToDelete,
//		GitDiffs{},
//	)
//}

func TestUpdateChangedRegenerateNoChange(t *testing.T) {
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/changed-no/`)
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/scripts/changed-nochanges.bat`)

	allPosts, tags, postsById, filesToDelete, changes, err := updateChangedRegenerate()
	if err != nil {
		t.Fatalf("Something dun goofed %v\n", err)
	}
	if len(allPosts.Channel.Items) > 0 {
		t.Fatalf("All posts had something in it")
	}
	if len(tags) > 0 {
		t.Fatalf("Tags had something in it")
	}
	if len(postsById) > 0 {
		t.Fatalf("Posts by id had something in it")
	}
	if len(filesToDelete) > 0 {
		t.Fatalf("Files to delete had something in it")
	}
	if len(changes.Added) > 0 {
		t.Fatalf("Changes added had things")
	}
}

// Changed

func TestUpdateChangedRegenerateDeleted(t *testing.T) {
	// Deleted
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/deleted/`)
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/deleted/`)
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/scripts/changed-deleted.bat`)
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates/`)

	allPosts, tags, postsById, filesToDelete, changes, err := updateChangedRegenerate()
	if err != nil {
		t.Fatalf("Something dun goofed %v\n", err)
	}
	if len(allPosts.Channel.Items) != 2 {
		fmt.Printf("%v\n", allPosts)
		t.Fatalf("Failed to parse the RSS file %s [%d]", filepath.Join(ConfigData.BaseDir, "rss.xml"), len(allPosts.Channel.Items))
	}
	if len(tags) > 0 {
		t.Fatalf("Tags had something in it")
	}
	if len(postsById) != 2 {
		t.Fatalf("Posts by id had something in it [%d]", len(postsById))
	}
	if len(filesToDelete) != 1 {
		t.Fatalf("Where's the delete?")
	}
	if len(changes.Added) > 0 {
		t.Fatalf("Changes added had things")
	}
}

func TestUpdateChangedRegenerateAdded(t *testing.T) {
	// Added
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/changed/`)
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/changed-repo/`)
	ConfigData.TemplateDir = filepath.Clean(testdataloader.GetBasePath() + `/../templates/`)
	gitCommand = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/scripts/changed-added.bat`)

	allPosts, tags, postsById, filesToDelete, changes, err := updateChangedRegenerate()
	if err != nil {
		t.Fatalf("Something dun goofed %v\n", err)
	}
	if len(allPosts.Channel.Items) != 1 {
		fmt.Printf("%v\n", allPosts)
		t.Fatalf("Failed to parse the RSS file %s [%d]", filepath.Join(ConfigData.BaseDir, "rss.xml"), len(allPosts.Channel.Items))
	}
	if len(tags) != 2 {
		t.Fatalf("Tags had something in it %d", len(tags))
	}
	if len(postsById) != 2 {
		t.Fatalf("Posts by id had something in it [%d]", len(postsById))
	}
	if len(filesToDelete) != 0 {
		t.Fatalf("What's to delete?")
	}
	if len(changes.Added) != 1 {
		t.Fatalf("Changes miscounted had things: A%d", len(changes.Added))
	}
	if len(changes.Modified) != 1 {
		t.Fatalf("Changes miscounted had things: M%d", len(changes.Modified))
	}

	// Posts from RSS too
}
func TestMastodonPostCheck(t *testing.T) {
	// When parsing a post, check if the Mastodon syndication is set, but empty.
	if !postWantsMastodonCrosspost(FrontMatter{SyndicationLinks: SyndicationLinksS{Mastodon: "XPOST"}}) {
		t.Fatalf("A post that should want a Mastodon crosspost, does not")
	}

	if postWantsMastodonCrosspost(FrontMatter{SyndicationLinks: SyndicationLinksS{}}) {
		t.Fatalf("A post that shouldn't want a Mastodon crosspost, does want")
	}

	if postWantsMastodonCrosspost(FrontMatter{}) {
		t.Fatalf("A post that shouldn't want a Mastodon crosspost, does want")
	}

	if postWantsMastodonCrosspost(FrontMatter{SyndicationLinks: SyndicationLinksS{Mastodon: "https://xxx"}}) {
		t.Fatalf("A post that already has a Mastodon crosspost wants another!")
	}
}

/*
func TestMastodonCrosspost(t *testing.T) {
	makeTestConfig()
	mep, err := postToMastodon("test!")
	if err != nil {
		t.Fatalf("Failed %s", err)
	}
	if mep == "" {
		t.Fatalf("Failed, silently")
	}
	// Crosspost the content of the Post to Mastodon
	// Update the Post to have the Mastodon post ID as the Mastodon syndication value
}
*/

func TestUpdateLocalFileWithMastodonLink(t *testing.T) {
	ConfigData.RepositoryDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/update/mstdn/`)
	fullname := "steve.md"
	os.WriteFile(filepath.Join(ConfigData.RepositoryDir, fullname), []byte("---\nTitle: bob\nSyndication:\n  Mastodon: XPOST\n---\nWell"), 0777)
	setMastodonLink(fullname, "thisisatestlink")
	z, err := os.ReadFile(filepath.Join(ConfigData.RepositoryDir, fullname))
	if err != nil {
		t.Fatalf("test failed, couldn't read %v", err)
	}
	if !strings.Contains(string(z), "  Mastodon: \"thisisatestlink\"\n") {
		t.Fatalf("did not update %s correctly '%s'", fullname, z)
	}

}
func TestUpdatePostToHost(t *testing.T) {
	// Post it back to Bitbucket.
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
			var data, err1 = os.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return os.WriteFile(filepath.Join(destination, relPath), data, 0777)
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
