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
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyler-sommer/stick"
	"github.com/tyler-sommer/stick/twig"
)

var formatStringDMonthYear = "2 January 2006"

func updateFullRegenerate() (RSS, map[string][]FrontMatter, map[string]Item, map[string]struct{}, GitDiffs, error) {
	postsById := map[string]Item{}
	PrintIfNotSilent("Full\n")
	allPosts := RSS{}
	ClearDir(ConfigData.BaseDir)
	os.MkdirAll(fmt.Sprintf("%s%s", ConfigData.BaseDir, "tag"), 0755)
	GitPull()
	changes, err := PopulateAllGitFiles(ConfigData.RepositoryDir)
	if err != nil {
		log.Fatalf("Failed to get files in directory %s\n", ConfigData.RepositoryDir)
	}
	tags, filesToDelete, postsById := getAllChangedTagsAndDeletedFiles(changes, postsById)
	tags, postsById, err = processFileUpdates(changes, tags, postsById)
	return allPosts, tags, postsById, filesToDelete, changes, err
}

func updateChangedRegenerate() (RSS, map[string][]FrontMatter, map[string]Item, map[string]struct{}, GitDiffs, error) {
	postsById := map[string]Item{}
	PrintIfNotSilent("Changed\n")

	// Get all posts from the all published posts RSS file
	allPosts, _ := ReadRSS(ConfigData.BaseDir + "/rss.xml")
	for _, i := range allPosts.Channel.Items {
		postsById[i.GUID] = i
	}
	changes := GitRunDiff()
	// Get the tags to update and files to delete
	tags, filesToDelete, postsById := getAllChangedTagsAndDeletedFiles(changes, postsById)
	// Update the files
	GitPull()
	// Get the changed tags, building the new pages as we go
	tags, postsById, err := processFileUpdates(changes, tags, postsById)
	return allPosts, tags, postsById, filesToDelete, changes, err
}

