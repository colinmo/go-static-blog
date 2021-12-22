package cmd

import (
	"fmt"
	"io/ioutil"
	"math"
	"testing"
	"time"
)

// TestReadRSS test parses some RSS Feeds
func TestBarChart(t *testing.T) {
	Days = 30
	expectedChart0To30 := `<svg version="1.1" baseProfile="full" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:ev="http://www.w3.org/2001/xml-events" width="100" height="16" data-total="435" data-title="Blog Posts"><path fill="rgba(0,0,0,0.5)" stroke="rgba(0,0,0,0.5)" stroke-width="1" d="M0.000000,16.000000 L0.000000,16.000000 L3.333333,16.000000 L3.333333,16.000000 ZM3.333333,16.000000 L3.333333,15.466667 L6.666667,15.466667 L6.666667,16.000000 ZM6.666667,16.000000 L6.666667,14.933333 L10.000000,14.933333 L10.000000,16.000000 ZM10.000000,16.000000 L10.000000,14.400000 L13.333333,14.400000 L13.333333,16.000000 ZM13.333333,16.000000 L13.333333,13.866667 L16.666667,13.866667 L16.666667,16.000000 ZM16.666667,16.000000 L16.666667,13.333333 L20.000000,13.333333 L20.000000,16.000000 ZM20.000000,16.000000 L20.000000,12.800000 L23.333333,12.800000 L23.333333,16.000000 ZM23.333333,16.000000 L23.333333,12.266667 L26.666667,12.266667 L26.666667,16.000000 ZM26.666667,16.000000 L26.666667,11.733333 L30.000000,11.733333 L30.000000,16.000000 ZM30.000000,16.000000 L30.000000,11.200000 L33.333333,11.200000 L33.333333,16.000000 ZM33.333333,16.000000 L33.333333,10.666667 L36.666667,10.666667 L36.666667,16.000000 ZM36.666667,16.000000 L36.666667,10.133333 L40.000000,10.133333 L40.000000,16.000000 ZM40.000000,16.000000 L40.000000,9.600000 L43.333333,9.600000 L43.333333,16.000000 ZM43.333333,16.000000 L43.333333,9.066667 L46.666667,9.066667 L46.666667,16.000000 ZM46.666667,16.000000 L46.666667,8.533333 L50.000000,8.533333 L50.000000,16.000000 ZM50.000000,16.000000 L50.000000,8.000000 L53.333333,8.000000 L53.333333,16.000000 ZM53.333333,16.000000 L53.333333,7.466667 L56.666667,7.466667 L56.666667,16.000000 ZM56.666667,16.000000 L56.666667,6.933333 L60.000000,6.933333 L60.000000,16.000000 ZM60.000000,16.000000 L60.000000,6.400000 L63.333333,6.400000 L63.333333,16.000000 ZM63.333333,16.000000 L63.333333,5.866667 L66.666667,5.866667 L66.666667,16.000000 ZM66.666667,16.000000 L66.666667,5.333333 L70.000000,5.333333 L70.000000,16.000000 ZM70.000000,16.000000 L70.000000,4.800000 L73.333333,4.800000 L73.333333,16.000000 ZM73.333333,16.000000 L73.333333,4.266667 L76.666667,4.266667 L76.666667,16.000000 ZM76.666667,16.000000 L76.666667,3.733333 L80.000000,3.733333 L80.000000,16.000000 ZM80.000000,16.000000 L80.000000,3.200000 L83.333333,3.200000 L83.333333,16.000000 ZM83.333333,16.000000 L83.333333,2.666667 L86.666667,2.666667 L86.666667,16.000000 ZM86.666667,16.000000 L86.666667,2.133333 L90.000000,2.133333 L90.000000,16.000000 ZM90.000000,16.000000 L90.000000,1.600000 L93.333333,1.600000 L93.333333,16.000000 ZM93.333333,16.000000 L93.333333,1.066667 L96.666667,1.066667 L96.666667,16.000000 ZM96.666667,16.000000 L96.666667,0.533333 L100.000000,0.533333 L100.000000,16.000000 Z"></path></svg>`
	ddays := make(map[int]float64, Days)
	for i := 0; i < Days; i++ {
		ddays[i] = float64(i)
	}
	mep := barSVG(ddays, float64(Days), 0, 435)
	if string(mep) != expectedChart0To30 {
		t.Fatalf(`Bad bar chart %s`, mep)
	}
}

