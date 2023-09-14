package cmd

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	XMLName         xml.Name  `xml:"item"`
	Title           string    `xml:"title"`
	Description     string    `xml:"description"`
	PublicationDate string    `xml:"pubDate"`
	PubDateAsDate   time.Time `xml:"-"`
	GUID            string    `xml:"guid"`
	Tags            []string  `xml:"subject"`
}
type AtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	Length  string   `xml:"length,attr,omitempty"`
}

type Channel struct {
	XMLName       xml.Name `xml:"channel"`
	Title         string   `xml:"title"`
	Link          string   `xml:"link"`
	Description   string   `xml:"description"`
	Language      string   `xml:"language"`
	Copyright     string   `xml:"copyright"`
	LastBuildDate string   `xml:"lastBuildDate"`
	Generator     string   `xml:"generator"`
	WebMaster     string   `xml:"webMaster"`
	TimeToLive    string   `xml:"ttl"`
	AtomLink      AtomLink
	Items         []Item `xml:"item"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
	XmlnsA  string   `xml:"xmlns:atom,attr"`
	XmlnsB  string   `xml:"xmlns:rssTags,attr"`
}

func ReadRSS(filename string) (RSS, error) {
	var feed RSS

	xmlFile, err := os.Open(filename)
	if err != nil {
		// Default
		feed = RSS{
			Version: "2.0",
			Channel: Channel{},
			XmlnsA:  "http://www.w3.org/2005/Atom",
			XmlnsB:  "http://purl.org/dc/elements/1.1/",
		}
		return feed, nil
	}
	defer xmlFile.Close()
	byteValue, _ := io.ReadAll(xmlFile)
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "rssTags:", ""))
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "xmlns:", ""))

	err = xml.Unmarshal(byteValue, &feed)
	if err == nil {
		for i := range feed.Channel.Items {
			feed.Channel.Items[i].PubDateAsDate, _ = time.Parse(
				//"Mon, 02 Jan 2006 15:04:05 -0700",
				time.RFC1123Z,
				//"2006-01-02 15:04:05 -0700 MST",
				feed.Channel.Items[i].PublicationDate)

		}
	}

	return feed, err
}

func WriteRSS(feed RSS, filename string) error {
	feed.Version = "2.0"
	feed.XmlnsA = "http://www.w3.org/2005/Atom"
	feed.XmlnsB = "http://purl.org/dc/elements/1.1/"
	feed.Channel.Title = ConfigData.Metadata.Title
	feed.Channel.Language = ConfigData.Metadata.Language
	feed.Channel.Link = ConfigData.BaseURL
	feed.Channel.Description = ConfigData.Metadata.Description
	feed.Channel.LastBuildDate = time.Now().Format(time.RFC1123Z)
	feed.Channel.TimeToLive = strconv.Itoa(ConfigData.Metadata.Ttl)
	feed.Channel.WebMaster = ConfigData.Metadata.Webmaster
	feed.Channel.Generator = "Ridiculous Go Homebrew"
	feed.Channel.Copyright = "Creative Commons 3.0 with Attribution"
	feed.Channel.AtomLink = AtomLink{
		Href: "https://vonexplaino.com/blog/rss.xml",
		Rel:  "self",
		Type: "application/rss+xml",
	}
	// Ensure sorted in reverse date order
	sort.SliceStable(feed.Channel.Items, func(p, q int) bool {
		return feed.Channel.Items[p].PubDateAsDate.After(feed.Channel.Items[q].PubDateAsDate)
	})
	byteValue, _ := xml.MarshalIndent(feed, "", "    ")
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "subject>", "rssTags:subject>"))

	err := os.WriteFile(filepath.Join(ConfigData.BaseDir, filename), append([]byte("<?xml version=\"1.0\"?>\n"), byteValue...), 0777)
	return err
}

func PostToItem(frontmatter FrontMatter) Item {
	if len(frontmatter.FeatureImage) > 0 && frontmatter.FeatureImage[0:1] == "/" {
		frontmatter.FeatureImage = ConfigData.BaseURL[0:len(ConfigData.BaseURL)-6] + frontmatter.FeatureImage
	}
	return Item{
		XMLName:         xml.Name{Space: "", Local: "Item"},
		Title:           frontmatter.Title,
		Description:     frontmatter.Synopsis,
		PublicationDate: frontmatter.CreatedDate.Format(time.RFC1123Z),
		PubDateAsDate:   frontmatter.CreatedDate,
		GUID:            frontmatter.Link,
		Tags:            frontmatter.Tags,
	}
}

func ItemToPost(item Item) FrontMatter {
	return FrontMatter{
		Title:       item.Title,
		Synopsis:    item.Description,
		Created:     item.PublicationDate,
		CreatedDate: item.PubDateAsDate,
		Link:        item.GUID,
		ID:          item.GUID,
		Tags:        item.Tags,
	}
}
