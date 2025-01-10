package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func postWantsBlueskyCrosspost(fm FrontMatter) bool {
	return fm.SyndicationLinks.Bluesky == "XPOST"
}

func setBlueskyLink(filename string, link string) {
	filename = filepath.Join(ConfigData.RepositoryDir, filename)
	mep, err := os.ReadFile(filename)
	if err == nil {
		replc := regexp.MustCompile(`Bluesky:[ '"]*XPOST[ '"]*`)
		mep := replc.ReplaceAll(mep, []byte(fmt.Sprintf(`Bluesky: "%s"`, link)))
		os.WriteFile(filename, mep, 0777)
	}
}

func loginToBluesky() string {
	type blueskyLoginResponse struct {
		AccessJWT  string `json:"accessJwt"`
		RefreshJWT string `json:"refreshJwt"`
	}

	buffer, _ := json.Marshal(struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}{Identifier: ConfigData.Syndication.Bluesky.Userid, Password: ConfigData.Syndication.Bluesky.Password})
	request, _ := http.NewRequest(
		"POST",
		ConfigData.Syndication.Bluesky.URL+"xrpc/com.atproto.server.createSession",
		bytes.NewBuffer(buffer),
	)
	request.Header.Set("Content-type", "application/json")
	resp, err := Client.Do(request)
	if err != nil {
		log.Fatal(err.Error())
	}

	var res blueskyLoginResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if resp.StatusCode != 200 {
		log.Fatalf("status code was %d", resp.StatusCode)
	}
	if res.AccessJWT != "" {
		return res.AccessJWT
	} else {
		log.Fatalf("failed in post to bluesky attempt %s|%d", res, resp.StatusCode)
	}
	return ""
}

type indexStruct struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}
type featureStruct struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}
type facetStruct struct {
	Index    indexStruct     `json:"index"`
	Features []featureStruct `json:"features"`
}

func makeBlueskyPost(frontmatter *FrontMatter) (string, []facetStruct) {
	toSyndicate := frontmatter.Synopsis
	facets := []facetStruct{}
	posttype := strings.ToLower(frontmatter.Type)
	if posttype == "indieweb" {
		prefix := "\n"
		for _, x := range [][]string{
			{frontmatter.InReplyTo, "In reply to"},
			{frontmatter.RepostOf, "Repost of"},
			{frontmatter.LikeOf, "Like of"},
			{frontmatter.FavoriteOf, "Favourite of"},
			{frontmatter.BookmarkOf, "Bookmark of"},
		} {
			if len(x[0]) > 0 {
				toSyndicate += prefix + "\n" + x[1] + ": "
				prefix = ""
				start := len(toSyndicate)
				toSyndicate = toSyndicate + x[0]
				facets = append(facets, facetStruct{
					Index: indexStruct{
						ByteStart: start,
						ByteEnd:   len(toSyndicate),
					},
					Features: []featureStruct{
						{Type: "app.bsky.richtext.facet#link", URI: x[0]},
					},
				})
			}
		}
	} else if posttype != "tweet" && posttype != "toot" {
		start := len(toSyndicate)
		toSyndicate = toSyndicate + "\r\n\r\n" + frontmatter.Link
		facets = append(facets, facetStruct{
			Index: indexStruct{
				ByteStart: start,
				ByteEnd:   len(toSyndicate),
			},
			Features: []featureStruct{
				{Type: "app.bsky.richtext.facet#link", URI: frontmatter.Link},
			},
		})
	}
	if len(frontmatter.Tags) > 0 {
		toSyndicate = toSyndicate + "\n#" + strings.Join(frontmatter.Tags, " #")
	}
	return toSyndicate, facets
}

func postToBluesky(message string, facets []facetStruct, createdAt time.Time) (string, error) {
	type blueskyPostResponse struct {
		URI string `json:"uri"`
		Cid string `json:"cid"`
	}

	type blueskyPostRecord struct {
		Text      string        `json:"text"`
		Facets    []facetStruct `json:"facets"`
		CreatedAt string        `json:"createdAt"`
	}
	type blueskyPostPackage struct {
		Repo       string            `json:"repo"`
		Collection string            `json:"collection"`
		Record     blueskyPostRecord `json:"record"`
	}

	token := loginToBluesky()

	data := blueskyPostPackage{
		Repo:       ConfigData.Syndication.Bluesky.Userid,
		Collection: "app.bsky.feed.post",
		Record: blueskyPostRecord{
			Text:      message,
			CreatedAt: createdAt.Format(time.RFC3339),
			Facets:    facets,
		},
	}
	buffer, _ := json.Marshal(data)
	PrintIfNotSilent(string(buffer))

	request, _ := http.NewRequest(
		"POST",
		ConfigData.Syndication.Bluesky.URL+"xrpc/com.atproto.repo.createRecord",
		bytes.NewBuffer(buffer),
	)
	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	resp, err := Client.Do(request)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed in posting to bluesky %s[%d]", string(respBytes), resp.StatusCode)
	}
	var res blueskyPostResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if res.URI != "" {
		bits := strings.Split(res.URI, "/")
		newURI := fmt.Sprintf("https://bsky.app/profile/vonexplaino.com/post/%s", bits[len(bits)-1])
		return newURI, nil
	} else {
		return "", fmt.Errorf("failed in post to bluesky attempt %s|%d", res, resp.StatusCode)
	}
}
