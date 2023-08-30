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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	html2 "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/spf13/cobra"
	"github.com/tyler-sommer/stick"
	"github.com/tyler-sommer/stick/twig"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

var fromFile *string
var toFile *string
var wordpressThumbnailTemplate = "https://s0.wp.com/mshots/v1/%s?vpw=480&vph=380"

// makepageCmd represents the makepage command
var makepageCmd = &cobra.Command{
	Use:   "makepage",
	Short: "Convert markdown page into html page",
	Long: `Converts a markdown page into an html page:

Uses golang markdown and a local html template file to generate blog posts.`,
	Run: func(cmd *cobra.Command, args []string) {
		var txt2 []byte
		var err error
		var html string
		var frontMatter FrontMatter

		if *fromFile == "" {
			stdin := bufio.NewReader(os.Stdin)
			stdin.Read(txt2)
			html, frontMatter, err = parseString(string(txt2), "")
			if err != nil {
				fmt.Printf("Failed to parse %v", txt2)
				os.Exit(2)
			}
		} else {
			html, frontMatter, err = parseFile(*fromFile)
			if err != nil {
				fmt.Printf("Could not parse the file %s\n", *fromFile)
				os.Exit(2)
			}
		}

		if *toFile == "" {
			*toFile = filepath.Join(ConfigData.BaseDir, frontMatter.Slug+".html")
		}
		os.MkdirAll(filepath.Dir(*toFile), 0755)
		err = os.WriteFile(*toFile, []byte(html), 0744)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(makepageCmd)

	fromFile = makepageCmd.Flags().StringP("from", "f", "", "File to convert from")
	toFile = makepageCmd.Flags().StringP("to", "t", "", "File to convert to")

	// Default Markdown parser
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			meta.New(meta.WithStoresInDocument()),
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html2.WithLineNumbers(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func convertGallery(mep []byte) []byte {
	re := regexp.MustCompile(`(<section [^>]*gallery[^>]*) markdown="1"([^>]*>)(?sm)(.*?)(</[^>]*>)`)
	mep1 := re.FindAllStringSubmatch(string(mep), -1)
	// Convert the image collection from Markdown to HTML
	var buf2 bytes.Buffer
	md.Convert([]byte(mep1[0][3]), &buf2)
	// Convert the individual images into the converted Gallery version
	re2 := regexp.MustCompile(`<a href="([^"]*)"><img src="([^"]*)" alt="([^"]*)" title="([^"]*)"></a>`)
	mep2 := re2.FindAllStringSubmatch(buf2.String(), -1)
	stringOut := ""
	for i := 0; i < len(mep2); i++ {
		stringOut += fmt.Sprintf(`<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-%d" tabindex="-1" />
				<label for="gallery-2020-4-%d-%d">
					<img src="%s" />
				</label>
				<figure>
					<img loading="lazy" src="%s" alt="%s" />
					<figcaption>
						<em>%s</em>
					</figcaption>
				</figure>
		`,
			gallery_index,
			gallery_index,
			i+1,
			gallery_index,
			i,
			mep2[i][2],
			mep2[i][1],
			mep2[i][3],
			mep2[i][4],
		)
	}
	// Print out the converted HTML, removing the "markdown="1"""
	temp := []byte(fmt.Sprintf(
		`%s%s<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-0" />
				<label></label><figure></figure>%s<input type="radio" name="gallery-2020-4-%d" id="gallery-2020-4-%d-close" />
				<label for="gallery-2020-4-%d-close">X</label>%s`,
		mep1[0][1],
		mep1[0][2],
		gallery_index,
		gallery_index,
		stringOut,
		gallery_index,
		gallery_index,
		gallery_index,
		mep1[0][4],
	))
	gallery_index++
	return temp
}

func convertMarkdownHtml(mep []byte) []byte {
	re := regexp.MustCompile(`(<[^>]*) markdown="1"([^>]*>)(?sm)((.|\r|\n)*?)(</[^>]*>)`)
	mep1 := re.FindAllStringSubmatch(string(mep), -1)
	var buf2 bytes.Buffer
	md.Convert([]byte(mep1[0][3]), &buf2)
	html := buf2.String()
	if mep1[0][3][0:1] != "\n" {
		html = html[3 : len(html)-5]
	}

	// Print out the converted HTML, removing the "markdown="1"""
	return []byte(fmt.Sprintf(
		`%s%s%s%s`,
		mep1[0][1],
		mep1[0][2],
		html,
		mep1[0][5],
	))
}

var gallery_index int
var md goldmark.Markdown

func filterTagLink(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return "tag/" + textToSlug(stick.CoerceString(val))
}

func getFirstWords(text string, lineWidth int) string {
	re := regexp.MustCompile(`<p>((.|\r|\n)*?)</p>`)
	texts := re.FindSubmatch([]byte(text))
	if len(texts) == 0 {
		return text
	}
	re = regexp.MustCompile(`<[^>]+>`)
	text = re.ReplaceAllLiteralString(string(texts[0]), "")
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return text
	}
	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			return wrapped + "..."
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return wrapped

}

// parseFile reads in the file provided and returns the html conversion and yaml frontmatter
func parseFile(filename string) (string, FrontMatter, error) {
	var html2 string
	var err error
	var frontMatter FrontMatter

	txt, err := os.ReadFile(filename)
	if err != nil {
		return html2, frontMatter, err
	}
	return parseString(string(txt), filename)
}

// parseString parses the passed string and returns the html conversion and yaml frontmatter
func parseString(body string, filename string) (string, FrontMatter, error) {
	var html2 string
	var err error
	var frontMatter FrontMatter

	// Parse the frontmatter at the start of the file
	split := strings.SplitN(body[3:], "---", 2)
	if len(split) != 2 {
		return html2, frontMatter, err
	}
	frontMatter, err = parseFrontMatter(split[0], filename)
	if err != nil {
		return html2, frontMatter, err
	}

	// Convert the Gallery tags
	var buf2 bytes.Buffer
	re := regexp.MustCompile(`<section [^>]*gallery[^>]* markdown="1"[^>]*>(?sm)(.*?)</section>`)
	gallery_index = 0
	bodybyte := re.ReplaceAllFunc(
		[]byte(body),
		convertGallery,
	)
	// Convert the markdown=1 tags
	re = regexp.MustCompile(`<[^>]* markdown="1"[^>]*>(?sm)(.*?)</[^>]*>`)
	bodybyte = re.ReplaceAllFunc(
		bodybyte,
		convertMarkdownHtml,
	)
	// Now convert what's left to Markdown
	md.Convert(bodybyte, &buf2)
	html2 = buf2.String()

	// Synopsis if empty
	if len(frontMatter.Synopsis) == 0 {
		frontMatter.Synopsis = getFirstWords(html2, 310)
	}

	// Run HTML into Template
	articleType := fmt.Sprintf("%v", frontMatter.Type)
	d, _ := os.Getwd()
	tDir := filepath.Join(d, "templates")
	if len(ConfigData.TemplateDir) > 0 {
		tDir = ConfigData.TemplateDir
	}
	env := twig.New(stick.NewFilesystemLoader(tDir))

	buf := bytes.NewBufferString("")
	env.Filters["tag_link"] = filterTagLink
	if err := env.Execute(
		strings.ToLower(articleType)+".html.twig",
		buf,
		toTwigVariables(&frontMatter, html2)); err != nil {
		fmt.Printf("Couldn't write the file\n")
		log.Fatal(err)
	}

	html2 = buf.String()

	return html2, frontMatter, err
}

type Event struct {
	Start     string `yaml:"StartDate"`
	End       string `yaml:"EndDate"`
	StartDate time.Time
	EndDate   time.Time
	Status    string `yaml:"Status"`
	Location  string `yaml:"Location"`
}

type Contact struct {
	Name      string `yaml:"name"`
	Honorific string `yaml:"honorific"`
	Email     string `yaml:"email"`
	Photo     string `yaml:"u-photo"`
	URL       string `yaml:"u-url"`
	Key       string `yaml:"u-key"`
	LinkedIn  string `yaml:"linkedin"`
	Logo      string `yaml:"u-logo"`
	Title     string `yaml:"p-job-title"`
}

type Education struct {
	Name      string `yaml:"p-name"`
	Start     string `yaml:"dt-start"`
	End       string `yaml:"dt-end"`
	StartDate time.Time
	EndDate   time.Time
	URL       string `yaml:"u-url"`
	Category  string `yaml:"p-category"`
	Location  string `yaml:"p-location"`
}

type Experience struct {
	Name          string `yaml:"p-name"`
	Summary       string `yaml:"p-summary"`
	Start         string `yaml:"dt-start"`
	StartDate     time.Time
	Description   string `yaml:"p-description"`
	URL           string `yaml:"u-url"`
	Location      string `yaml:"p-location"`
	Category      string `yaml:"p-category"`
	Published     string `yaml:"dt-published"`
	PublishedDate time.Time
	Author        string `yaml:"p-author"`
}

type TimedExperience struct {
	FivePlus  []string `yaml:"5+ years"`
	OneToFive []string `yaml:"1-5 years"`
	New       []string `yaml:"<1 year"`
}

type SkillGroup struct {
	Name    string   `yaml:"name"`
	Members []string `yaml:"members"`
}

type Skill struct {
	SeniorDev      []SkillGroup    `yaml:"seniordev"`
	Developer      []SkillGroup    `yaml:"developer"`
	Intern         []SkillGroup    `yaml:"intern"`
	HobbyPro       []SkillGroup    `yaml:"hobbypro"`
	Hobbiest       []SkillGroup    `yaml:"hobbiest"`
	Dabbler        []SkillGroup    `yaml:"dabbler"`
	Programming    TimedExperience `yaml:"Programming languages"`
	Libraries      TimedExperience `yaml:"Libraries/ services/ technologies"`
	Accreditations []string        `yaml:"Principal methodology accreditations"`
}
type Resume struct {
	Contact     Contact      `yaml:"Contact"`
	Education   []Education  `yaml:"Education"`
	Experience  []Experience `yaml:"Experience"`
	Skill       Skill        `yaml:"Skill"`
	Affiliation []string     `yaml:"Affiliation"`
}

type SyndicationLinksS struct {
	Twitter   string `yaml:"Twitter"`
	Instagram string `yaml:"Instagram"`
	Mastodon  string `yaml:"Mastodon"`
}

type ItemS struct {
	URL    string  `yaml:"url"`
	Image  string  `yaml:"image"`
	Name   string  `yaml:"name"`
	Type   string  `yaml:"type"`
	Rating float32 `yaml:"rating"`
}

type FrontMatter struct {
	ID               string            `yaml:"Id"`
	Title            string            `yaml:"Title"`
	Tags             []string          `yaml:"Tags"`
	Created          string            `yaml:"Created"`
	Updated          string            `yaml:"Updated"`
	Type             string            `yaml:"Type"`
	Status           string            `yaml:"Status"`
	Synopsis         string            `yaml:"Synopsis"`
	Author           string            `yaml:"Author"`
	FeatureImage     string            `yaml:"FeatureImage"`
	AttachedMedia    []string          `yaml:"AttachedMedia"`
	SyndicationLinks SyndicationLinksS `yaml:"Syndication"`
	Slug             string            `yaml:"Slug"`
	Event            Event             `yaml:"Event"`
	Resume           Resume            `yaml:"Resume"`
	Link             string            `yaml:"Link"`
	InReplyTo        string            `yaml:"in-reply-to"`
	BookmarkOf       string            `yaml:"bookmark-of"`
	FavoriteOf       string            `yaml:"favorite-of"`
	RepostOf         string            `yaml:"repost-of"`
	LikeOf           string            `yaml:"like-of"`
	Item             ItemS             `yaml:"Item"`
	RelativeLink     string
	CreatedDate      time.Time
	UpdatedDate      time.Time
}

func textToSlug(intext string) string {
	re := regexp.MustCompile("[^.a-zA-Z0-9-]")
	slug := strings.ToLower(re.ReplaceAllString(intext, "-"))
	re = regexp.MustCompile("-+")
	slug = re.ReplaceAllString(slug, "-")
	re = regexp.MustCompile("-+$")
	slug = re.ReplaceAllString(slug, "")
	re = regexp.MustCompile("^-+")
	slug = re.ReplaceAllString(slug, "")
	return slug
}

func setEmptyStringDefault(value string, ifempty string) string {
	if len(value) == 0 {
		return ifempty
	}
	return value
}

func frontMatterDefaults(frontMatter *FrontMatter, filename string) {
	created, err2 := parseUnknownDateFormat(frontMatter.Created)
	if err2 != nil {
		created = time.Now()
		frontMatter.Created = created.Format("2006-01-02T15:04:05-0700")
	}
	frontMatter.CreatedDate = created
	if frontMatter.Updated == "" {
		frontMatter.Updated = frontMatter.Created
		frontMatter.UpdatedDate = frontMatter.CreatedDate
	} else {
		updated, err2 := parseUnknownDateFormat(frontMatter.Updated)
		if err2 != nil {
			frontMatter.Updated = frontMatter.Created
		} else {
			frontMatter.UpdatedDate = updated
			frontMatter.Updated = updated.Format("2006-01-02T15:04:05-0700")
		}
	}

	frontMatter.Slug = setEmptyStringDefault(frontMatter.Slug, textToSlug(frontMatter.Title))
	ext := filepath.Ext(frontMatter.Slug)
	if ext != ".html" {
		frontMatter.Slug = frontMatter.Slug + ".html"
	}
	frontMatter.Status = setEmptyStringDefault(frontMatter.Status, "live")

	if len(frontMatter.Tags) == 0 {
		frontMatter.Tags = []string{}
	}
	if len(frontMatter.ID) == 0 && len(filename) > len(ConfigData.RepositoryDir) {
		frontMatter.ID = filename[len(ConfigData.RepositoryDir):]
	}
}

func frontMatterValidateExperience(frontMatter *FrontMatter) {
	for i, x := range frontMatter.Resume.Experience {
		if len(x.Description) > 0 {
			var buf2 bytes.Buffer
			md.Convert([]byte(x.Description), &buf2)
			frontMatter.Resume.Experience[i].Description = buf2.String()
		}
		if len(x.Summary) > 0 {
			var buf2 bytes.Buffer
			md.Convert([]byte(x.Summary), &buf2)
			frontMatter.Resume.Experience[i].Summary = strings.Replace(strings.Replace(buf2.String(), "<p>", "", 1), "</p>", "", 1)
		}
		frontMatter.Resume.Experience[i].StartDate, _ = parseUnknownDateFormat(x.Start)
		frontMatter.Resume.Experience[i].PublishedDate, _ = parseUnknownDateFormat(x.Published)
	}
	for i, d := range frontMatter.Resume.Education {
		frontMatter.Resume.Education[i].StartDate, _ = parseUnknownDateFormat(d.Start)
		frontMatter.Resume.Education[i].EndDate, _ = parseUnknownDateFormat(d.End)
	}
}

func frontMatterValidate(frontMatter *FrontMatter, filename string) []string {
	var collectedErrors []string
	// Valids
	validTypes := []string{"article", "reply", "indieweb", "tweet", "resume", "event", "page", "review"}
	if filename != "" && frontMatter.Type == "" {
		frontMatter.Type = defaultType(validTypes, filename)
	} else {
		frontMatter.Type = strings.ToLower(frontMatter.Type)
	}
	if !contains(validTypes, frontMatter.Type) {
		collectedErrors = append(collectedErrors, "bad type: "+frontMatter.Type)
	}
	if !contains([]string{"draft", "live", "retired"}, frontMatter.Status) {
		collectedErrors = append(collectedErrors, "bad status: "+frontMatter.Status)
	}
	// Need to do this after Type is validated
	if frontMatter.Link == "" {
		if frontMatter.Type == "page" {
			frontMatter.Link, _ = url.JoinPath(ConfigData.BaseURL, baseDirectoryForPosts, strings.ToLower(frontMatter.Type), frontMatter.Slug)
		} else {
			frontMatter.Link, _ = url.JoinPath(ConfigData.BaseURL, baseDirectoryForPosts, strings.ToLower(frontMatter.Type), frontMatter.CreatedDate.Format("2006/01"), frontMatter.Slug)
		}
	}
	// Need to do this after Link is created
	if len(frontMatter.FeatureImage) == 0 {
		frontMatter.FeatureImage = defaultFeatureImage(frontMatter)
	}
	var splitted = strings.Split(frontMatter.Link, "/posts")
	if len(splitted) < 2 {
		collectedErrors = append(collectedErrors, fmt.Sprintf("Could not get a posts link for %s", frontMatter.Link))
	} else {
		frontMatter.RelativeLink = splitted[1]
	}
	if len(frontMatter.Resume.Contact.Name) > 0 {
		frontMatterValidateExperience(frontMatter)
	}
	return collectedErrors
}
func defaultType(validTypes []string, filename string) string {
	re := regexp.MustCompile(`[/\\]posts[/\\](` + strings.Join(validTypes, "|") + `)[/\\]`)
	indexes := re.FindStringSubmatch(filename)
	if indexes != nil {
		return strings.ToLower(indexes[1])
	}
	return ""
}

func defaultFeatureImage(frontMatter *FrontMatter) string {
	if len(frontMatter.InReplyTo) > 0 {
		return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.InReplyTo))
	} else if len(frontMatter.BookmarkOf) > 0 {
		return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.BookmarkOf))
	} else if len(frontMatter.FavoriteOf) > 0 {
		return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.FavoriteOf))
	} else if len(frontMatter.RepostOf) > 0 {
		return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.RepostOf))
	} else if len(frontMatter.LikeOf) > 0 {
		return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.LikeOf))
	}
	return fmt.Sprintf(wordpressThumbnailTemplate, url.QueryEscape(frontMatter.Link))
}

