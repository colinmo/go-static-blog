/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var formatStringDMonthYear = "2 January 2006"

/*
Perform a full regenerate from source.
1. Build the new site in a new folder under TempDir.
2. Replace the existing symlink with this location
*/
func updateFullRegenerate() (
	allPosts RSS,
	tags map[string][]FrontMatter,
	postsById map[string]Item,
	filesToDelete map[string]struct{},
	changes GitDiffs,
	err error) {

	PrintIfNotSilent("Full\n")
	postsById = map[string]Item{}
	allPosts = RSS{}
	tags = map[string][]FrontMatter{}
	filesToDelete = map[string]struct{}{}

	// Make new target directory
	PrintIfNotSilent("Temp Dir\n")
	SwapDir2 := ConfigData.BaseDir
	dirName := time.Now().Format("20060102150405")
	ConfigData.BaseDir = filepath.Join(ConfigData.TempDir, dirName)
	for _, d := range []string{"tag", "media", "posts"} {
		dirPath := filepath.Join(ConfigData.BaseDir, d)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("Make %s dir error %v\n", d, err)
		}
		info, err := os.Stat(dirPath)
		if os.IsNotExist(err) {
			log.Fatal("File does not exist.")
		}
		if !info.IsDir() {
			log.Fatalf("%s is not a directory", dirPath)
		}
	}

	// Run the generate into the target directory
	GitPull()
	changes, err = PopulateAllGitFiles(ConfigData.RepositoryDir)
	if err != nil {
		os.RemoveAll(ConfigData.BaseDir)
		ConfigData.BaseDir = SwapDir2
		err = fmt.Errorf("failed to get files in the directory %s [%s]", ConfigData.RepositoryDir, err)
		return
	}
	tags, filesToDelete, postsById = getAllChangedTagsAndDeletedFiles(changes, postsById)
	tags, postsById, _ = processFileUpdates(changes, tags, postsById)

	// Swap the directory symlink
	PrintIfNotSilent("\nSwap across\n")
	replaceDirectory(ConfigData.BaseDir, SwapDir2)
	// Remove old dir
	clearOtherPaths(ConfigData.TempDir, dirName)
	ConfigData.BaseDir = SwapDir2
	return
}

func clearOtherPaths(inDir, notThisOne string) {
	items, _ := os.ReadDir(inDir)
	for _, item := range items {
		if item.Name() < notThisOne {
			os.RemoveAll(filepath.Join(inDir, item.Name()))
		}
	}
}

func replaceDirectory(tempDir, blogDir string) {
	var err error
	blogDir = filepath.Clean(blogDir)
	err = os.Remove(blogDir)
	if err != nil && (len(err.Error()) <= 25 || err.Error()[len(err.Error())-25:] != `no such file or directory`) {
		log.Fatalf("Could not remove dir %v for linking to %v\n%v", blogDir, tempDir, err.Error())
	}
	err = os.Symlink(tempDir, blogDir)
	if err != nil {
		log.Fatalf("one: %s\n", err)
	}
}

func updateChangedRegenerate() (
	allPosts RSS,
	tags map[string][]FrontMatter,
	postsById map[string]Item,
	filesToDelete map[string]struct{},
	changes GitDiffs,
	err error) {

	PrintIfNotSilent("Changed\n")
	postsById = map[string]Item{}
	tags = map[string][]FrontMatter{}
	filesToDelete = map[string]struct{}{}

	// Get all posts from the all published posts RSS file
	allPosts, err = ReadRSS(filepath.Join(ConfigData.BaseDir, "all-rss.xml"))
	if err != nil {
		err = fmt.Errorf("failed to read the RSS file %v", err)
		return
	}
	for _, i := range allPosts.Channel.Items {
		postsById[i.GUID] = i
	}
	changes = GitRunDiff()
	// Get the tags to update and files to delete
	tags, filesToDelete, postsById = getAllChangedTagsAndDeletedFiles(changes, postsById)
	// Update the files
	GitPull()
	// Get the changed tags, building the new pages as we go
	tags, postsById, err = processFileUpdates(changes, tags, postsById)
	return
}

