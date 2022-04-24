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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Days int
var SVGOptions struct {
	All       bool
	Codestats bool
	Trakt     bool
	Blog      bool
	Feedly    bool
	Withings  bool
	Steps     bool
}

var relativeDataLocation = "../regenerate/data/"

func blogstatsDefaults() {
	// Defaults
	if Days == 0 {
		Days = 30
	}
	if SVGOptions.All {
		SVGOptions.Codestats = true
		SVGOptions.Trakt = true
		SVGOptions.Blog = true
		SVGOptions.Feedly = true
		SVGOptions.Withings = true
		SVGOptions.Steps = true
	}
}

func blogstatsStart() error {
	// Process
	var blogErr, traktErr, statsErr, feedErr, withErr error
	if SVGOptions.Blog {
		blogErr = generateBlogStats()
	}
	if SVGOptions.Trakt {
		generateTraktStats()
	}
	if SVGOptions.Codestats {
		generateCSStats()
	}
	if SVGOptions.Feedly {
		generateFeedlyStats()
	}
	if SVGOptions.Withings {
		generateWithingsStats()
	}
	if !(blogErr == nil && traktErr == nil && statsErr == nil && feedErr == nil && withErr == nil) {
		fmt.Printf("Failures occurred in generation\n")
		return errors.New("failed in generation attempt")
	}
	if !(SVGOptions.Blog || SVGOptions.Trakt || SVGOptions.Codestats || SVGOptions.Feedly || SVGOptions.Withings) {
		fmt.Printf("No options selected\n")
		return errors.New("nothing to do")
	}
	return nil
}

// blogstatsCmd represents the Blog Stats command
var blogstatsCmd = &cobra.Command{
	Use:   "blogstats",
	Short: "Regenerates blog stats svg",
	Long:  `Reads the XML file for the site and generates the blog stats for the last X days`,
	Run: func(cmd *cobra.Command, args []string) {
		blogstatsDefaults()
		blogstatsStart()
	},
}

func generateBlogStats() error {
	filenameOfBlogSvg := ConfigData.BaseDir + "../regenerate/data/blog.svg"
	// Read the XML file
	known, err := ReadRSS(ConfigData.BaseDir + "rss.xml")
	if err != nil {
		// Empty
		known = RSS{}
	}
	days, max := getDaysArray(known)
	chart := barSVG(days, max, 0, -1)
	// Create the SVG
	return ioutil.WriteFile(filenameOfBlogSvg, chart, 0777)
}

type ShowAndMovie struct {
	Show  map[string]string `json:"show"`
	Movie map[string]string `json:"movie"`
}
type TraktStats struct {
	LastUpdated     string `json:"last_updated"`
	LastUpdatedDate time.Time
	Values          map[string]ShowAndMovie `json:"values"`
}

//
type IDsResponse struct {
	Trakt  int    `json:"trakt"`
	TVDB   int    `json:"tvdb"`
	IMDB   string `json:"imdb"`
	TMDB   int    `json:"tmdb"`
	TVRage int    `json:"tvrage"`
	Slug   string `json:"slug"`
}
type ShowResponse struct {
	Title string      `json:"title"`
	Year  int         `json:"year"`
	IDs   IDsResponse `json:"ids"`
}
type TraktResponse struct {
	ID            int64  `json:"id"`
	WatchedAt     string `json:"watched_at"`
	WatchedAtDate time.Time
	Action        string `json:"action"`
	Type          string `json:"type"`
	Episode       struct {
		Season int         `json:"season"`
		Number int         `json:"number"`
		Title  string      `json:"title"`
		IDs    IDsResponse `json:"ids"`
	} `json:"episode"`
	Show  ShowResponse `json:"show"`
	Movie ShowResponse `json:"movie"`
}

func generateTraktStats() {
	filenameOfTraktSvg := ConfigData.BaseDir + "../regenerate/data/trakt-cache.json.svg"
	filenameOfTraktCache := ConfigData.BaseDir + "../regenerate/data/trakt-cache.json"
	stats := readTraktStatsFile(filenameOfTraktCache)
	// Update cached info from source
	stats = updateTraktStats(stats)
	writeTraktStatsFile(filenameOfTraktCache, stats)
	// Get two lines
	movies, shows, max, min := getTraktStatsForDays(Days, stats)
	line1, total1 := lineAlone(shows, max, min, map[string]string{"strokeDashArray": "", "stroke": "black"}, "Shows", map[string]bool{"showZero": true, "showBall": false})
	line2, total2 := lineAlone(movies, max, min, map[string]string{"strokeDashArray": "2", "stroke": "black"}, "Movies", map[string]bool{"showZero": true, "showBall": false})
	// Create and store the SVG
	graph := SVGGraphFromPaths(total1+total2, fmt.Sprintf("%d,%d", int(total1), int(total2)), -1, line1+line2)
	ioutil.WriteFile(filenameOfTraktSvg, graph, 0666)
}

