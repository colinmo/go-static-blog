package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func errorResponse(code int, msg string) {
	fmt.Printf("Status:%d %s\r\n", code, http.StatusText(code))
	fmt.Printf("Content-Type: text/plain\r\n")
	fmt.Printf("\r\n")
	fmt.Printf("%s\r\n", msg)
	os.Exit(-99)
}

type Indieweb struct {
	InReplyTo  string `yaml:"in-reply-to"`
	BookmarkOf string `yaml:"bookmark-of"`
	FavoriteOf string `yaml:"favorite-of"`
	RepostOf   string `yaml:"repost-of"`
	LikeOf     string `yaml:"like-of"`
}

type FrontMatter struct {
	ID            string   `yaml:"Id"`
	Title         string   `yaml:"Title"`
	Tags          []string `yaml:"Tags"`
	Created       string   `yaml:"Created"`
	Updated       string   `yaml:"Updated"`
	Type          string   `yaml:"Type"`
	Status        string   `yaml:"Status"`
	Synopsis      string   `yaml:"Synopsis"`
	Author        string   `yaml:"Author"`
	FeatureImage  string   `yaml:"FeatureImage"`
	AttachedMedia []string `yaml:"AttachedMedia"`
	IndieWeb      Indieweb `yaml:"IndieWeb"`
	Slug          string   `yaml:"Slug"`
	Link          string   `yaml:"Link"`
	InReplyTo     string   `yaml:"in-reply-to"`
	BookmarkOf    string   `yaml:"bookmark-of"`
	FavoriteOf    string   `yaml:"favorite-of"`
	RepostOf      string   `yaml:"repost-of"`
	LikeOf        string   `yaml:"like-of"`
	RelativeLink  string
	CreatedDate   time.Time
	UpdatedDate   time.Time
}

/*
 * Should I detect no password fields
 * so show a GET form, instead?
 */
func main() {
	// Receive a request
	var req *http.Request
	var err error
	var messages []string
	req, err = cgi.Request()
	if err != nil {
		errorResponse(500, "parse: cannot get cgi request"+err.Error())
	}

	// Use req to handle request
	messages = append(messages, "Processing the submission")
	req.ParseForm()
	if len(req.PostFormValue("username")) == 0 || len(req.PostFormValue("password")) == 0 {
		errorResponse(500, "parse: username and password is required")
	}

	postType := "indieweb"
	if req.PostFormValue("article") == "article" {
		postType = "article"
	}
	link := req.PostFormValue("link")
	timeNow := time.Now()
	frontMatter := FrontMatter{
		Title:       req.PostFormValue("title"),
		Tags:        strings.Split(req.PostFormValue("tags"), ","),
		Created:     timeNow.Format("2006-01-02T15:04:05-0700"),
		Updated:     timeNow.Format("2006-01-02T15:04:05-0700"),
		Type:        postType,
		Status:      "live",
		Synopsis:    req.PostFormValue("summary"),
		Author:      "Professor von Explaino",
		CreatedDate: timeNow,
		UpdatedDate: timeNow,
	}
	switch req.PostFormValue("indieweb") {
	case "likeof":
		frontMatter.LikeOf = link
		frontMatter.IndieWeb.LikeOf = link
		frontMatter.Title = "Like: " + frontMatter.Title
	case "bookmarkof":
		frontMatter.BookmarkOf = link
		frontMatter.IndieWeb.BookmarkOf = link
		frontMatter.Title = "Bookmark: " + frontMatter.Title
	case "repostof":
		frontMatter.RepostOf = link
		frontMatter.IndieWeb.RepostOf = link
		frontMatter.Title = "Repost: " + frontMatter.Title
	case "favoriteof":
		frontMatter.FavoriteOf = link
		frontMatter.IndieWeb.FavoriteOf = link
		frontMatter.Title = "Favorite: " + frontMatter.Title
	case "inreplyto":
		frontMatter.InReplyTo = link
		frontMatter.IndieWeb.InReplyTo = link
		frontMatter.Title = "InReplyTo: " + frontMatter.Title
	}
	// Safe filename
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	frontMatter.ID = strings.ToLower(string(re.ReplaceAll([]byte(frontMatter.Title), []byte("-"))))
	frontMatter.Slug = frontMatter.ID + ".html"

	errorMessages, err := errorChecking(frontMatter)
	if err != nil {
		errorResponse(400, strings.Join(errorMessages, "\r\n"))
	}
	fmt.Printf("Content-Type: text/plain\r\n")
	if err = sendToBitbucket(
		frontMatter.ID,
		frontMatter.Type,
		createPage(frontMatter, req.PostFormValue("content")),
		req.PostFormValue("username"),
		req.PostFormValue("password")); err == nil {
		fmt.Printf("%s", strings.Join(messages, "\r\n"))
	} else {
		fmt.Printf("Bogus %v\n", err)
	}
}

func createPage(frontmatter FrontMatter, post string) string {
	dude, err := yaml.Marshal(frontmatter)
	if err != nil {
		errorResponse(500, "parse: failed to marshal the frontmatter")
	}
	re := regexp.MustCompile("relativelink.*\ncreated.*\nupdated.*\n")

	return string(
		re.ReplaceAll(
			[]byte(fmt.Sprintf("%s===\n%s", dude, post)), []byte("")))
}

func sendToBitbucket(filename string, articleType string, contents string, username string, password string) error {
	data := url.Values{
		fmt.Sprintf(
			"/posts/%s/%s/%s.md",
			strings.ToLower(articleType),
			time.Now().Format("2006/01"),
			filename): {contents},
		"message": {fmt.Sprintf("%s posting", articleType)},
		"author":  {"Colin Morris <relapse@gmail.com>"},
	}
	req, _ := http.NewRequest(
		"POST",
		"https://api.bitbucket.org/2.0/repositories/vonexplaino/blog/src",
		bytes.NewBuffer([]byte(data.Encode())))
	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))},
		"Content-Type":  {"application/x-www-form-urlencoded"},
	}

	resp, err := Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
	} else {
		errorResponse(400, fmt.Sprintf("post: failed to talk to bitbucket %d - %s\n", resp.StatusCode, http.StatusText(resp.StatusCode)))
	}
	return nil
}

func errorChecking(frontmatter FrontMatter) ([]string, error) {
	var message []string

	if len(frontmatter.Title) == 0 {
		message = append(message, "Title is required")
	}
	if len(frontmatter.Synopsis) == 0 {
		message = append(message, "Synopsis is required")
	}

	if len(message) > 0 {
		return message, errors.New("validation: failed to validate the post")
	}
	return message, nil
}