func TestRssToArray(t *testing.T) {
	ConfigData.BlogStats.Days = 30
	// Read the XML file
	known, err := ReadRSS(`f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\rss_published.xml`)
	if err != nil {
		// Empty
		known = RSS{}
	}
	days, max := getDaysArray(known)
	if max == 0 {
		t.Errorf("Well we got %f max\n", max)
	}
	if len(days) == 0 {
		t.Errorf("No data")
	}
}

func TestReadTrakt(t *testing.T) {
	mep := readTraktStats(`f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\trakt-small.json`)

	if mep.LastUpdated != "2021-11-06T14:00:00Z" {
		t.Errorf("Last Updated misparsed %s\n", mep.LastUpdated)
	}
	if mep.LastUpdatedDate.String() != "2021-11-07 00:00:00 +1000 AEST" {
		t.Errorf("Didn't parse date %s\n", mep.LastUpdatedDate.String())
	}
	if mep.Values["18496"].Show["6555865217"] != "The Umbrella Academy Right Back Where We Started" {
		t.Errorf("Got the wrong title for entry 18496:%s\n", mep.Values["18496"].Show["6555865217"])
	}
}

func TestGetDataFromTrakt(t *testing.T) {
	ConfigData.AboutMe.Trakt = Trakt{
		ID:           "012a55d63341d7781478c392cc868f7bc6d55de72765a860ad3c61d6dbe8e2ee",
		Secret:       "52182d6590d7c3db27659d28fae4bcc0723aa6fe275ab313ffc84b3326bb79d9",
		AccessToken:  "29f6133cdfe9f33d782b4e0c72c69ca541b5713e48f7b5e8e1c08b38f30efc4f",
		RefreshToken: "6f76d9298731c709b264fb8f74c6d2d57d9fff3b53faf5cb9eff07b341834346",
		Cache:        "trakt-cache.json",
	}
	mep := TraktStats{
		LastUpdated: "2021-10-06T14:00:00Z",
	}
	mep.LastUpdatedDate, _ = parseUnknownDateFormat(mep.LastUpdated)
	lastLook := mep.LastUpdatedDate
	mep.Values = make(map[string]ShowAndMovie)
	mep = updateTraktStats(mep)
	if mep.LastUpdatedDate == lastLook {
		t.Errorf("Failed to update date of process\n")
	}
	inner, ok := mep.Values["18909"]
	if !ok {
		t.Errorf("Failed to find 18909\n")
	}
	if inner.Show["tt4052886"] != "Lucifer: BlueBallz" {
		t.Errorf("Title was wrong %s\n", inner.Show["tt4052886"])
	}
}

func TestLineChart(t *testing.T) {
	expectedChart0To30 := `<svg version="1.1" baseProfile="full" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:ev="http://www.w3.org/2001/xml-events" width="100" height="16" data-total="435" data-title="Title Test"><path fill="none" stroke="green" stroke-width="1" stroke-dasharray="0" d="M0.000000,16.000000L3.333333,15.542857 L6.666667,15.085714 L10.000000,14.628571 L13.333333,14.171429 L16.666667,13.714286 L20.000000,13.257143 L23.333333,12.800000 L26.666667,12.342857 L30.000000,11.885714 L33.333333,11.428571 L36.666667,10.971429 L40.000000,10.514286 L43.333333,10.057143 L46.666667,9.600000 L50.000000,9.142857 L53.333333,8.685714 L56.666667,8.228571 L60.000000,7.771429 L63.333333,7.314286 L66.666667,6.857143 L70.000000,6.400000 L73.333333,5.942857 L76.666667,5.485714 L80.000000,5.028571 L83.333333,4.571429 L86.666667,4.114286 L90.000000,3.657143 L93.333333,3.200000 L96.666667,2.742857 "></path></svg>`
	days := make(map[int]float64, 30)
	for i := 0.0; i < 30.0; i += 1.0 {
		days[int(i)] = i
	}
	line, count := lineAlone(
		days,
		35,
		0,
		"",
		"green",
		"Title",
		true,
		false,
	)
	mep := SVGGraphFromPaths(
		count,
		"Title Test",
		-1,
		line,
	)
	if string(mep) != expectedChart0To30 {
		t.Fatalf(`Bad bar chart %s`, mep)
	}
}