func readTraktStatsFile(filename string) TraktStats {
	// Get the cached info
	cacheFile, err := os.Open(filename)
	buff := TraktStats{}
	if err == nil {
		defer cacheFile.Close()
		byteValue, _ := ioutil.ReadAll(cacheFile)
		err := json.Unmarshal(byteValue, &buff)
		if err != nil {
			log.Fatalf("Failed to parse the Trakt cache %v\n", err)
		}
		if len(buff.LastUpdated) > 0 {
			buff.LastUpdatedDate, _ = parseUnknownDateFormat(buff.LastUpdated)
		}
		if buff.Values == nil {
			buff.Values = make(map[string]ShowAndMovie)
		}
	} else {
		l, _ := time.LoadLocation(blogTimezone)
		buff.LastUpdatedDate = time.Date(1970, time.January, 1, 0, 0, 0, 0, l)
		buff.LastUpdated = buff.LastUpdatedDate.Format(gmtDateFormat)
		buff.Values = make(map[string]ShowAndMovie)
	}
	return buff
}

func writeTraktStatsFile(filename string, stats TraktStats) error {
	marshalled, err := json.Marshal(stats)
	if err == nil {
		ioutil.WriteFile(filename, marshalled, 0666)
	}
	return err
}

func buildTraktLinkClient(startAt string, page int, limit int) []byte {
	url := fmt.Sprintf("https://api.trakt.tv/users/colinmo/history/?start_at=%s&page=%d&limit=%d",
		startAt,
		page,
		limit,
	)
	request, _ := http.NewRequest(
		"GET",
		url,
		bytes.NewBuffer([]byte{}),
	)
	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set(jsonHeaders[1][0], jsonHeaders[1][1])
	request.Header.Set("trakt-api-version", "2")
	request.Header.Set("trakt-api-key", ConfigData.AboutMe.Trakt.ID)

	resp, err := Client.Do(request)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	if len(body) == 0 {
		log.Fatal("Failed to get contents from Trakt\n")
	}
	return body
}

func getMovieShowValues(x TraktResponse, stats *TraktStats) {
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	x.WatchedAtDate, _ = parseUnknownDateFormat(x.WatchedAt)
	if stats.LastUpdatedDate.Before(x.WatchedAtDate) {
		stats.LastUpdatedDate = x.WatchedAtDate
		stats.LastUpdated = stats.LastUpdatedDate.Format(gmtDateFormat)
	}
	id := fmt.Sprintf("%d", (int(math.Ceil(x.WatchedAtDate.Sub(startOfEverything).Hours() / 24))))
	_, isThere := stats.Values[id]
	if !isThere {
		stats.Values[id] = ShowAndMovie{
			Movie: make(map[string]string),
			Show:  make(map[string]string),
		}
	}
	if len(x.Movie.Title) > 0 {
		stats.Values[id].Movie[x.Movie.IDs.IMDB] = x.Movie.Title
	} else {
		stats.Values[id].Show[x.Episode.IDs.IMDB] = x.Show.Title + ": " + x.Episode.Title
	}
}

// Update the local cache of TraktStats
func updateTraktStats(stats TraktStats) TraktStats {
	page := 1
	limit := 20
	baseFrom := stats.LastUpdatedDate.Add(time.Hour * -24 * 15).Format(gmtDateFormat)
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	stats.LastUpdatedDate = startOfEverything
	stats.LastUpdated = stats.LastUpdatedDate.Format(gmtDateFormat)
	for {
		body := buildTraktLinkClient(baseFrom, page, limit)
		var parsed []TraktResponse

		err := json.Unmarshal(body, &parsed)
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		if len(parsed) == 0 {
			return stats
		}
		for _, x := range parsed {
			getMovieShowValues(x, &stats)
		}
		page = page + 1
	}
}

func getTraktStatsForDays(days int, stats TraktStats) (map[int]float64, map[int]float64, float64, float64) {
	movies := make(map[int]float64)
	shows := make(map[int]float64)
	max := 0
	min := 0
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	thisIndex := int(math.Ceil(time.Since(startOfEverything).Hours() / 24))
	endIndex := thisIndex - days
	i := days - 1
	for ; thisIndex > endIndex; thisIndex-- {
		entry, found := stats.Values[fmt.Sprintf("%d", thisIndex)]
		if found {
			mCount := len(entry.Movie)
			sCount := len(entry.Show)
			movies[i] = float64(mCount)
			shows[i] = float64(sCount)
			if max < mCount {
				max = mCount
			}
			if max < sCount {
				max = sCount
			}
		} else {
			movies[i] = 0
			shows[i] = 0
		}
		i--
	}
	return movies, shows, float64(max), float64(min)
}

