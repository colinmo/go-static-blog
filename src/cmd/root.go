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
	"fmt"
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
type Mastodon struct {
	URL   string
	Token string
}
type Bluesky struct {
	URL      string
	Userid   string
	Password string
}
type Thumbnails struct {
	Height    uint
	Width     uint
	Extension string
	Type      string
}
type Syndics struct {
	Mastodon Mastodon
	Bluesky  Bluesky
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
	Thumbnails    Thumbnails
	Syndication   Syndics
	TagSnippets   []string
	Moods         Moods
}

var ConfigData ConfigDataStruct

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vonblog",
	Short: "von Explaino blog automator",
	Long:  "The command line tool for regenerating and managing the Professor von Explaino website.",
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
		// Thumbnails
		ConfigData.Thumbnails.Width = uint(viper.GetInt("thumbnails.width"))
		ConfigData.Thumbnails.Height = uint(viper.GetInt("thumbnails.height"))
		ConfigData.Thumbnails.Extension = viper.GetString("thumbnails.extension")
		ConfigData.Thumbnails.Type = viper.GetString("thumbnails.type")
		ConfigData.TempDir = viper.GetString("tempDir")
		// Syndications
		ConfigData.Syndication.Mastodon.URL = viper.GetString("syndication.mastodon.url")
		ConfigData.Syndication.Mastodon.Token = viper.GetString("syndication.mastodon.token")
		ConfigData.Syndication.Bluesky.URL = viper.GetString("syndication.bluesky.url")
		ConfigData.Syndication.Bluesky.Userid = viper.GetString("syndication.bluesky.userid")
		ConfigData.Syndication.Bluesky.Password = viper.GetString("syndication.bluesky.password")
		// MISC
		ConfigData.TagSnippets = viper.GetStringSlice("tagSnippets")
		// MOODS
		ConfigData.Moods.Filename = viper.GetString("moods.filename")
		ConfigData.Moods.Token = viper.GetString("moods.token")
	}
}

var DateOfExecution = time.Now()

/** Global functions? **/
func PrintIfNotSilent(toPrint string) {
	if !Silent {
		fmt.Print(toPrint)
	}
}
