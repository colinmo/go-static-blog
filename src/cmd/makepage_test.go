package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func pageExists(filename string) error {
	return godog.ErrPending
}

func runCreatePage(filename string) error {
	return godog.ErrPending
}

func shouldSeeAFileWithContents(filename string, contents string) error {
	return godog.ErrPending
}

func TestParseUnknownDateFormat(t *testing.T) {
	type testTimes struct {
		From string
		To   time.Time
	}
	thisLocation := time.Now().Location()
	stringsToParse := []testTimes{
		{
			From: "Sun, 31 May 2009 10:30:14 +0000",
			To:   time.Date(2009, 5, 31, 20, 30, 14, 0, thisLocation),
		},
		{
			From: "Fri Dec 7 22:22:12 2018 +1000",
			To:   time.Date(2018, 12, 7, 22, 22, 12, 00, thisLocation),
		},
		{
			From: "2009-08-04 09:51:30 +0000",
			To:   time.Date(2009, 8, 4, 19, 51, 30, 0, thisLocation),
		},
		{
			From: "27 May 2014 23:19:41 +1000",
			To:   time.Date(2014, 5, 27, 23, 19, 41, 0, thisLocation),
		},
		{
			From: "May 2, 2014 6:25 am +1000",
			To:   time.Date(2014, 5, 2, 6, 25, 0, 0, thisLocation),
		},
	}
	for _, stringToParse := range stringsToParse {
		date, err := parseUnknownDateFormat(stringToParse.From)
		if err != nil || !date.Equal(stringToParse.To) {
			t.Fatalf(
				"Failed to parse %s, got %v, wanted %v, %v\n",
				stringToParse.From,
				date,
				stringToParse.To,
				err)
		}
	}
}

func TestCreateCodePage(t *testing.T) {
	testroot := `E:\xampp\vonexplaino-bitbucket-static`
	ConfigData.RepositoryDir = testroot
	ConfigData.BaseURL = "https://vonexplaino.com/blog/"
	ConfigData.TemplateDir = "/home/colin/Laboratory/go-static-blog/templates/"
	result, frontMatter, error := parseString(`---
Title: Code
Created: 2015-11-18 20:32:00 +1000
Type: page
---
## Common book

Learnings and personal library.

* [Selenium](https://vonexplaino.com/blog/posts/page/selenium-ide.html)
* [PHP/ Composer](https://vonexplaino.com/blog/posts/page/php-composer.html)
* [SVG](https://vonexplaino.com/blog/posts/page/svg.html)
* [Apache](https://vonexplaino.com/blog/posts/page/apache.html)
* [Git](https://vonexplaino.com/blog/posts/article/2020/02/git-status.html)

Or just look at the [Code tag](https://vonexplaino.com/blog/tag/code-1.html)

## Stuff to play with

<div class="wunderkammer" markdown="1">
[![Business Card Maze](/blog/media/code/TitlePlate-Maze.png) *Maze generator*](/code/maze)
[![Cog Maker](/blog/media/code/TitlePlate-Cog.png) *Cog Maker*](/code/cog)
[![Deck of Many Things](/blog/media/code/TitlePlate-DoMT.png) *Deck of Many Things*](/code/domt)
[![Random Magical Effect](/blog/media/code/TitlePlate-RME.png) *Random Magical Effect*](/code/rme)
[![Site Jageriser](/blog/media/code/TitlePlate-Jaeger.png) *Site Jageriser*](/blog/posts/article/2018/01/jageriser-wanna-play.html)
[![Fortune Deck](/blog/media/code/TitlePlate-Fortune.png) *Fortune Deck*](/code/fortune-deck)
[![Fuzion Lifepath Generator](/blog/media/code/TitlePlate-Fuzion.png) *Fuzion Lifepath Generator*](/code/lifepath)
[![GURPS 5 Lite Character Generator](/blog/media/code/TitlePlate-GURPS.png) *GURPS 5 Lite Character Generator*](/code/gurps)
[![RISUS Character Generator](/blog/media/code/TitlePlate-RISUS.png) *RISUS Character Generator*](/code/risus)
[![Trinity Character Creator](/blog/media/code/TitlePlate-Trinity.png) *Trinity Character Creator*](/code/trinity-character-creator/)
</div>

See more of [Everway](http://everwayan.blogspot.com.au/p/everway-links.html), where the Fortune deck is from.

See more of [GURPS](http://www.sjgames.com/gurps/).

See more of [D&D 5e](http://dnd.wizards.com/) where the Deck of Many things is from. The random magical effects are from [the Net Libram Book of Random Magical Effects v2.00](http://centralia.aquest.com/downloads/NLRMEv2.pdf)

See more of [RISUS](http://www.risusiverse.com/).

See more of [Girl Genius](http://www.girlgeniusonline.com/), where the Jagers are from.

See more of [Trinity](http://theonyxpath.com/category/worlds/trinitycontinuum/).

<p style="text-align: center" markdown="1">![Days since last fatal error](https://vonexplaino.com/code/days-since.svg)</p>
<style>
    .wunderkammer {
        display: grid;
        max-width: 100%;
    }
</style>`, ConfigData.RepositoryDir+"posts/page/2015/11/code.md")
	if error != nil {
		t.Fatalf("Failed to parse Code: %v\n", error)
	}
	if len(result) == 0 {
		t.Fatalf("Failed to create a result")
	}
	if frontMatter.Title == "" {
		t.Fatalf("Failed to marse MD")
	}
	if strings.Contains(result, `markdown="1"`) {
		t.Fatalf("Did not parse all markdowns")
	}
}