func getDaysArray(known RSS) (map[int]float64, float64) {
	// Process last X days (based on config or by setting)
	today := time.Now()
	days := make(map[int]float64, ConfigData.BlogStats.Days)
	for i := 0; i < ConfigData.BlogStats.Days; i++ {
		days[i] = 0
	}
	for _, x := range known.Channel.Items {
		diff := int(math.Ceil(today.Sub(x.PubDateAsDate).Hours() / 24))
		if diff < ConfigData.BlogStats.Days {
			days[ConfigData.BlogStats.Days-1-diff]++
		} else {
			break
		}
	}
	max := days[0]
	for _, x := range days {
		if max < x {
			max = x
		}
	}

	return days, max
}

// CODESTATS
type CodeStatsResponse struct {
	Dates    map[string]int `json:"dates"`
	NewXPs   int            `json:"new_xp"`
	TotalXPs int            `json:"total_xp"`
	User     string         `json:"user"`
}

func generateCSStats() error {
	filenameOfSvg := ConfigData.BaseDir + "../regenerate/data/cs.svg"
	parsed, err := getObjectFromAPI()
	if err != nil {
		return err
	}
	// Get the last Days entries
	days, max := csToDays(parsed)
	// Make ze graph
	line, total := lineAlone(days, max, 0, map[string]string{"strokeDashArray": "", "stroke": colorBlackOpacity50}, "CodeStats", map[string]bool{"showZero": true, "showBall": false})
	chart := SVGGraphFromPaths(total, "CodeStats", -1, line)
	// Create the SVG
	err = ioutil.WriteFile(filenameOfSvg, chart, 0777)
	if err != nil {
		fmt.Printf("Failed to write %s:%v\n", filenameOfSvg, err)
		return err
	}
	return nil
}

func getObjectFromAPI() (CodeStatsResponse, error) {
	var parsed CodeStatsResponse
	url := "https://codestats.net/api/users/vonExplaino"
	request, _ := http.NewRequest(
		"GET",
		url,
		bytes.NewBuffer([]byte{}),
	)
	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set(jsonHeaders[1][0], jsonHeaders[1][1])

	resp, err := Client.Do(request)
	if err != nil {
		return parsed, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return parsed, err
	}
	if len(body) == 0 {
		return parsed, errors.New("failed to get contents from CodeStats")
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return parsed, err
	}
	return parsed, nil
}

func csToDays(parsed CodeStatsResponse) (map[int]float64, float64) {
	perDayStats := make(map[int]float64)
	dayDuration := time.Duration(-1) * time.Hour * 24
	thisDay := time.Now()
	max := 0.0
	for i := 1; i <= Days; i++ {
		value, ok := parsed.Dates[thisDay.Format("2006-01-02")]
		if ok {
			perDayStats[Days-i] = float64(value)
			if perDayStats[Days-i] > max {
				max = perDayStats[Days-i]
			}
		} else {
			perDayStats[Days-i] = 0
		}
		thisDay = thisDay.Add(dayDuration)
	}
	return perDayStats, max
}