func deleteFiles(filesToDelete map[string]struct{}) {
	for filename := range filesToDelete {
		PrintIfNotSilent("Deleting " + filepath.Join(ConfigData.BaseDir, filename))
		os.Remove(filepath.Join(ConfigData.BaseDir, filename))
	}
}

func regenerateIndexAndRSS(allPosts RSS, postsById map[string]Item) map[string][]FrontMatter {
	allTagMap := map[string][]FrontMatter{}
	allPosts.Channel.Items = []Item{}
	allItems := []FrontMatter{}
	for _, i := range postsById {
		newPost := ItemToPost(i)
		allPosts.Channel.Items = append(allPosts.Channel.Items, i)
		allItems = append(allItems, newPost)
		for _, j := range newPost.Tags {
			if _, ok := allTagMap[j]; !ok {
				allTagMap[j] = []FrontMatter{}
			}
			allTagMap[j] = append(allTagMap[j], newPost)
		}
	}
	WriteRSS(allPosts, "/all-rss.xml", -1)
	WriteRSS(allPosts, "/rss.xml", 10)
	WriteListHTML(allItems, "index", "Journal")
	for _, top := range allItems {
		if top.Type != "indieweb" && top.Status != "draft" {
			err := WriteLatestPost(top)
			if err != nil {
				fmt.Printf("\nFailed to update homepage with latest %s\n", err)
			}
			break
		}
	}
	return allTagMap
}

func outputStats(changes GitDiffs) {
	if !Silent {
		// Output stats
		if Totals {
			fmt.Printf("\nTotals: A: %d, M: %d, D: %d\n", len(changes.Added), len(changes.Modified)+len(changes.RenameEdit)+len(changes.Unmerged), len(changes.Deleted))
		}

	}
}

func deleteAndRegenerate(
	allPosts RSS,
	tags map[string][]FrontMatter,
	postsById map[string]Item,
	filesToDelete map[string]struct{},
	changes GitDiffs,
) {
	// Delete any linked deleted HTML or Media pages
	deleteFiles(filesToDelete)
	// Regenerate the index pages and RSS feeds
	createPageAndRSSForTags(tags)
	// Regenerate the all published posts RSS file
	allTagMap := regenerateIndexAndRSS(allPosts, postsById)
	// Create tag-page for Code and Steampunk embedding
	for _, tag := range ConfigData.TagSnippets {
		PrintIfNotSilent(fmt.Sprintf("Regenerating snippet for %s (%d) - ", tag, len(allTagMap[tag])))
		content, err := createTagPageSnippetForTag(tag, allTagMap[tag])
		if err == nil {
			PrintIfNotSilent("ok\n")
			os.WriteFile(filepath.Join(ConfigData.BaseDir, "tag-snippet-"+tag+".html"), content, 0666)
		} else {
			PrintIfNotSilent(fmt.Sprintf("Failed %v", err))
			fmt.Printf("failed %v\n", err)
		}
	}
	outputStats(changes)
}

func createTagPageSnippetForTag(tag string, tagsForString []FrontMatter) ([]byte, error) {
	var templateTags map[string]interface{}
	var relatedTags map[string][]struct {
		Link  string
		Title string
	}
	var err error

	relatedTags = map[string][]struct {
		Link  string
		Title string
	}{}
	for _, e := range tagsForString {
		for _, f := range e.Tags {
			if f != tag {
				if _, ok := relatedTags[f]; !ok {
					relatedTags[f] = []struct {
						Link  string
						Title string
					}{}
				}
				relatedTags[f] = append(relatedTags[f], struct {
					Link  string
					Title string
				}{Link: e.Link, Title: e.Title})
			}
		}
	}
	buf := bytes.NewBufferString("")

	templateTags = map[string]interface{}{}
	templateTags["related_tags"] = relatedTags
	if err := templ.ExecuteTemplate(
		buf,
		"tag-related-tags",
		templateTags,
	); err != nil {
		fmt.Printf("Couldn't write the file3\n")
		log.Fatal(err)
	}

	return buf.Bytes(), err
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the blog",
	Long:  `Runs the markdown to html conversion process over the site`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the list of changed files
		var changes GitDiffs
		var allPosts RSS
		var err error
		var tags map[string][]FrontMatter
		var filesToDelete map[string]struct{}
		var postsById map[string]Item

		SetupTemplate()

		if FullRegenerate {
			allPosts, tags, postsById, filesToDelete, changes, err = updateFullRegenerate()
		} else {
			allPosts, tags, postsById, filesToDelete, changes, err = updateChangedRegenerate()
		}

		if err != nil {
			log.Fatalf("Something happened updating files\n%v\n", err)
		} else {
			deleteAndRegenerate(allPosts, tags, postsById, filesToDelete, changes)
		}
	},
}