func TestSVGGraphFromPaths(t *testing.T) {
	expectedChart0To30 := `<svg version="1.1" baseProfile="full" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:ev="http://www.w3.org/2001/xml-events" width="100" height="16" data-total="870" data-title="Title Test"><path fill="none" stroke="green" stroke-width="1" stroke-dasharray="0" d="M0.000000,16.000000L3.333333,15.542857 L6.666667,15.085714 L10.000000,14.628571 L13.333333,14.171429 L16.666667,13.714286 L20.000000,13.257143 L23.333333,12.800000 L26.666667,12.342857 L30.000000,11.885714 L33.333333,11.428571 L36.666667,10.971429 L40.000000,10.514286 L43.333333,10.057143 L46.666667,9.600000 L50.000000,9.142857 L53.333333,8.685714 L56.666667,8.228571 L60.000000,7.771429 L63.333333,7.314286 L66.666667,6.857143 L70.000000,6.400000 L73.333333,5.942857 L76.666667,5.485714 L80.000000,5.028571 L83.333333,4.571429 L86.666667,4.114286 L90.000000,3.657143 L93.333333,3.200000 L96.666667,2.742857 "></path> <path fill="none" stroke="blue" stroke-width="1" stroke-dasharray="1" d="M0.000000,2.285714L3.333333,2.742857 L6.666667,3.200000 L10.000000,3.657143 L13.333333,4.114286 L16.666667,4.571429 L20.000000,5.028571 L23.333333,5.485714 L26.666667,5.942857 L30.000000,6.400000 L33.333333,6.857143 L36.666667,7.314286 L40.000000,7.771429 L43.333333,8.228571 L46.666667,8.685714 L50.000000,9.142857 L53.333333,9.600000 L56.666667,10.057143 L60.000000,10.514286 L63.333333,10.971429 L66.666667,11.428571 L70.000000,11.885714 L73.333333,12.342857 L76.666667,12.800000 L80.000000,13.257143 L83.333333,13.714286 L86.666667,14.171429 L90.000000,14.628571 L93.333333,15.085714 L96.666667,15.542857 "></path></svg>`

	days := make(map[int]float64, 30)
	for i := 0.0; i < 30.0; i += 1.0 {
		days[int(i)] = i
	}
	line1, count1 := lineAlone(
		days,
		35,
		0,
		"",
		"green",
		"Title",
		true, false,
	)
	days2 := make(map[int]float64, 30)
	for i := 0.0; i < 30.0; i += 1.0 {
		days2[int(i)] = 30.0 - i
	}
	line2, count2 := lineAlone(
		days2,
		35,
		0,
		"1",
		"blue",
		"Title",
		true, false,
	)
	mep := SVGGraphFromPaths(
		count1+count2,
		"Title Test", -1,
		line1+" "+line2,
	)
	if string(mep) != expectedChart0To30 {
		t.Fatalf(`Bad bar chart %s`, mep)
	}
}

func TestGetTraktStatsForDays(t *testing.T) {
	mep := TraktStats{
		LastUpdated: "2021-10-10 13:13:13 AEST",
		Values:      make(map[string]ShowAndMovie),
	}
	l, _ := time.LoadLocation("Australia/Brisbane")
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	startIndex := int(math.Ceil(time.Since(startOfEverything).Hours() / 24))
	//	mep.Values[]
	index := fmt.Sprintf("%d", startIndex)
	mep.Values[index] = ShowAndMovie{
		Show:  make(map[string]string),
		Movie: make(map[string]string),
	}
	mep.Values[index].Show["x"] = "Mep"
	mep.Values[index].Show["y"] = "Mep"
	mep.Values[index].Show["z"] = "Mep"

	index = fmt.Sprintf("%d", startIndex-3)
	mep.Values[index] = ShowAndMovie{
		Show:  make(map[string]string),
		Movie: make(map[string]string),
	}
	mep.Values[index].Movie["x"] = "Mep"
	mep.Values[index].Movie["y"] = "Mep"
	mep.Values[index].Show["z"] = "Mep"

	index = fmt.Sprintf("%d", startIndex-10)
	mep.Values[index] = ShowAndMovie{
		Show:  make(map[string]string),
		Movie: make(map[string]string),
	}
	mep.Values[index].Movie["x"] = "Mep"
	mep.Values[index].Show["y"] = "Mep"
	mep.Values[index].Movie["z"] = "Mep"
	mep.Values[index].Movie["xx"] = "Mep"
	mep.Values[index].Show["xy"] = "Mep"
	mep.Values[index].Movie["xz"] = "Mep"

	index = fmt.Sprintf("%d", startIndex-50)
	mep.Values[index] = ShowAndMovie{
		Show:  make(map[string]string),
		Movie: make(map[string]string),
	}
	mep.Values[index].Movie["x"] = "Mep"
	mep.Values[index].Show["y"] = "Mep"
	mep.Values[index].Movie["z"] = "Mep"
	mep.Values[index].Movie["xx"] = "Mep"
	mep.Values[index].Show["xy"] = "Mep"
	mep.Values[index].Movie["xz"] = "Mep"

	movies, shows, max, min := getTraktStatsForDays(30, mep)
	if max != 4 {
		t.Fatalf("Max wrong %f\n", max)
	}
	if min != 0 {
		t.Fatalf("Min wrong %f\n", min)
	}
	if len(movies) != 30 {
		t.Fatalf("Wrong number of days for movies %d\n", len(movies))
	}
	if len(shows) != 30 {
		t.Fatalf("Wrong number of days for shows %d\n", len(shows))
	}
	if movies[3] != 2 {
		t.Fatalf("Movies 3 has wrong number %f\n", movies[3])
	}
}

