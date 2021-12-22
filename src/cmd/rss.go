package cmd

import (
	"encoding/xml"
	"io/ioutil"
	"os"
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
	FeatureImage    string    `xml:"featureImage"`
	PubDateAsDate   time.Time `xml:"-"`
	GUID            string    `xml:"guid"`
	Tags            []string  `xml:"tag"`
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
	Items         []Item   `xml:"item"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
	Xmlns   string   `xml:"rssTags,attr"`
}

func ReadRSS(filename string) (RSS, error) {
	var feed RSS

	xmlFile, err := os.Open(filename)
	if err != nil {
		// Default
		feed = RSS{
			Version: "2.0",
			Channel: Channel{},
			Xmlns:   "https://vonexplaino.com/rssTags",
		}
		return feed, nil
	}
	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "rssTags:", ""))
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "xmlns:", ""))

	xml.Unmarshal(byteValue, &feed)
	for i := range feed.Channel.Items {
		feed.Channel.Items[i].PubDateAsDate, _ = time.Parse(
			//			"Mon, 02 Jan 2006 15:04:05 -0700",
			"2006-01-02 15:04:05 -0700 MST",
			feed.Channel.Items[i].PublicationDate)
	}
	return feed, err
}

func WriteRSS(feed RSS, filename string) error {
	feed.Version = "2.0"
	feed.Channel.Title = ConfigData.Metadata.Title
	feed.Channel.Language = ConfigData.Metadata.Language
	feed.Channel.Link = ConfigData.BaseURL
	feed.Channel.Description = ConfigData.Metadata.Description
	feed.Channel.LastBuildDate = time.Now().Format(time.RFC1123Z)
	feed.Channel.TimeToLive = strconv.Itoa(ConfigData.Metadata.Ttl)
	feed.Channel.WebMaster = ConfigData.Metadata.Webmaster
	feed.Channel.Generator = "Ridiculous Go Homebrew"
	feed.Channel.Copyright = "Creative Commons 3.0 with Attribution"
	// Ensure sorted in reverse date order
	sort.SliceStable(feed.Channel.Items, func(p, q int) bool {
		return feed.Channel.Items[p].PubDateAsDate.After(feed.Channel.Items[q].PubDateAsDate)
	})
	byteValue, _ := xml.MarshalIndent(feed, "", "    ")
	byteValue = []byte(strings.Replace(string(byteValue), `rssTags=""`, `xmlns:atom="http://www.w3.org/2005/Atom" xmlns:featureImage="https://vonexplaino.com/featureImage" xmlns:rssTags="https://vonexplaino.com/rssTags"`, 1))
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "tag>", "rssTags:tag>"))
	byteValue = []byte(strings.Replace(string(byteValue), "<item>", `<atom:link href="https://vonexplaino.com/blog/rss.xml" rel="self" type="application/rss+xml" /><item>`, 1))
	byteValue = []byte(strings.ReplaceAll(string(byteValue), "featureImage>", "featureImage:featureImage>"))

	err := ioutil.WriteFile(ConfigData.BaseDir+filename, append([]byte("<?xml version=\"1.0\"?>\n"), byteValue...), 0777)
	return err
}

func PostToItem(frontmatter FrontMatter) Item {
	return Item{
		XMLName:         xml.Name{Space: "", Local: "Item"},
		Title:           frontmatter.Title,
		Description:     frontmatter.Synopsis,
		FeatureImage:    frontmatter.FeatureImage,
		PublicationDate: frontmatter.CreatedDate.Format(time.RFC1123Z),
		PubDateAsDate:   frontmatter.CreatedDate,
		GUID:            frontmatter.Link,
		Tags:            frontmatter.Tags,
	}
}

func ItemToPost(item Item) FrontMatter {
	return FrontMatter{
		Title:        item.Title,
		Synopsis:     item.Description,
		Created:      item.PublicationDate,
		CreatedDate:  item.PubDateAsDate,
		FeatureImage: item.FeatureImage,
		Link:         item.GUID,
		ID:           item.GUID,
		Tags:         item.Tags,
	}
}