// FEEDLY
type FeedlyItem struct {
	Fingerprint string   `json:"fingerprint"`
	Language    string   `json:"language"`
	ID          string   `json:"ID"`
	Keywords    []string `json:"keywords"`
	OriginID    string   `json:"originId"`
	Origin      struct {
		Title    string `json:"title"`
		StreamID string `json:"streamId"`
		HtmlURL  string `json:"htmlUrl"`
	} `json:"origin"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Unread    bool   `json:"unread"`
	Crawled   int64  `json:"crawled"`
	Published int64  `json:"published"`
}
type FeedlyResponse struct {
	ID           string       `json:"id"`
	Updated      int          `json:"updated"`
	Continuation string       `json:"continuation"`
	Items        []FeedlyItem `json:"items"`
}
type FeedlyStats struct {
	LastUpdated     string         `json:"last_updated"`
	LastSeen        string         `json:"last_seen"`
	Days            map[string]int `json:"days"`
	LastUpdatedDate time.Time
}

func generateFeedlyStats() {
	filenameOfFeedlySvg := ConfigData.BaseDir + relativeDataLocation + ConfigData.AboutMe.Feedly.Cache + ".svg"
	filenameOfFeedlyCache := ConfigData.BaseDir + relativeDataLocation + ConfigData.AboutMe.Feedly.Cache
	stats := readFeedlyStatsFile(filenameOfFeedlyCache)
	// Update cached info from source
	stats = updateFeedlyStats(stats)
	writeFeedlyStatsFile(filenameOfFeedlyCache, stats)
	// Get the line
	entries, max, min := getFeedlyStatsForDays(Days, stats)
	line1, total1 := lineAlone(entries, max, min, map[string]string{"strokeDashArray": "", "stroke": colorBlackOpacity50}, "Entries", map[string]bool{"showZero": true, "showBall": false})
	// Create and store the SVG
	graph := SVGGraphFromPaths(total1, fmt.Sprintf("%d", int(total1)), -1, line1)
	ioutil.WriteFile(filenameOfFeedlySvg, graph, 0666)

}

func readFeedlyStatsFile(filename string) FeedlyStats {
	// Get the cached info
	cacheFile, err := os.Open(filename)
	var buff FeedlyStats
	if err == nil {
		defer cacheFile.Close()
		byteValue, _ := ioutil.ReadAll(cacheFile)
		err := json.Unmarshal(byteValue, &buff)
		if err != nil {
			log.Fatalf("Failed to parse the Feedly cache %v\n", err)
		}
	} else {
		buff.LastSeen = ""
		buff.Days = make(map[string]int)
	}
	if len(buff.LastUpdated) > 0 {
		buff.LastUpdatedDate, _ = parseUnknownDateFormat(buff.LastUpdated)
	}
	return buff
}

func writeFeedlyStatsFile(filename string, stats FeedlyStats) error {
	marshalled, err := json.Marshal(stats)
	if err == nil {
		ioutil.WriteFile(filename, marshalled, 0666)
	}
	return err
}

func bodyFromFeedlyLinkClient(continuation string) ([]byte, error) {
	url := fmt.Sprintf("%s&continuation=%s",
		ConfigData.AboutMe.Feedly.URL,
		continuation,
	)
	request, _ := http.NewRequest(
		"GET",
		url,
		bytes.NewBuffer([]byte{}),
	)

	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set(jsonHeaders[1][0], jsonHeaders[1][1])
	request.Header.Set("Authorization", "OAuth "+ConfigData.AboutMe.Feedly.Key)

	resp, err := Client.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	if len(body) == 0 {
		log.Fatal("Failed to get contents from Feedly\n")
	}
	return body, nil
}
func processFeedlyStats(items []FeedlyItem, stats *FeedlyStats, newLastSeen string) (int64, bool) {
	var lastPublished int64
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	for _, x := range items {
		if x.Fingerprint == stats.LastSeen {
			// We've already recorded this one! Break
			stats.LastSeen = newLastSeen
			return lastPublished, false
		}
		publishedDate := fmt.Sprintf(
			"%.0f",
			time.Unix(
				0,
				x.Published*int64(time.Millisecond),
			).Sub(startOfEverything).Hours()/24)
		_, already := stats.Days[publishedDate]
		if already {
			stats.Days[publishedDate]++
		} else {
			stats.Days[publishedDate] = 1
		}
		lastPublished = x.Published
	}
	return lastPublished, true
}
func updateFeedlyStats(stats FeedlyStats) FeedlyStats {
	stats.LastUpdatedDate = time.Now()
	stats.LastUpdated = stats.LastUpdatedDate.Format(gmtDateFormat)
	newLastSeen := ""
	continuation := ""
	page := 0
	thirtyDaysAgo := stats.LastUpdatedDate.Add(time.Duration(-30*24) * time.Hour).UnixMilli()
	for {
		body, _ := bodyFromFeedlyLinkClient(continuation)
		var parsed FeedlyResponse
		err := json.Unmarshal(body, &parsed)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if len(parsed.Items) == 0 {
			return stats
		}
		if len(stats.LastSeen) == 0 && len(newLastSeen) == 0 {
			newLastSeen = parsed.Items[0].Fingerprint
		}
		lastPublished, continueBool := processFeedlyStats(parsed.Items, &stats, newLastSeen)
		continuation = parsed.Continuation
		if len(continuation) == 0 || thirtyDaysAgo > lastPublished || !continueBool {
			stats.LastSeen = newLastSeen
			return stats
		}
		page++
		if page > 140 {
			os.Exit(5)
		}
	}
}

func getFeedlyStatsForDays(days int, stats FeedlyStats) (map[int]float64, float64, float64) {
	entries := make(map[int]float64)
	max := 0
	min := 0
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	thisIndex := int(math.Ceil(time.Since(startOfEverything).Hours()/24)) - 1
	endIndex := thisIndex - days
	i := 0
	for ; thisIndex > endIndex; thisIndex-- {
		entry, found := stats.Days[fmt.Sprintf("%d", thisIndex)]
		if found {
			entries[i] = float64(entry)
			if max < entry {
				max = entry
			}
		} else {
			entries[i] = 0
		}
		i++
	}
	entries2 := make(map[int]float64)
	for i = 0; i < len(entries); i++ {
		entries2[Days-i-1] = entries[i]
	}
	return entries2, float64(max), float64(min)

}

// WITHINGS
type WithingsOauthResponse struct {
	Body struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
		UserID       int64  `json:"userid"`
	} `json:"body"`
	Status int `json:"status"`
}
type WithingsResponse1 struct {
	Status int `json:"status"`
	Body   struct {
		UpdateTime  int64  `json:"updatetime"`
		TimeZone    string `json:"timezone"`
		MeasureGrps []struct {
			GroupID      string `json:"grpid"`
			Attrib       int    `json:"attrib"`
			Date         int64  `json:"date"`
			Created      int64  `json:"created"`
			Category     int    `json:"category"`
			DeviceID     string `json:"deviceid"`
			HashDeviceID string `json:"hash_deviceid"`
			Measures     []struct {
				Value    int `json:"value"`
				Type     int `json:"type"`
				Unit     int `json:"unit"`
				Algo     int `json:"algo"`
				FM       int `json:"fm"`
				Apppfmid int `json:"apppfmid"`
				AppLiver int `json:"appliver"`
			} `json:"measures"`
			Comment string `json:"comment"`
		}
	} `json:"body"`
}
type WithingsResponse2 struct {
	Status int `json:"status"`
	Body   struct {
		Activities []struct {
			Steps         int     `json:"steps"`
			Distance      float64 `json:"distance"`
			TotalCalories float64 `json:"totalcalories"`
			Date          string  `json:"date"`
		} `json:"activities"`
		More   bool `json:"more"`
		Offset int  `json:"offset"`
	}
}
type WithingsMeasure struct {
	Kg            float64 `json:"kg"`
	Steps         int     `json:"steps"`
	Distance      float64 `json:"distance"`
	TotalCalories float64 `json:"totalcalories"`
}
type WithingsStats struct {
	LastUpdated     string `json:"last_updated"`
	LastUpdatedDate time.Time
	Values          map[string]WithingsMeasure `json:"values"`
}

func generateWithingsStats() {
	filenameOfWithingsCache := ConfigData.BaseDir + relativeDataLocation + ConfigData.AboutMe.Withings.Cache
	filenameOfWithingsWeightSvg := ConfigData.BaseDir + relativeDataLocation + ConfigData.AboutMe.Withings.Cache + "-weight.svg"
	filenameOfWithingsStepsSvg := ConfigData.BaseDir + relativeDataLocation + ConfigData.AboutMe.Withings.Cache + "-steps.svg"
	stats := readWithingsStats(filenameOfWithingsCache)
	// Update cached info from source
	stats = updateWithingsStats(stats)
	writeWithingsStats(filenameOfWithingsCache, stats)
	// Get the line
	weight, steps, wMax, _, wDiff, sMax, _ := getWithingsStatsForDays(Days, stats)
	graph1 := barSVG(weight, 103.0, 95.0, -1)
	line1, total1 := lineAlone(weight, 103.0, 95.0, map[string]string{"strokeDashArray": "", "stroke": colorBlackOpacity50}, "Kgs", map[string]bool{"showZero": false, "showBall": true})
	graph1p5 := SVGGraphFromPaths(total1, fmt.Sprintf("%f", wMax), wDiff, line1)
	graph1 = []byte(strings.Replace(string(graph1), "</svg>", fmt.Sprintf("%s</svg>", graph1p5), -1))
	line2, total2 := lineAlone(steps, sMax, 0, map[string]string{"strokeDashArray": "", "stroke": colorBlackOpacity50}, "Steps", map[string]bool{"showZero": false, "showBall": false})
	graph2 := SVGGraphFromPaths(total2, fmt.Sprintf("%d", int(total2)), -1, line2)
	// Store the SVG
	ioutil.WriteFile(filenameOfWithingsWeightSvg, graph1, 0666)
	ioutil.WriteFile(filenameOfWithingsStepsSvg, graph2, 0666)
}

func readWithingsStats(filename string) WithingsStats {
	// Get the cached info
	cacheFile, err := os.Open(filename)
	var buff WithingsStats
	if err == nil {
		defer cacheFile.Close()
		byteValue, _ := ioutil.ReadAll(cacheFile)
		err := json.Unmarshal(byteValue, &buff)
		if err != nil {
			log.Fatalf("Failed to parse the Withings cache %v\n", err)
		}
	}
	if len(buff.LastUpdated) > 0 {
		buff.LastUpdatedDate, _ = parseUnknownDateFormat(buff.LastUpdated)
	}
	return buff
}
func writeWithingsStats(filename string, stats WithingsStats) error {
	marshalled, err := json.Marshal(stats)
	if err == nil {
		ioutil.WriteFile(filename, marshalled, 0666)
	}
	return err
}

func updateWithingsTokenIfRequired(accessToken string) (string, error) {
	now := time.Now().Unix()
	var err error
	if len(accessToken) == 0 || now > int64(ConfigData.AboutMe.Withings.ExpiresAt) {
		refreshToken := ConfigData.AboutMe.Withings.RefreshToken
		if len(refreshToken) > 0 {
			data := url.Values{
				"action":        {"requesttoken"},
				"client_id":     {ConfigData.AboutMe.Withings.Client},
				"client_secret": {ConfigData.AboutMe.Withings.Secret},
				"grant_type":    {"refresh_token"},
				"refresh_token": {ConfigData.AboutMe.Withings.RefreshToken},
			}
			req, _ := http.NewRequest(
				"POST",
				ConfigData.AboutMe.Withings.OauthURL,
				strings.NewReader(data.Encode()))
			resp, err := Client.Do(req)
			if err != nil {
				return "", errors.New("could not refresh the token")
			}
			var res WithingsOauthResponse
			json.NewDecoder(resp.Body).Decode(&res)
			accessToken := res.Body.AccessToken
			if len(accessToken) == 0 {
				return "", errors.New("could not parse the token refresh response from withings")
			}
			ConfigData.AboutMe.Withings.AccessToken = accessToken
			ConfigData.AboutMe.Withings.RefreshToken = res.Body.RefreshToken
			ConfigData.AboutMe.Withings.ExpiresAt = time.Now().Unix() + res.Body.ExpiresIn
			viper.Set("aboutme.withings.accesstoken", accessToken)
			viper.Set("aboutme.withings.refreshtoken", res.Body.RefreshToken)
			viper.Set("aboutme.withings.expiresat", time.Now().Unix()+res.Body.ExpiresIn)
			viper.WriteConfig()
		} else {
			return "", errors.New("no access or refresh token available")
		}
	}
	return accessToken, err
}

func updateWithingsStatvalues(stats *WithingsStats, res WithingsResponse1, startOfEverything time.Time) {
	l, _ := time.LoadLocation(blogTimezone)
	for _, group := range res.Body.MeasureGrps {
		watchedAt := time.Unix(group.Date, 0)
		watchedAt = time.Date(watchedAt.Year(), watchedAt.Month(), watchedAt.Day(), watchedAt.Hour(), watchedAt.Minute(), watchedAt.Second(), watchedAt.Nanosecond(), l)
		entryID := fmt.Sprintf("%d", (int(math.Ceil(watchedAt.Sub(startOfEverything).Hours() / 24))))
		_, ok := stats.Values[entryID]
		if !ok {
			stats.Values[entryID] = WithingsMeasure{
				Kg:            0,
				Steps:         0,
				Distance:      0,
				TotalCalories: 0,
			}
		}
		for _, measure := range group.Measures {
			stats.Values[entryID] = WithingsMeasure{
				Kg:            float64(measure.Value) * math.Pow(10, float64(measure.Unit)),
				Steps:         stats.Values[entryID].Steps,
				Distance:      stats.Values[entryID].Distance,
				TotalCalories: stats.Values[entryID].TotalCalories,
			}
		}
	}
}
func updateWithingsStats(stats WithingsStats) WithingsStats {
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	// Refresh token if needed
	accessToken, err := updateWithingsTokenIfRequired(ConfigData.AboutMe.Withings.AccessToken)
	if err != nil {
		log.Fatalf("Failed to refresh withings token")
	}
	lastUpdate := stats.LastUpdatedDate.Unix()
	lastUpdateString := fmt.Sprintf("%d", lastUpdate-60*60*24*30)
	stats.LastUpdatedDate = time.Now()
	stats.LastUpdated = stats.LastUpdatedDate.Format("2006-01-02 15:04:05 -0700 MST")

	// Mass
	data := url.Values{
		"action":     {"getmeas"},
		"meastype":   {"1"},
		"category":   {"1"},
		"lastupdate": {lastUpdateString},
	}
	request, _ := http.NewRequest(
		"POST",
		ConfigData.AboutMe.Withings.MassURL,
		bytes.NewBuffer([]byte(data.Encode())),
	)
	request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")

	resp, err := Client.Do(request)
	if err != nil {
		log.Fatalf("Failed to get weight")
	}
	var res WithingsResponse1
	json.NewDecoder(resp.Body).Decode(&res)
	if res.Status == 0 {
		updateWithingsStatvalues(&stats, res, startOfEverything)
	} else {
		PrintIfNotSilent(fmt.Sprintf("Failed to parse withings response1 %d", res.Status))
	}
	// Steps
	offset := 0
	for {
		data = url.Values{
			"action":      {"getactivity"},
			"data_fields": {"steps,distance,totalcalories"},
			"lastupdate":  {lastUpdateString},
			"offset":      {fmt.Sprintf("%d", offset)},
		}
		request, _ = http.NewRequest("POST", ConfigData.AboutMe.Withings.StepsURL, bytes.NewBuffer([]byte(data.Encode())))
		request.Header.Set(jsonHeaders[0][0], jsonHeaders[0][1])
		request.Header.Set("Content-type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := Client.Do(request)
		if err != nil {
			PrintIfNotSilent("Failed to get steps")
		}
		var res WithingsResponse2
		json.NewDecoder(resp.Body).Decode(&res)
		for _, activity := range res.Body.Activities {
			watchedAt, _ := time.Parse("2006-01-02", activity.Date)
			watchedAt = time.Date(watchedAt.Year(), watchedAt.Month(), watchedAt.Day(), watchedAt.Hour(), watchedAt.Minute(), watchedAt.Second(), watchedAt.Nanosecond(), l)
			entryID := fmt.Sprintf("%d", (int(math.Ceil(watchedAt.Sub(startOfEverything).Hours() / 24))))
			_, ok := stats.Values[entryID]
			if !ok {
				stats.Values[entryID] = WithingsMeasure{
					Kg:            0,
					Steps:         0,
					Distance:      0,
					TotalCalories: 0,
				}
			} else {
				stats.Values[entryID] = WithingsMeasure{
					Kg:            stats.Values[entryID].Kg,
					Steps:         activity.Steps,
					Distance:      activity.Distance,
					TotalCalories: activity.TotalCalories,
				}
			}
		}
		if !res.Body.More {
			break
		}
		offset = res.Body.Offset
	}

	return stats
}

func getWithingsStatsForDays(days int, stats WithingsStats) (map[int]float64, map[int]float64, float64, float64, float64, float64, float64) {
	weight := make(map[int]float64)
	steps := make(map[int]float64)
	weightMax := 0.0
	weightMin := 0.0
	stepsMax := 0.0
	stepsMin := 0.0
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	thisIndex := int(math.Ceil(time.Since(startOfEverything).Hours()/24)) - 1
	endIndex := thisIndex - days
	i := 0
	weightFirst := -1.0
	weightLast := -1.0
	for ; thisIndex > endIndex; thisIndex-- {
		entry, found := stats.Values[fmt.Sprintf("%d", thisIndex)]
		if found {
			weight[i] = float64(entry.Kg)
			if weightFirst == -1.0 && weight[i] > 0 {
				weightFirst = weight[i]
			}
			if weight[i] > 0 {
				weightLast = weight[i]
			}
			steps[i] = float64(entry.Steps)
			weightMax = math.Max(weightMax, weight[i])
			weightMin = math.Min(weightMin, weight[i])
			stepsMax = math.Max(stepsMax, steps[i])
			stepsMin = math.Min(stepsMin, steps[i])
		} else {
			weight[i] = 0.0
			steps[i] = 0.0
		}
		i++
	}
	weight2 := make(map[int]float64)
	for i = 0; i < len(weight); i++ {
		weight2[days-i-1] = weight[i]
	}
	steps2 := make(map[int]float64)
	for i = 0; i < len(steps); i++ {
		steps2[days-i-1] = steps[i]
	}
	return weight2, steps2, weightMax, weightMin, weightLast - weightFirst, stepsMax, stepsMin
}

//GENERIC FUNCTIONS
func barSVG(days map[int]float64, max float64, min float64, total float64) []byte {
	chartHeight := 16.0
	chartHeightStep := 0.0
	if max > 0 {
		chartHeightStep = chartHeight / (max - min)
	}
	chartWidth := 100.0
	chartWidthStep := chartWidth / float64(len(days))
	color := "rgba(0,0,0,0.5)"
	title := "Blog Posts"
	x := 0.0
	y := 0.0
	line := fmt.Sprintf(`<path fill="%s" stroke="%s" stroke-width="1" d="`, color, color)
	totalPosts := 0.0

	statFirst := -1.0
	statLast := -1.0
	for i := 0; i < len(days); i++ {
		postcount := days[i]
		totalPosts += postcount
		y = (float64(postcount) - min) * chartHeightStep
		if statFirst == -1.0 && float64(postcount) > 0 {
			statFirst = float64(postcount)
		}
		if float64(postcount) > 0 {
			statLast = float64(postcount)
		}
		x2 := x + chartWidthStep
		line = line + fmt.Sprintf(
			`M%f,%f L%f,%f L%f,%f L%f,%f Z`,
			x,
			chartHeight,
			x,
			chartHeight-y,
			x2,
			chartHeight-y,
			x2,
			chartHeight,
		)
		x = x2
	}
	line += `"></path>`
	if total > -1 {
		totalPosts = total
	}

	return []byte(fmt.Sprintf(`<svg version="1.1" baseProfile="full" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:ev="http://www.w3.org/2001/xml-events" width="100" height="16" data-total="%d" data-diff="%.2f" data-title="%s">%s</svg>`, int(totalPosts), statFirst-statLast, title, line))
}