func deleteAndRegenerate(allPosts RSS, tags map[string][]FrontMatter, postsById map[string]Item, filesToDelete map[string]struct{}, changes GitDiffs) {
	// Delete any linked deleted HTML or Media pages
	for filename := range filesToDelete {
		os.Remove(ConfigData.BaseDir + filename)
	}
	// Regenerate the index pages and RSS feeds
	createPageAndRSSForTags(tags, filesToDelete)
	// Regenerate the all published posts RSS file
	allPosts.Channel.Items = []Item{}
	allItems := []FrontMatter{}
	for _, i := range postsById {
		allPosts.Channel.Items = append(allPosts.Channel.Items, i)
		allItems = append(allItems, ItemToPost(i))
	}
	WriteRSS(allPosts, "/rss.xml")
	WriteListHTML(allItems, "index", "Journal")
	for _, top := range allItems {
		if top.Type != "indieweb" && top.Status != "draft" {
			WriteLatestPost(top)
			break
		}
	}
	if Totals {
		fmt.Printf("\nTotals: A: %d, M: %d, D: %d\n", len(changes.Added), len(changes.Modified)+len(changes.RenameEdit)+len(changes.Unmerged), len(changes.Deleted))
	}
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

		if FullRegenerate {
			allPosts, tags, postsById, filesToDelete, changes, err = updateFullRegenerate()
		} else {
			allPosts, tags, postsById, filesToDelete, changes, err = updateChangedRegenerate()
		}
		if err != nil {
			log.Fatalf("Something happened updating files %v\n", err)
		}
		deleteAndRegenerate(allPosts, tags, postsById, filesToDelete, changes)
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
		tags, _, _ = getTagsFromPost(filename, tags)
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
		html, frontmatter, err = parseFile(ConfigData.RepositoryDir + postName)
		if err == nil {
			for _, tag := range frontmatter.Tags {
				tag = strings.ToLower(tag)
				tags[tag] = append(tags[tag], frontmatter)
			}
		} else {
			log.Fatalf("Couldn't get tags %v [%s]\n", err, ConfigData.RepositoryDir+postName)
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
		_, frontmatter, err := parseFile(ConfigData.RepositoryDir + postName)
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

func processMDFile(tags *map[string][]FrontMatter, postsById *map[string]Item, filename string) error {
	// // If .md Process into HTML
	var err error
	t2, frontmatter, html := getTagsFromPost(filename, *tags)
	*tags = t2
	targetFile := ConfigData.BaseDir + baseDirectoryForPosts + frontmatter.RelativeLink
	targetDir, _ := path.Split(targetFile)
	if _, err = os.Stat(targetFile); os.IsNotExist(err) {
		os.MkdirAll(targetDir, 0755)
	}
	err = os.WriteFile(targetFile, []byte(html), 0755)
	if frontmatter.Status != "draft" && (frontmatter.Type == "article" || frontmatter.Type == "review" || (frontmatter.Type == "indieweb" && (len(frontmatter.BookmarkOf) > 0 || len(frontmatter.LikeOf) > 0))) {
		(*postsById)[frontmatter.Link] = PostToItem(frontmatter)
		PrintIfNotSilent("P")
	} else {
		PrintIfNotSilent("D")
	}
	return err
}

func processMediaFile(filename string) error {
	targetFile := ConfigData.BaseDir + filename
	err := FileCopy(ConfigData.RepositoryDir+filename, targetFile)
	PrintIfNotSilent("M")
	return err
}

func processUnknownFile(filename string) error {
	info, err := os.Stat(ConfigData.RepositoryDir + filename)
	if os.IsNotExist(err) {
		fmt.Printf("Cannot do something with nothing %s\n", ConfigData.RepositoryDir+filename)
		log.Fatal("FAILED")
	} else if info.IsDir() {
		// nothing
	} else {
		split := strings.Split(filename, ".")
		var extension string
		if len(split) > 1 {
			extension = split[1]
		} else {
			extension = ""
		}
		if !(extension == "m4v" || extension == "xcf" || filename[len(filename)-6:] == "README" || extension == "html" || extension == "txt" || extension == "json") {
			fileType, err := GetFileType(ConfigData.RepositoryDir + filename)
			fmt.Printf("Could not copy %s|%s|%v\n", filename, fileType, err)
			log.Fatalf("FAILED")
		}
	}
	return err
}

func processFileUpdates(changes GitDiffs, tags map[string][]FrontMatter, postsById map[string]Item) (map[string][]FrontMatter, map[string]Item, error) {
	var err error
	for _, group := range [][]string{
		changes.Added,
		changes.CopyEdit,
		changes.Modified,
		changes.RenameEdit,
		changes.Unmerged} {
		for _, filename := range group {
			filename = strings.ReplaceAll(filename, `\`, `/`)
			if filename[len(filename)-3:] == ".md" {
				err = processMDFile(&tags, &postsById, filename)
			} else if (filename[0:5] == "media" || filename[0:6] == "/media") && (IsMedia(ConfigData.RepositoryDir+filename) || filename[len(filename)-4:] == ".mov") {
				processMediaFile(filename)
			} else {
				err = processUnknownFile(filename)
			}
		}
	}
	return tags, postsById, err
}

func createPageAndRSSForTags(tags map[string][]FrontMatter, filesToDelete map[string]struct{}) {
	// NEW IDEA
	// Read the main RSS feed of all content first, that'll give us all entries AND all tags
	// Then we delete any tags+entries that been deleted
	// Add the rest
	// And regenerate the main page and all changed tags
	baseDir := ConfigData.BaseDir
	matchExp, _ := regexp.Compile(`^(.*)-[\d+].xml$`)
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		fmt.Printf("Failed to read existing RSS files from [%s]\n", baseDir)
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
			// New!
			rss.Channel = Channel{
				Title:         "Professor von Explain Feed Tagged " + tag,
				Link:          ConfigData.BaseURL + "tag/" + textToSlug(tag) + ".xml",
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
		WriteRSS(rss, fmt.Sprintf("%s.xml", filename))
		WriteListHTML(items, filename, "Tag: "+tag)
	}
}

func TwigifyPage(
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
	posts := (*feed)[0:chunkSize]
	twigTags := toTwigListVariables(posts, title, page)
	tDir := ConfigData.TemplateDir
	env := twig.New(stick.NewFilesystemLoader(tDir))
	env.Filters["tag_link"] = filterTagLink

	twigTags["base_url"] = ConfigData.BaseURL
	twigTags["link_prefix"] = ConfigData.BaseURL + filenamePrefix + "-"
	twigTags["last_page"] = pageCount
	twigTags["next_page"] = page + 1
	twigTags["prev_page"] = page - 1
	twigTags["first_page_start"] = firstPageStart
	twigTags["first_page_end"] = firstPageEnd
	twigTags["last_page_start"] = lastPageStart
	twigTags["last_page_end"] = lastPageEnd
	if page > 1 {
		twigTags["prev_page_start"] = prevPageStart.Format(formatStringDMonthYear)
		twigTags["prev_page_end"] = prevPageEnd.Format(formatStringDMonthYear)
	}
	*prevPageStart = posts[0].CreatedDate
	*prevPageEnd = posts[chunkSize-1].CreatedDate
	*feed = (*feed)[chunkSize:]
	if !lastPage {
		feedLen := len(*feed)
		twigTags["next_page_start"] = (*feed)[0].CreatedDate.Format(formatStringDMonthYear)
		if feedLen < chunkSize {
			twigTags["next_page_end"] = (*feed)[feedLen-1].CreatedDate.Format(formatStringDMonthYear)
			chunkSize = feedLen
		} else {
			twigTags["next_page_end"] = (*feed)[chunkSize-1].CreatedDate.Format(formatStringDMonthYear)
		}
	}

	buf := bytes.NewBufferString("")
	if err := env.Execute(
		"list.html.twig",
		buf,
		twigTags); err != nil {
		log.Fatal(err)
	}
	err := ioutil.WriteFile(fmt.Sprintf("%s%s-%d.html", ConfigData.BaseDir, filenamePrefix, page), buf.Bytes(), 0777)
	return err
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
		err := TwigifyPage(
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
	tDir := ConfigData.TemplateDir
	env := twig.New(stick.NewFilesystemLoader(tDir))
	env.Filters["tag_link"] = filterTagLink
	buf := bytes.NewBufferString("")
	if err := env.Execute(
		"latest-article.html.twig",
		buf,
		toTwigVariables(&entry, "")); err != nil {
		log.Fatal(err)
	}
	err := ioutil.WriteFile(
		fmt.Sprintf("%s%s.html", ConfigData.BaseDir, "latest-post"),
		buf.Bytes(),
		0644)
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
	err := filepath.Walk(dir+"/media",
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
	err = filepath.Walk(dir+"/posts",
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
	targetDir, _ := path.Split(destination)
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatalf("Failed making root dirs for %s, %v\n", targetDir, err)
		}
	}
	var data, err1 = ioutil.ReadFile(source)
	if err1 != nil {
		return err1
	}
	return ioutil.WriteFile(destination, data, 0777)
}