var FullRegenerate bool
var Silent bool
var Totals bool

func getAllChangedTagsAndDeletedFiles(changes GitDiffs, postsById map[string]Item) (map[string][]FrontMatter, map[string]struct{}, map[string]Item) {
	var tags map[string][]FrontMatter
	var filesToDelete map[string]struct{}
	var linkString string
	// Get the old tags from the changed files
	for _, filename := range changes.CopyEdit {
		tags, _, _ = getTagsFromPost(filename, tags)
	}
	for _, filename := range changes.Deleted {
		tags, _, _ = getTagsFromPost(filepath.Join(ConfigData.BaseDir, filename), tags)
		// Get the linked HTML page for deleted files
		filesToDelete, linkString = getTargetFilenameFromPost(filename, filesToDelete)
		// Delete it from the Tag list as found in the RSS file
		if linkString != "" {
			delete(postsById, linkString)
		}
	}
	for _, filename := range changes.Modified {
		tags, _, _ = getTagsFromPost(filename, tags)
	}
	for _, filename := range changes.RenameEdit {
		tags, _, _ = getTagsFromPost(filename, tags)
	}
	for _, filename := range changes.Unmerged {
		tags, _, _ = getTagsFromPost(filename, tags)
	}
	return tags, filesToDelete, postsById
}

func getTagsFromPost(postName string, tags map[string][]FrontMatter) (map[string][]FrontMatter, FrontMatter, string) {
	var html string
	var frontmatter FrontMatter
	var err error
	if tags == nil {
		tags = make(map[string][]FrontMatter)
	}
	if postName[len(postName)-3:] == ".md" {
		html, frontmatter, err = parseFile(filepath.Join(ConfigData.RepositoryDir, postName))
		if err == nil && frontmatter.Status != "draft" {
			for _, tag := range frontmatter.Tags {
				tag = strings.ToLower(tag)
				tags[tag] = append(tags[tag], frontmatter)
			}
		}
	}
	return tags, frontmatter, html
}

func getTargetFilenameFromPost(postName string, files map[string]struct{}) (map[string]struct{}, string) {
	link := ""
	if files == nil {
		files = make(map[string]struct{})
	}
	if postName[len(postName)-3:] == ".md" {
		_, frontmatter, err := parseFile(filepath.Join(ConfigData.RepositoryDir, postName))
		if err == nil {
			files[frontmatter.RelativeLink] = struct{}{}
			link = frontmatter.Link
		} else {
			log.Fatalf("Couldn't get filename %v\n", err)
		}
	} else {
		files[postName] = struct{}{}
	}
	return files, link
}

func indieWeb(link, label string) string {
	if len(link) > 0 {
		return fmt.Sprintf("\n\n%s %s", label, link)
	}
	return ""
}