func TestTraktChart(t *testing.T) {
	movies := make(map[int]float64)
	shows := make(map[int]float64)
	for i := 0; i < 30; i++ {
		movies[i] = 0
		shows[i] = 0
	}
	movies[3] = 2
	movies[10] = 4
	shows[0] = 3
	shows[3] = 1
	shows[9] = 3

	line1, total1 := lineAlone(shows, 4, 0, "", "green", "Shows", true, false)
	line2, total2 := lineAlone(movies, 4, 0, "", "blue", "Movies", true, false)
	// Create and store the SVG
	graph := SVGGraphFromPaths(total1+total2, "Trakt stats", -1, line1+line2)
	if len(graph) == 0 {
		t.Fatalf("Nuts")
	}
}

func TestDownloadCS(t *testing.T) {
	parsed := getObjectFromAPI()
	if len(parsed.Dates) == 0 {
		t.Fatalf("Got nothing")
	}
	if parsed.Dates["2020-03-11"] != 219 {
		t.Fatalf("Got bad")
	}
}
func TestCSChart(t *testing.T) {
	Days = 30
	testDays := CodeStatsResponse{
		Dates: make(map[string]int),
	}
	today := time.Now()
	dayDuration := time.Duration(-1) * time.Hour * 24
	testDays.Dates[today.Format("2006-01-02")] = 219
	testDays.Dates[today.Add(dayDuration*3).Format("2006-01-02")] = 123123

	days, max := csToDays(testDays)
	fmt.Printf("%v\n", days)
	if days[29] != 219.0 {
		t.Fatalf("First entry failed %f\n", days[0])
	}
	if days[26] != 123123.0 {
		t.Fatalf("Fourth entry failed %f\n", days[3])
	}
	line, total := lineAlone(days, max, 0, "", "rgb(0,0,0,0.5)", "CodeStats", true, false)
	chart := SVGGraphFromPaths(total, "CodeStats", -1, line)
	if string(chart[0:1]) != "<" {
		t.Fatalf("Well it's stuffed")
	}
	//t.Fatalf("%s\n", chart)
}

func TestReadFeedlyStats(t *testing.T) {
	stats := readFeedlyStats(`f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\feedly.cache.json`)
	if len(stats.Days) == 0 {
		t.Fatalf("Failed to read the Feedly stats")
	}
	if stats.LastUpdated != "2021-12-11T14:01:01Z" {
		t.Fatalf("Last updated date wrong")
	}
}

func TestGetFeedlyStatsForDays(t *testing.T) {
	Days = 30
	stats := FeedlyStats{
		LastUpdated: "2021-12-11T14:01:01Z",
		Days:        make(map[string]int),
	}
	l, _ := time.LoadLocation("Australia/Brisbane")
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	startIndex := int(math.Ceil(time.Since(startOfEverything).Hours()/24)) - 1
	//	mep.Values[]
	index := fmt.Sprintf("%d", startIndex)
	stats.Days[index] = 44
	index = fmt.Sprintf("%d", startIndex-3)
	stats.Days[index] = 452

	entries, max, min := getFeedlyStatsForDays(Days, stats)
	if max != 452 {
		t.Fatalf("Max wrong %f\n", max)
	}
	if min != 0 {
		t.Fatalf("Min wrong %f\n", min)
	}
	if len(entries) != Days {
		t.Fatalf("Wrong number of days for entry %d\n", len(entries))
	}
	if entries[29] != 44 {
		fmt.Printf("%v\n", entries)
		t.Fatalf("Movies 1 has wrong number %f\n", entries[0])
	}
}

func TestReadWithingsStats(t *testing.T) {
	stats := readWithingsStats(`f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\withings-cache.json`)
	if len(stats.Values) == 0 {
		t.Fatalf("Failed to read the Feedly stats")
	}
	if stats.LastUpdated != "2021-12-07T14:01:03+00:00" {
		t.Fatalf("Last updated date wrong")
	}
}

func TestGetWithingsStatsForDays(t *testing.T) {

}