func parseFrontMatter(inFrontMatter string, filename string) (FrontMatter, error) {
	var frontMatter FrontMatter
	err := yaml.Unmarshal([]byte(inFrontMatter), &frontMatter)
	if err != nil {
		fmt.Printf("Failed to parse frontmatter %s\n", inFrontMatter)
		log.Fatal(err)
		return frontMatter, err
	}
	frontMatterDefaults(&frontMatter, filename)
	collectedErrors := frontMatterValidate(&frontMatter, filename)

	if len(collectedErrors) > 0 {
		err = errors.New(strings.Join(collectedErrors, ", "))
	}
	return frontMatter, err
}

func parseUnknownTimezone(dateString string) *time.Location {
	// Timezones
	var l *time.Location
	if dateString[len(dateString)-1:] == "Z" {
		l, _ = time.LoadLocation("UTC")
		return l
	}
	re := regexp.MustCompile(`([+-]\d{4}|[+-]\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(dateString)
	hrplus, minplus := 0, 0
	if matches != nil {
		if len(matches[1]) == 5 {
			hrplus, _ = strconv.Atoi(matches[1][0:3])
			minplus, _ = strconv.Atoi(matches[1][3:5])
		} else {
			hrplus, _ = strconv.Atoi(matches[1][0:3])
			minplus, _ = strconv.Atoi(matches[1][4:6])
		}
	} else {
		re = regexp.MustCompile(`([A-Z]{3}[+-]\d\d)`)
		matches = re.FindStringSubmatch(dateString)
		if matches != nil {
			hrplus, _ = strconv.Atoi(matches[1][4:])
		}
	}
	l = time.FixedZone("postzone", hrplus*3600+minplus*60)
	return l
}

func parseUnknownTime(dateString string, re *regexp.Regexp) (int, int, int) {
	hr := 0
	mi := 0
	se := 0
	// Time
	matches := re.FindStringSubmatch(dateString)
	if matches == nil {
		matches = re.FindStringSubmatch(dateString)
		if matches != nil {
			hr, _ = strconv.Atoi(matches[1])
			mi, _ = strconv.Atoi(matches[2])
			se = 0
			if strings.ToLower(matches[3]) == "pm" {
				hr += 12
			}
		}
	} else {
		hr, _ = strconv.Atoi(matches[1])
		mi, _ = strconv.Atoi(matches[2])
		se, _ = strconv.Atoi(matches[3])
	}
	return hr, mi, se
}

func parseUnknownDate(dateString string) (int, int, int, error) {
	yr, mn, dy := 0, 0, 0
	var newTime time.Time
	// Date
	re := regexp.MustCompile(`(\d{1,2})\s*(\w{3})\s*(\d{4})`)
	date := re.FindStringSubmatch(dateString)
	if date == nil {
		re = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
		date = re.FindStringSubmatch(dateString)
		if date == nil {
			re = regexp.MustCompile(`(\w{3})\w*\s+(\d{1,2})[,\s]+(\d{4})`)
			date = re.FindStringSubmatch(dateString)
			if date == nil {
				return yr, mn, dy, errors.New("could not parse date")
			} else {
				dy, _ = strconv.Atoi(date[2])
				yr, _ = strconv.Atoi(date[3])
				newTime, _ = time.Parse("Jan", date[1])
				mn = int(newTime.Month())
			}
		} else {
			dy, _ = strconv.Atoi(date[3])
			yr, _ = strconv.Atoi(date[1])
			mn, _ = strconv.Atoi(date[2])

		}
	} else {
		dy, _ = strconv.Atoi(date[1])
		yr, _ = strconv.Atoi(date[3])
		newTime, _ = time.Parse("Jan", date[2])
		mn = int(newTime.Month())
	}
	return yr, mn, dy, nil
}

func parseUnknownDateFormat(dateString string) (time.Time, error) {
	var newTime time.Time
	var err error
	var hr, mi, se, dy, yr int
	var l *time.Location
	var mn int

	if len(dateString) == 0 {
		return newTime, err
	}
	l = parseUnknownTimezone(dateString)
	re := regexp.MustCompile(`(\d{1,2}):(\d{1,2})[: ]((\d{1,2})|([ap]m))`)
	hr, mi, se = parseUnknownTime(dateString, re)
	dateString = re.ReplaceAllString(dateString, " ")
	yr, mn, dy, err = parseUnknownDate(dateString)
	// Create date in specified timezone
	newTime = time.Date(yr, time.Month(mn), dy, hr, mi, se, 0, l)
	// Convert to blog timezone
	loc, _ := time.LoadLocation(blogTimezone)
	newTime = newTime.In(loc)

	return newTime, err
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func toTwigVariables(frontMatter *FrontMatter, content string) map[string]stick.Value {

	if frontMatter.Link == "" {
		frontMatter.Link, _ = url.JoinPath(ConfigData.BaseURL, baseDirectoryForPosts, strings.ToLower(frontMatter.Type), frontMatter.CreatedDate.Format("2006/01/02"), frontMatter.Slug)
	}

	return map[string]stick.Value{
		"id":               frontMatter.ID,
		"title":            frontMatter.Title,
		"tags":             frontMatter.Tags,
		"synopsis":         frontMatter.Synopsis,
		"created_date":     frontMatter.CreatedDate,
		"updated_date":     frontMatter.UpdatedDate,
		"type":             frontMatter.Type,
		"featureimage":     frontMatter.FeatureImage,
		"content":          content,
		"link":             frontMatter.Link,
		"inreplyto":        frontMatter.InReplyTo,
		"bookmarkof":       frontMatter.BookmarkOf,
		"likeof":           frontMatter.LikeOf,
		"favoriteof":       frontMatter.FavoriteOf,
		"repostof":         frontMatter.RepostOf,
		"resume":           frontMatter.Resume,
		"item":             frontMatter.Item,
		"syndicationlinks": frontMatter.SyndicationLinks,
	}
}

func toTwigListVariables(frontMatters []FrontMatter, title string, page int) map[string]stick.Value {

	x := make([]map[string]stick.Value, 0)
	for _, mep := range frontMatters {
		x = append(x, map[string]stick.Value{
			"id":               mep.ID,
			"title":            mep.Title,
			"tags":             mep.Tags,
			"created":          mep.Created,
			"updated":          mep.Updated,
			"type":             mep.Type,
			"status":           mep.Status,
			"synopsis":         mep.Synopsis,
			"author":           mep.Author,
			"featureimage":     mep.FeatureImage,
			"attachedmedia":    mep.AttachedMedia,
			"syndicationlinks": mep.SyndicationLinks,
			"slug":             mep.Slug,
			"event":            mep.Event,
			"resume":           mep.Resume,
			"link":             mep.Link,
			"inreplyto":        mep.InReplyTo,
			"bookmarkof":       mep.BookmarkOf,
			"favoriteof":       mep.FavoriteOf,
			"repostof":         mep.RepostOf,
			"likeof":           mep.LikeOf,
			"relativelink":     mep.RelativeLink,
			"createddate":      mep.CreatedDate,
			"updateddate":      mep.UpdatedDate,
			"item":             mep.Item,
		})
	}
	baseDir, _ := url.JoinPath(ConfigData.BaseURL, "posts/")
	return map[string]stick.Value{
		"title":       title + " Page " + strconv.Itoa(page),
		"page":        page,
		"link_prefix": baseDir,
		"list":        x,
	}
}