func postWantsCrosspost(frontmatter *FrontMatter, filename string) {
	if postWantsMastodonCrosspost(*frontmatter) {
		toSyndicate := frontmatter.Synopsis
		if frontmatter.Type == "indieweb" {
			toSyndicate = toSyndicate +
				fmt.Sprintf("%s%s%s%s%s",
					indieWeb(frontmatter.InReplyTo, "In reply to"),
					indieWeb(frontmatter.RepostOf, "Repost of"),
					indieWeb(frontmatter.LikeOf, "Like of"),
					indieWeb(frontmatter.FavoriteOf, "Favourite of"),
					indieWeb(frontmatter.BookmarkOf, "Bookmark of"),
				)
		} else {
			toSyndicate = toSyndicate + "\n\n" + frontmatter.Link
		}
		if len(frontmatter.Tags) > 0 {
			toSyndicate = toSyndicate + "\n#" + strings.Join(frontmatter.Tags, " #")
		}
		mastodonLink, err := postToMastodon(toSyndicate)
		if err == nil {
			mastodonLink, _ = url.JoinPath(`https://mstdn.social/@vonExplaino/`, mastodonLink)
			setMastodonLink(filename, mastodonLink)
			GitAdd(filename)
			GitCommit(fmt.Sprintf("XPOST - %s", mastodonLink))
			GitPush()
			frontmatter.SyndicationLinks.Mastodon = mastodonLink
		} else {
			PrintIfNotSilent("X")
		}
	}
	if postWantsBlueskyCrosspost(*frontmatter) {
		toSyndicate, facets := makeBlueskyPost(frontmatter)
		blueskyLink, err := postToBluesky(toSyndicate, facets, frontmatter.CreatedDate)
		if err == nil {
			setBlueskyLink(filename, blueskyLink)
			GitAdd(filename)
			GitCommit(fmt.Sprintf("BPOST - %s", blueskyLink))
			GitPush()
			frontmatter.SyndicationLinks.Bluesky = blueskyLink
		} else {
			PrintIfNotSilent(err.Error())
			PrintIfNotSilent("Y")
		}
	}
}
func processMDFile(tags *map[string][]FrontMatter, postsById *map[string]Item, filename string) error {
	// // If .md Process into HTML
	var err error
	t2, frontmatter, html := getTagsFromPost(filename, *tags)
	if frontmatter.Status == "draft" {
		PrintIfNotSilent("D")
		delete(*postsById, frontmatter.Link)
		return nil
	}
	*tags = t2
	targetFile := filepath.Join(ConfigData.BaseDir, baseDirectoryForPosts, frontmatter.RelativeLink)
	targetDir, _ := filepath.Split(targetFile)
	if _, err = os.Stat(targetFile); os.IsNotExist(err) {
		os.MkdirAll(targetDir, 0755)
	}
	err = os.WriteFile(targetFile, []byte(html), 0755)
	if frontmatter.Type == "article" ||
		frontmatter.Type == "review" ||
		(frontmatter.Type == "indieweb" &&
			(len(frontmatter.BookmarkOf) > 0 ||
				len(frontmatter.LikeOf) > 0)) {
		(*postsById)[frontmatter.Link] = PostToItem(frontmatter)
	}
	postWantsCrosspost(&frontmatter, filename)
	PrintIfNotSilent("P")
	return err
}

func postWantsMastodonCrosspost(fm FrontMatter) bool {
	return fm.SyndicationLinks.Mastodon == "XPOST"
}

func setMastodonLink(filename string, mastodonLink string) {
	filename = filepath.Join(ConfigData.RepositoryDir, filename)
	mep, err := os.ReadFile(filename)
	if err == nil {
		replc := regexp.MustCompile(`Mastodon:[ '"]*XPOST[ '"]*`)
		mep := replc.ReplaceAll(mep, []byte(fmt.Sprintf(`Mastodon: "%s"`, mastodonLink)))
		os.WriteFile(filename, mep, 0777)
	}
}

func postToMastodon(message string) (string, error) {
	type mastodonMostResponse struct {
		ID string `json:"id"`
	}
	data := url.Values{
		"status":     {message},
		"visibility": {"public"}, // testing
	}
	request, _ := http.NewRequest(
		"POST",
		ConfigData.Syndication.Mastodon.URL+"v1/statuses",
		bytes.NewBuffer([]byte(data.Encode())),
	)
	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+ConfigData.Syndication.Mastodon.Token)
	resp, err := Client.Do(request)
	if err != nil {
		return "", err
	}
	var res mastodonMostResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed in posting to mastodon [%d]", resp.StatusCode)
	}
	if res.ID != "" {
		return res.ID, nil
	} else {
		return "", fmt.Errorf("failed in post to mastodon attempt %s|%d", res, resp.StatusCode)
	}
}

