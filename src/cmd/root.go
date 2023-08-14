/*
Copyright © 2021 Colin Morris

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
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

// GLOBAL VARIABLES
var cfgFile string

// var colorBlackOpacity50 = "rgb(0,0,0,0.5)"
// var gmtDateFormat = "2006-01-02T15:04:05.0000Z"
var jsonHeaders = [][]string{
	{"Accept-Language", "en"},
	{"Content-type", "application/json"},
}
var blogTimezone = "Australia/Brisbane"
var baseDirectoryForPosts = "posts/"

type Metadata struct {
	Title       string
	Description string
	Language    string
	Ttl         int
	Webmaster   string
}
type BlogStats struct {
	Days int
}
type Trakt struct {
	ID           string
	Secret       string
	AccessToken  string
	RefreshToken string
	Cache        string
}
type Feedly struct {
	Key   string
	Cache string
	URL   string
}
type Withings struct {
	Client       string
	Secret       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	Cache        string
	OauthURL     string
	MassURL      string
	StepsURL     string
}
type Mastodon struct {
	URL   string
	Token string
}
type AboutMe struct {
	Trakt    Trakt
	Feedly   Feedly
	Withings Withings
}
type Thumbnails struct {
	Height    uint
	Width     uint
	Extension string
	Type      string
}
type Syndics struct {
	Mastodon Mastodon
}
type Moods struct {
	Filename string
	Token    string
}
type ConfigDataStruct struct {
	BaseDir       string
	BaseURL       string
	TempDir       string
	RepositoryDir string
	PerPage       int
	TemplateDir   string
	Metadata      Metadata
	BlogStats     BlogStats
	AboutMe       AboutMe
	Thumbnails    Thumbnails
	Syndication   Syndics
	TagSnippets   []string
	Moods         Moods
}

var ConfigData ConfigDataStruct

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vonblog",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
			examples and usage of using your application. For example:

			Cobra is a CLI library for Go that empowers applications.
			This application is a tool to generate the needed files
			to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vonblog.yaml)")
	rootCmd.PersistentFlags().StringVar(&ConfigData.BaseDir, "basedir", "", "base dir to output to")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// HTTP Client initting for test mocking
	Client = &http.Client{}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".vonblog" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".vonblog")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		ConfigData.BaseDir = viper.GetString("baseDir")
		ConfigData.BaseURL = viper.GetString("baseUrl")
		ConfigData.RepositoryDir = viper.GetString("repositoryDir")
		ConfigData.PerPage = viper.GetInt("perpage")
		ConfigData.TemplateDir = viper.GetString("templateDir")
		ConfigData.Metadata.Title = viper.GetStringMapString("metadata")["title"]
		ConfigData.Metadata.Description = viper.GetStringMapString("metadata")["description"]
		ConfigData.Metadata.Language = viper.GetStringMapString("metadata")["language"]
		x := viper.GetStringMapString("metadata")["ttl"]
		ConfigData.Metadata.Ttl, _ = strconv.Atoi(x)
		ConfigData.Metadata.Webmaster = viper.GetString("metadata.webmaster")
		ConfigData.BlogStats.Days, _ = strconv.Atoi(viper.GetStringMapString("blogstats")["days"])
		// ABOUT ME
		ConfigData.AboutMe.Trakt.ID = viper.GetString("aboutme.trakt.id")
		ConfigData.AboutMe.Trakt.Secret = viper.GetString("aboutme.trakt.secret")
		ConfigData.AboutMe.Trakt.AccessToken = viper.GetString("aboutme.trakt.accesstoken")
		ConfigData.AboutMe.Trakt.RefreshToken = viper.GetString("aboutme.trakt.refreshtoken")
		ConfigData.AboutMe.Trakt.Cache = viper.GetString("aboutme.trakt.cache")
		ConfigData.AboutMe.Feedly.Key = viper.GetString("aboutme.feedly.key")
		ConfigData.AboutMe.Feedly.Cache = viper.GetString("aboutme.feedly.cache")
		ConfigData.AboutMe.Feedly.URL = viper.GetString("aboutme.feedly.url")
		ConfigData.AboutMe.Withings.Client = viper.GetString("aboutme.withings.client")
		ConfigData.AboutMe.Withings.Secret = viper.GetString("aboutme.withings.secret")
		ConfigData.AboutMe.Withings.AccessToken = viper.GetString("aboutme.withings.accesstoken")
		ConfigData.AboutMe.Withings.RefreshToken = viper.GetString("aboutme.withings.refreshtoken")
		ConfigData.AboutMe.Withings.ExpiresAt = viper.GetInt64("aboutme.withings.expiresat")
		ConfigData.AboutMe.Withings.Cache = viper.GetString("aboutme.withings.cache")
		ConfigData.AboutMe.Withings.OauthURL = viper.GetString("aboutme.withings.oauthurl")
		ConfigData.AboutMe.Withings.MassURL = viper.GetString("aboutme.withings.massurl")
		ConfigData.AboutMe.Withings.StepsURL = viper.GetString("aboutme.withings.stepsurl")
		// Thumbnails
		ConfigData.Thumbnails.Width = uint(viper.GetInt("thumbnails.width"))
		ConfigData.Thumbnails.Height = uint(viper.GetInt("thumbnails.height"))
		ConfigData.Thumbnails.Extension = viper.GetString("thumbnails.extension")
		ConfigData.Thumbnails.Type = viper.GetString("thumbnails.type")
		ConfigData.TempDir = viper.GetString("tempDir")
		// Syndications
		ConfigData.Syndication.Mastodon.URL = viper.GetString("syndication.mastodon.url")
		ConfigData.Syndication.Mastodon.Token = viper.GetString("syndication.mastodon.token")
		// MISC
		ConfigData.TagSnippets = viper.GetStringSlice("tagSnippets")
		// MOODS
		ConfigData.Moods.Filename = viper.GetString("moods.filename")
		ConfigData.Moods.Token = viper.GetString("moods.token")
	}
}

var DateOfExecution = time.Now()

func getStartOfEverything() (time.Time, *time.Location) {
	l, _ := time.LoadLocation(blogTimezone)
	startOfEverything := time.Date(1970, 1, 1, 0, 0, 0, 0, l)
	return startOfEverything, l
}

/** Global functions? **/
func PrintIfNotSilent(toPrint string) {
	if !Silent {
		fmt.Print(toPrint)
	}
}

func MyReadFilename(f string) ([]byte, error) {
	g, err := os.Open(f)
	if err != nil {
		return []byte(""), err
	}
	defer g.Close()
	return MyReadFile(g)
}
func MyReadFile(g *os.File) ([]byte, error) {
	var buf bytes.Buffer
	io.Copy(&buf, g)
	return buf.Bytes(), nil
}