func lineAlone(
	points map[int]float64,
	max float64,
	min float64,
	strokeOptions map[string]string,
	//strokeDashArray string,
	//stroke string,
	title string,
	options map[string]bool,
	//showZero bool,
	//showBall bool,
) (string, float64) {

	// Defaults
	chartHeight := 16.0
	chartHeightStep := 0.0
	chartHeightStep = chartHeight / (max - min)
	chartWidth := 100.0
	chartWidthStep := chartWidth / float64(len(points))
	strokeDashArray := strokeOptions["strokeDashArray"]
	stroke := strokeOptions["stroke"]
	showZero := options["showZero"]
	showBall := options["showBall"]
	if strokeDashArray == "" {
		strokeDashArray = "0"
	}
	if stroke == "" {
		stroke = colorBlackOpacity50
	}
	x := 0.0
	y := chartHeight - (points[0]-min)*chartHeightStep

	// Build
	line := fmt.Sprintf(
		`<path fill="none" stroke="%s" stroke-width="1" stroke-dasharray="%s" d="M%f,%f`,
		stroke,
		strokeDashArray,
		x,
		y)
	total := points[0]

	lastEntry := y
	extraPath := ""
	radius := 0.5
	for i := 1; i < len(points); i++ {
		entrycount := points[i]
		total += entrycount
		x += chartWidthStep
		if entrycount == 0 {
			y = chartHeight
		} else {
			y = chartHeight - (entrycount-min)*chartHeightStep
		}
		moveOrLine := "L"
		if !showZero &&
			(lastEntry >= chartHeight || y >= chartHeight) {
			moveOrLine = "M"
		}
		line = line + fmt.Sprintf(
			`%s%f,%f `,
			moveOrLine,
			x,
			y,
		)
		if showBall && !showZero && y < chartHeight {
			extraPath += fmt.Sprintf(
				" M%f,%f a %f,%f 0 1,0 %f,0 a %f,%f, 0 1,0 %f,0",
				x-radius,
				y,
				radius,
				radius,
				radius*2,
				radius,
				radius,
				radius*-2,
			)
		}
		lastEntry = y
	}
	line += extraPath + `"></path>`

	return line, total
}

func SVGGraphFromPaths(total float64, title string, diff float64, line string) []byte {
	strTotal := fmt.Sprintf("%f", total)
	if math.Ceil(total) == total {
		strTotal = fmt.Sprintf("%d", int(math.Ceil(total)))
	}
	diffString := ""
	if diff > -1 {
		diffString = fmt.Sprintf(`data-diff="%.2f" `, diff)
	}
	return []byte(fmt.Sprintf(`<svg version="1.1" baseProfile="full" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:ev="http://www.w3.org/2001/xml-events" width="100" height="16" data-total="%s" %sdata-title="%s">%s</svg>`, strTotal, diffString, title, line))
}

func init() {
	rootCmd.AddCommand(blogstatsCmd)
	blogstatsCmd.Flags().IntVarP(&Days, "days", "d", 0, "How many days to generate the graph for")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.All, "all", "a", false, "Create all svgs/ caches")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.Codestats, "codestats", "c", false, "Create codestats")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.Trakt, "trakt", "t", false, "Create Trakt")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.Blog, "blog", "b", false, "Create blog")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.Feedly, "feedly", "f", false, "Create Feedly")
	blogstatsCmd.Flags().BoolVarP(&SVGOptions.Withings, "withings", "w", false, "Create Withings")
}