func processMediaFile(filename string) error {
	targetFile := filepath.Join(ConfigData.BaseDir, filename)
	err := FileCopy(filepath.Join(ConfigData.RepositoryDir, filename), targetFile)
	PrintIfNotSilent("M")
	return err
}

func processUnknownFile(filename string) error {
	fullPath := filepath.Join(ConfigData.RepositoryDir, filename)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		fmt.Printf("Cannot do something with nothing %s\n", fullPath)
	} else if info.IsDir() {
		return nil
	} else {
		extension := filepath.Ext(filename)
		if !(extension == ".m4v" || extension == ".xcf" || filename[len(filename)-6:] == "README" || extension == ".html" || extension == ".txt" || extension == ".json") {
			fileType, err := GetFileType(fullPath)
			fmt.Printf("Could not copy %s|%s|%v\n", filename, fileType, err)
			return fmt.Errorf("unknown file type not copied %s|%s|%v", filename, fileType, err)
		}
	}
	return err
}

func processFileUpdates(changes GitDiffs, tags map[string][]FrontMatter, postsById map[string]Item) (map[string][]FrontMatter, map[string]Item, error) {
	var errors []string
	var err error
	for _, group := range [][]string{
		changes.Added,
		changes.CopyEdit,
		changes.Modified,
		changes.RenameEdit,
		changes.Unmerged} {
		for _, filename := range group {
			filename = strings.ReplaceAll(filename, `\`, `/`)
			extension := filepath.Ext(filename)
			if extension == ".md" {
				err = processMDFile(&tags, &postsById, filename)
			} else if (filename[0:5] == "media" || filename[0:6] == "/media") && (IsMedia(filepath.Join(ConfigData.RepositoryDir, filename)) || extension == ".mov") {
				err = processMediaFile(filename)
			} else {
				err = processUnknownFile(filename)
			}
			if err != nil {
				errors = append(errors, err.Error())
			}
		}
	}
	if len(errors) > 0 {
		err = fmt.Errorf("errors during update: %s", strings.Join(errors, ""))
	}
	return tags, postsById, err
}

func createPageAndRSSForTags(tags map[string][]FrontMatter) {
	// NEW IDEA
	// Read the main RSS feed of all content first, that'll give us all entries AND all tags
	// Then we delete any tags+entries that been deleted
	// Add the rest
	// And regenerate the main page and all changed tags
	baseDir := ConfigData.BaseDir
	matchExp, _ := regexp.Compile(`^(.*)-[\d+].xml$`)
	files, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Printf("Failed to read existing RSS files from [%s]\n[%s]\n", baseDir, err)
		log.Fatal(err)
	}

	// Get the new list of pages for each tag
	tagsAndFiles := make(map[string][]string)
	for _, f := range files {
		matches := matchExp.FindStringSubmatch(f.Name())
		if len(matches) > 0 {
			tagsAndFiles[matches[1]] = append(tagsAndFiles[matches[1]], f.Name())
		}
	}
	// For each tag
	for tag, frontMatters := range tags {
		var rss RSS
		var items []FrontMatter
		// Process all RSS feed files for that tag
		for _, f := range tagsAndFiles[textToSlug(tag)] {
			rss, err = ReadRSS(baseDir + string(os.PathSeparator) + f)
			if err != nil {
				fmt.Printf("Failed to read existing RSS file")
				log.Fatal(err)
			}
		}
		if rss.Channel.Title == "" {
			tagLink, _ := url.JoinPath(ConfigData.BaseURL, "tag", textToSlug(tag)+".xml")
			// New!
			rss.Channel = Channel{
				Title:         "Professor von Explain Feed Tagged " + tag,
				Link:          tagLink,
				Description:   "A feed of posts containing the tag '" + tag + "'",
				Language:      "",
				Copyright:     "",
				LastBuildDate: time.Now().String(),
				Generator:     "Hand crafted nonsense written in Go",
				WebMaster:     "professor@vonexplaino.com (Colin Morris)",
				TimeToLive:    "3600",
				Items:         []Item{},
			}
		}
		// Add new ones
		for _, post := range frontMatters {
			rss.Channel.Items = append(rss.Channel.Items, PostToItem(post))
			items = append(items, post)
		}
		// Regenerate RSS feeds and HTML pages for each Tag and Index
		filename := "tag/" + textToSlug(tag)
		WriteRSS(rss, fmt.Sprintf("%s.xml", filename), 20)
		WriteListHTML(items, filename, "Tag: "+tag)
	}
}

func TemplatifyPage(
	feed *[]FrontMatter,
	fileStrings []string,
	fileInts []int,
	pageStrings []string,
	prevPageStart *time.Time,
	prevPageEnd *time.Time,
) error {
	filenamePrefix := fileStrings[0]
	title := fileStrings[1]
	page := fileInts[0]
	pageCount := fileInts[1]
	firstPageStart := pageStrings[0]
	firstPageEnd := pageStrings[1]
	lastPageStart := pageStrings[2]
	lastPageEnd := pageStrings[3]
	chunkSize := ConfigData.PerPage
	lastPage := false
	if len(*feed) <= chunkSize {
		chunkSize = len(*feed)
		lastPage = true
	}
	buf := bytes.NewBufferString("")
	posts := (*feed)[0:chunkSize]
	templateTags := toTemplateListVariables(posts, title, page)
	templateTags["created_date"] = time.Now()
	templateTags["updated_date"] = time.Now()

	templateTags["base_url"] = ConfigData.BaseURL
	templateTags["link_prefix"], _ = url.JoinPath(ConfigData.BaseURL, filenamePrefix+"-")
	templateTags["last_page"] = pageCount
	templateTags["next_page"] = page + 1
	templateTags["prev_page"] = page - 1
	templateTags["first_page_start"] = firstPageStart
	templateTags["first_page_end"] = firstPageEnd
	templateTags["last_page_start"] = lastPageStart
	templateTags["last_page_end"] = lastPageEnd
	if page > 1 {
		templateTags["prev_page_start"] = prevPageStart.Format(formatStringDMonthYear)
		templateTags["prev_page_end"] = prevPageEnd.Format(formatStringDMonthYear)
	}
	*prevPageStart = posts[0].CreatedDate
	*prevPageEnd = posts[chunkSize-1].CreatedDate
	*feed = (*feed)[chunkSize:]
	if !lastPage {
		feedLen := len(*feed)
		templateTags["next_page_start"] = (*feed)[0].CreatedDate.Format(formatStringDMonthYear)
		if feedLen < chunkSize {
			templateTags["next_page_end"] = (*feed)[feedLen-1].CreatedDate.Format(formatStringDMonthYear)
			chunkSize = feedLen
		} else {
			templateTags["next_page_end"] = (*feed)[chunkSize-1].CreatedDate.Format(formatStringDMonthYear)
		}
	}

	if err := templ.ExecuteTemplate(
		buf,
		"list",
		templateTags,
	); err != nil {
		log.Fatal(err)
	}
	return os.WriteFile(fmt.Sprintf("%s-%d.html", filepath.Join(ConfigData.BaseDir, filenamePrefix), page), buf.Bytes(), 0777)
}

func WriteListHTML(feed []FrontMatter, filenamePrefix string, title string) error {
	if len(feed) == 0 {
		return nil
	}
	page := 1
	pageCount := int(math.Ceil(float64(len(feed)) / float64(ConfigData.PerPage)))
	sort.SliceStable(feed, func(p, q int) bool {
		return feed[p].CreatedDate.After(feed[q].CreatedDate)
	})

	//Dates
	firstPageStart := feed[0].CreatedDate.Format(formatStringDMonthYear)
	firstPageEnd := ""
	lastPageStart := firstPageStart
	lastPageEnd := firstPageEnd
	if pageCount == 1 {
		firstPageEnd = feed[len(feed)-1].CreatedDate.Format(formatStringDMonthYear)
	} else if pageCount > 1 {
		firstPageEnd = feed[ConfigData.PerPage-1].CreatedDate.Format(formatStringDMonthYear)
		lastPageStart = feed[(pageCount-1)*ConfigData.PerPage].CreatedDate.Format(formatStringDMonthYear)
		lastPageEnd = feed[len(feed)-1].CreatedDate.Format(formatStringDMonthYear)
	}
	prevPageStart := time.Now()
	prevPageEnd := time.Now()
	for {
		if len(feed) == 0 {
			break
		}
		err := TemplatifyPage(
			&feed,
			[]string{
				filenamePrefix,
				title,
			},
			[]int{
				page,
				pageCount,
			},
			[]string{
				firstPageStart,
				firstPageEnd,
				lastPageStart,
				lastPageEnd,
			},
			&prevPageStart,
			&prevPageEnd,
		)
		if err != nil {
			return err
		}
		// Prep next iteration
		page = page + 1
	}
	return nil
}

func WriteLatestPost(entry FrontMatter) error {
	buf := bytes.NewBufferString("")
	if err := templ.ExecuteTemplate(
		buf,
		"latest-article",
		toTemplateVariables(&entry, ""),
	); err != nil {
		fmt.Printf("Couldn't write the file1\n")
		log.Fatal(err)
	}

	filename := filepath.Join(ConfigData.BaseDir, "posts/page/welcome.html")
	mep, err := os.ReadFile(filename)
	if err == nil {
		replc := regexp.MustCompile(`<!-- START LAST(\n|.)*END LAST -->`)
		mep := replc.ReplaceAll(mep, []byte(fmt.Sprintf(`<!-- START LAST -->%s<!-- END LAST -->`, buf)))
		os.WriteFile(filename, mep, 0777)
	}

	return err
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&FullRegenerate, "fullregenerate", "f", false, "Do a full regeneration of the site")
	updateCmd.Flags().BoolVarP(&Silent, "silent", "s", false, "Run silently")
	updateCmd.Flags().BoolVarP(&Totals, "totals", "t", false, "Show totals")
}

func ClearDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func PopulateAllGitFiles(dir string) (GitDiffs, error) {
	var foundDiffs GitDiffs
	dirlength := len(dir)
	err := filepath.Walk(filepath.Join(dir, "media"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			foundDiffs.Added = append(foundDiffs.Added, path[dirlength:])
			return nil
		})
	if err != nil {
		return foundDiffs, err
	}
	if len(foundDiffs.Added) > 0 {
		foundDiffs.Added = foundDiffs.Added[1:]
	}

	var foundDiffs2 GitDiffs
	dirlength = len(dir)
	err = filepath.Walk(filepath.Join(dir, "posts"),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			foundDiffs2.Added = append(foundDiffs2.Added, path[dirlength:])
			return nil
		})
	if err != nil {
		return foundDiffs, err
	}
	if len(foundDiffs2.Added) > 0 {
		foundDiffs2.Added = foundDiffs2.Added[1:]
	}
	foundDiffs.Added = append(foundDiffs.Added, foundDiffs2.Added...)
	return foundDiffs, err
}

func IsMedia(file string) bool {
	fileType, err := GetFileType(file)
	if err != nil {
		return false
	}

	fileTypeBasic := strings.Split(fileType, "/")[0]
	for _, n := range []string{
		"audio",
		"image",
		"video",
	} {
		if n == fileTypeBasic {
			return true
		}
	}
	return fileType == "application/zip" || fileType == "application/pdf" || (file[len(file)-4:] == ".svg" && (fileType == "text/xml; charset=utf-8" || fileType == "text/plain; charset=utf-8" || fileType == "text/html; charset=utf-8")) || fileType == "application/ogg" || file[len(file)-9:] == ".htaccess"
}

func GetFileType(file string) (string, error) {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return "", err
	} else if info.IsDir() {
		return "app/directory", nil
	}

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Get the content
	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return "CantRead", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func FileCopy(source, destination string) error {
	targetDir, _ := filepath.Split(destination)
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatalf("Failed making root dirs for %s, %v\n", targetDir, err)
		}
	}
	var data, err1 = os.ReadFile(source)
	if err1 != nil {
		return err1
	}
	return os.WriteFile(destination, data, 0777)
}