func TestUpdateWithingsStats(t *testing.T) {
	ConfigData.BaseDir = `f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\regenerate\`
	ConfigData.AboutMe.Withings.Client = "e5a14379c73b2dfacf1a0f00009f61ec9a065d309f0bed4fa0a1c60e1e0e3c65"
	ConfigData.AboutMe.Withings.Secret = "0b02bbb52bfa0c471fe4920b214139f898214c7f6eefeed074a9a2937095efb2"
	ConfigData.AboutMe.Withings.AccessToken = "1a5c63b085ca5cd2414c695d4fc3c834f640a589"
	ConfigData.AboutMe.Withings.RefreshToken = "38ffbb657a112512b05466f9e5378a47db7de159"
	ConfigData.AboutMe.Withings.ExpiresAt = 1639670466
	ConfigData.AboutMe.Withings.Cache = "withings-cache.json"
	ConfigData.AboutMe.Withings.OauthURL = "https://wbsapi.withings.net/v2/oauth2"
	ConfigData.AboutMe.Withings.MassURL = "https://wbsapi.withings.net/measure"
	ConfigData.AboutMe.Withings.StepsURL = "https://wbsapi.withings.net/v2/measure"
	x, _ := time.Parse("2006-01-02 15:04:05", "2021-12-01 00:00:00")
	stats := WithingsStats{
		LastUpdated:     "2021-12-01 00:00:00",
		LastUpdatedDate: x,
	}
	stats.Values = make(map[string]WithingsMeasure)
	stats = updateWithingsStats(stats)
}

func TestWithingsSVG(t *testing.T) {
	ConfigData.BaseDir = `f:\Dropbox\swap\golang\vonblog\features\tests\blogstats\regenerate\`
	ConfigData.AboutMe.Withings.Client = "e5a14379c73b2dfacf1a0f00009f61ec9a065d309f0bed4fa0a1c60e1e0e3c65"
	ConfigData.AboutMe.Withings.Secret = "0b02bbb52bfa0c471fe4920b214139f898214c7f6eefeed074a9a2937095efb2"
	ConfigData.AboutMe.Withings.AccessToken = "5c918bf8f395b8333bafe2ee4a5b1b33e8a2b78b"
	ConfigData.AboutMe.Withings.RefreshToken = "a0bc863389ec2f8d9e8a2e7c8d9273e6c4787bad"
	ConfigData.AboutMe.Withings.ExpiresAt = 1739426917
	ConfigData.AboutMe.Withings.Cache = "withings-cache.json"
	ConfigData.AboutMe.Withings.OauthURL = "https://wbsapi.withings.net/v2/oauth2"
	ConfigData.AboutMe.Withings.MassURL = "https://wbsapi.withings.net/measure"
	ConfigData.AboutMe.Withings.StepsURL = "https://wbsapi.withings.net/v2/measure"
	filenameOfWithingsCache := ConfigData.BaseDir + "../regenerate/data/" + ConfigData.AboutMe.Withings.Cache
	filenameOfWithingsWeightSvg := ConfigData.BaseDir + "../regenerate/data/" + ConfigData.AboutMe.Withings.Cache + "-weight.svg"
	filenameOfWithingsStepsSvg := ConfigData.BaseDir + "../regenerate/data/" + ConfigData.AboutMe.Withings.Cache + "-steps.svg"
	Days := 30

	x, _ := time.Parse("2006-01-02 15:04:05", "2021-12-01 00:00:00")
	stats := WithingsStats{
		LastUpdated:     "2021-12-01 00:00:00",
		LastUpdatedDate: x,
	}
	stats.Values = make(map[string]WithingsMeasure)
	stats = updateWithingsStats(stats)

	writeWithingsStats(filenameOfWithingsCache, stats)
	// Get the line
	weight, steps, _, _, wDiff, sMax, _ := getWithingsStatsForDays(Days, stats)
	//graph1 := barSVG(weight, 103.0, 95.0)
	line1, total1 := lineAlone(weight, 103.0, 95.0, "", "rgb(0,0,0,0.5)", "Entries", false, true)
	line2, total2 := lineAlone(steps, sMax, 0, "", "rgb(0,0,0,0.5)", "Entries", true, false)
	// Create and store the SVG
	graph1 := SVGGraphFromPaths(total1, fmt.Sprintf("%d", int(total1)), wDiff, line1)
	ioutil.WriteFile(filenameOfWithingsWeightSvg, graph1, 0666)
	graph2 := SVGGraphFromPaths(total2, fmt.Sprintf("%d", int(total2)), -1, line2)
	ioutil.WriteFile(filenameOfWithingsStepsSvg, graph2, 0666)
}