func TestCreateResume(t *testing.T) {
	testroot := `C:\Laboratory\temp\`
	ConfigData.RepositoryDir = testroot
	ConfigData.BaseURL = "https://vonexplaino.com/blog/"
	ConfigData.TemplateDir = "c:/users/relap/dropbox/swap/golang/vonblog/templates/"
	result, frontMatter, error := parseString(`---
Title: Colin Morris
Created: 2024-04-06T22:15:50+1000
Updated: 2024-04-06T22:15:50+1000
Tags: [code,colin]
Slug: resume-of-colin-morris
Synopsis: I strive to use my analytical, organisational and technical skills and experience to facilitate long lasting and enjoyable solutions for a variety of user desires.
Resume:
    Contact:
        name: Colin Morris
        honorific: Mr.
        email: professor@vonexplaino.com
        p-job-title: Solution Architect and Programmer
        u-photo: "/blog/media/2022/01/23/BusinessCard-Thumb.png"
        u-url: "https://vonexplaino.com/"
        u-key: E2895935D852A422
        u-logo: "https://vonexplaino.com/theme/images/header-horizontal.png"
        linkedin: "https://www.linkedin.com/in/colinmo"
    Education:
        -   p-name: Bachelor's degree in Information Technology (Honours)
            dt-start: 1994-01-01T00:00:00 +1000
            dt-end: 1997-11-01T00:00:00 +1000
            u-url: "mailto:verifications@griffith.edu.au"
            p-category: Tertiary
            p-location: Griffith University
        -   p-name: TOGAF&#169; Certified
            dt-start: 2016-01-01T00:00:00 +1000
            u-url: "https://www.youracclaim.com/badges/d2207369-850a-46d0-b924-28f7e0cf8ff5"
            p-category: Certification
            p-location: The Open Group
        -   p-name: Microsoft Azure certifications
            dt-start: 2022-09-21T00:00:00 +1000
            u-url: "https://learn.microsoft.com/en-gb/users/colinmorris-7354/"
            p-category: Certification
            p-location: Microsoft Learn
    FlatSkills:
        Methodologies:
            Agile: "p"
            Behaviour Driven Development: "p"
            Business_Analysis: "p"
            Business_Process_Improvement: "p"
            ITIL: ""
            Prince2: ""
            Solution_Architecture: "p"
            TOGAF: "p"
        Languages:
            CSS: p
            Go: p
            JavaScript: p
            HTML: p
            Perl: 
            PHP: p
            PL/SQL: 
            Python: 
            Shell scripts: p
            SQL: p
        Libraries:
            Azure Cloud: 
            Behat + Mink: 
            Chart.js: 
            D3.js: 
            DJango: 
            jQuery: 
            Microsoft DevOps: 
            New Relic: p
            Pandoc: p
            Regular expressions: p
            REST: p
            Selenium: 
            SOAP: 
            Swagger/ OpenAPI: p
            Symfony: 
    Affiliation: []
    Experience:
        -   p-name: Solution Architect (Integrator)
            p-summary: Provide expertise to identify and translate system requirements into software design documentation, identifying possible existing solutions (internal and external).
            dt-start: 2021-02-01T00:00:00 +1000
            p-description: |
                Working with researchers, academics, fellow architects, technical staff, and vendors to collaborately design, document, review, and implement solutions to the benefit of the university. I am attached to the research area, so the work includes novel areas and novel technologies such as AI. The architects work very closely with information management and cybersecurity, so we require a knowledge of these areas to coordinate well.
                
                Griffith is a TOGAF-based architecture structure.
            u-url: "https://www.griffith.edu.au/"
            p-location: Griffith University
            p-category: Work History
        -   p-name: Previous experience
            p-summary: I've been working in the education industry for over twenty years.
            dt-start: 1997-12-01T00:00:00 +1000
            dt-end: 2021-02-01T00:00:00 +1000
            p-description: |
                At Griffith University I've worked with the finance systems, the student systems, the HR systems, research systems, and everything in between. The roles have covered dedicated system support, project development, development team leader, and custom development lead.
            p-location: Griffith University
            p-category: Work History
        -   p-name: "Code of the Coder"
            dt-published: 2018-11-08T00:00:00 +1000
            u-url: "http://www.lulu.com/shop/colin-morris/code-of-the-coder/paperback/product-23864781.html"
            u-uid: "http://www.lulu.com/shop/colin-morris/code-of-the-coder/paperback/product-23864781.html"
            p-category: Publication
            p-summary: |
                <blockquote style="display: grid;grid-template-columns: 100px auto;justify-items: center; align-items: center;gap:14px;">
                <a href="https://www.lulu.com/shop/colin-morris/code-of-the-coder/paperback/product-23864781.html"><img src="/blog/media/2018/11/code-of-the-coder-cover.jpeg" width="100" alt="Book cover for Code of the Coder"></a>
                    People claim to be Code Ninja or CSS Samurai, but how many of them follow a code? How many of them practice daily katas to keep in the best condition? This book foolishly applies the Seven Virtues of Bushido and the Eighteen Disciplines of Togekure-ryu ninjutsu to the coding arts, mistakenly finding some wisdom along the way.
                </blockquote>
---
I've been working in the information technology industry since 1997. During that time I've had a variety of roles - from finance system maintenance through to an engineer on PeopleSoft technologies through to web development and currently I'm a solution architect. That's been a great benefit from working at Griffith University - the breadth of experiences and people I've been able to work with.
	`, ConfigData.RepositoryDir+"posts/resume/2021.md")
	if error != nil {
		t.Fatalf("Failed to parse Code: %v\n", error)
	}
	if len(result) == 0 {
		t.Fatalf("Failed to create a result")
	}
	if frontMatter.Title == "" {
		t.Fatalf("Failed to marse MD")
	}
	if strings.Contains(result, `markdown="1"`) {
		t.Fatalf("Did not parse all markdowns")
	}
	if len(frontMatter.Resume.FlatSkills.Methodologies) != 8 {
		fmt.Printf("%v\n", frontMatter.Resume.FlatSkills)
		t.Fatalf("Bad count of methodologies %d expected %d", len(frontMatter.Resume.FlatSkills.Methodologies), 8)
	}
	if frontMatter.Resume.FlatSkills.MethodologyOrder[0] != "Agile" || frontMatter.Resume.FlatSkills.MethodologyOrder[7] != "TOGAF" {
		t.Fatalf("Order wrong %v\n", frontMatter.Resume.FlatSkills.MethodologyOrder)
	}
}

func TestCreateReview(t *testing.T) {

	testroot := `C:\Laboratory\temp\`
	ConfigData.RepositoryDir = testroot
	ConfigData.BaseURL = "https://vonexplaino.com/blog/"
	ConfigData.TemplateDir = "c:/users/relap/dropbox/swap/golang/vonblog/templates/"
	result, frontMatter, error := parseString(`---
Title: "Review: In Sound Mind"
Tags: [game,epic]
Created: "2022-04-24T18:58:43+1000"
Updated: "2022-04-24T18:58:43+1000"
Type: review
Synopsis: "In Sound Mind is an imaginative first-person psychological horror with frenetic puzzles, unique boss fights, and original music by The Living Tombstone. Journey within the inner workings of the one place you can’t seem to escape—your own mind."
FeatureImage: /blog/media/2022/04/in-sound-mind.webp
Item:
    url: "https://store.epicgames.com/en-US/p/in-sound-mind"
    image: /blog/media/2022/04/in-sound-mind.webp
    name: In Sound Mind
    type: item
    rating: 5
---
In Sound Mind was one of the weekly free games earlier this year. Most of these games I pick up, play for a bit, get a smile, get bored, and get on with things. In Sound Mind's gameplay, steady reveal, tape-based psychology gimick and the "GOTY 10/10" acheivement had me hooked. Very little in the way of shooty times, really; and the stealth statistic seemed entirely pointless - but the game, atmosphere, and sheer mind-squirreliness was enthralling.

The game starts off in a building that's run down in a flooded city, but you find ways out into the minds of your patients. Boy are you in for a wild time in each of those, with a unique mechanic in almost each of them. The manifestations of mental anguish are spellbounding and the spook factor is high.

Very much recommended.

You can pet the cat.
	`,
		ConfigData.RepositoryDir+"posts/review/2022/04/in-sound-mind.md")
	if error != nil {
		t.Fatalf("Failed to parse: %v\n", error)
	}
	if len(result) == 0 {
		t.Fatalf("Failed to create a result")
	}
	if frontMatter.Title == "" {
		t.Fatalf("Failed to marshal MD")
	}
	if frontMatter.Item.Name == "" {
		t.Fatalf("Failed to get name")
	}
}

func TestTextToSlug(t *testing.T) {
	for expect, test := range map[string]string{
		"bobiscool":        "bobiscool",
		"bob-is-real-cool": "bob is real cool",
		"bob-is-cool-mate": "bob is cool mate!",
		"bobis-coolmate":   "bobis!@#!@#coolmate",
		"am-i-cool-now":    "'am i cool now?'",
	} {
		if expect != textToSlug(test) {
			t.Fatalf("Failed to convert %s=>%s [%s]", test, expect, textToSlug(test))
		}
	}
}

func TestDefaultTypes(t *testing.T) {
	if defaultType([]string{"article"}, "/posts/article/2022/01/dude.md") != "article" {
		t.Fatalf("Failed to get an type from a url")
	}
	if defaultType([]string{"article", "reply"}, "/posts/dude/2022/01/dude.md") != "" {
		t.Fatalf("Erroneously got a type from a bad path")
	}
}

func TestDefaultFeatureImage(t *testing.T) {
	var x string
	wordpressThumbnailTemplate = `https://wp.dude.com/thumbnail?%s`
	tests := []struct {
		fm   FrontMatter
		xpct string
	}{
		{
			fm:   FrontMatter{InReplyTo: "https://testme.com/"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
		{
			fm:   FrontMatter{BookmarkOf: "https://testme.com/"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
		{
			fm:   FrontMatter{FavoriteOf: "https://testme.com/"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
		{
			fm:   FrontMatter{RepostOf: "https://testme.com/"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
		{
			fm:   FrontMatter{LikeOf: "https://testme.com/"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
		{
			fm:   FrontMatter{InReplyTo: "https://testme.com/", LikeOf: "https://mep.com"},
			xpct: `https://wp.dude.com/thumbnail?https%3A%2F%2Ftestme.com%2F`,
		},
	}

	for _, fTest := range tests {
		x = defaultFeatureImage(&fTest.fm)
		if x != fTest.xpct {
			t.Fatalf("Bad thumbnail for [%s][%s]", x, fTest.xpct)
		}
	}
}

func TestToTwigVariables(t *testing.T) {
	dude := FrontMatter{
		Title: "Dude",
	}
	content := "xxXX!"
	mike := toTwigVariables(&dude, content)
	if mike["content"] != "xxXX!" {
		t.Fatalf("Did not get content")
	}
	if mike["title"] != "Dude" {
		t.Fatalf("Did not get title %v\n", mike)
	}
}
