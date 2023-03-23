/*
*
Copyright © 2023 Colin Morris <relapse@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type MoodOptionsS struct {
	Score      int
	Text       string
	Date       string
	DateAsDate time.Time
	Filename   string
	Read       bool
	Write      bool
	Token      string
}

var MoodOptions MoodOptionsS

type MoodEntryS struct {
	Text  string `yaml:"text" json:"text"`
	Score int    `yaml:"score" json:"score"`
}
type MoodEntriesS struct {
	Moods map[int]map[time.Month]map[int]MoodEntryS `json:"moods" yaml:"moods"`
}

var moodEntries MoodEntriesS

// moodCmd updates/ displays mood
var moodCmd = &cobra.Command{
	Use:   "mood",
	Short: "Tracks the mood",
	Long:  `Tracks the mood in a YAML file - a 0-10 score of the mood 'goodness' as well as a text string attached describing the mood`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(MoodOptions.Filename) == 0 {
			MoodOptions.Filename = ConfigData.Moods.Filename
		}
		if len(ConfigData.Moods.Token) == 0 ||
			len(MoodOptions.Token) == 0 {
			log.Fatalf("Invalid token information")
		}
		if MoodOptions.Token != ConfigData.Moods.Token {
			log.Fatalf("Invalid token")
		}
		if MoodOptions.Read {
			moodEntries.readMoods()

			if err == nil {
				x, _ := json.Marshal(moodEntries)
				fmt.Printf("%s\n", x)
			} else {
				fmt.Printf("%s\n", err)
			}
		} else if MoodOptions.Write {
			// ensure required parameters
			errors := []string{}
			if len(MoodOptions.Text) == 0 {
				errors = append(errors, "text")
			}
			if MoodOptions.Score < 0 || MoodOptions.Score > 10 {
				errors = append(errors, "score")
			}
			MoodOptions.DateAsDate, err = time.Parse("20060102", MoodOptions.Date)
			if err != nil {
				errors = append(errors, "Date")
			}
			if len(errors) == 0 {
				err := moodEntries.readMoods()
				if err == nil {
					if _, ok := moodEntries.Moods[MoodOptions.DateAsDate.Year()]; !ok {
						moodEntries.Moods[MoodOptions.DateAsDate.Year()] = map[time.Month]map[int]MoodEntryS{}
					}
					if _, ok := moodEntries.Moods[MoodOptions.DateAsDate.Year()][MoodOptions.DateAsDate.Month()]; !ok {
						moodEntries.Moods[MoodOptions.DateAsDate.Year()][MoodOptions.DateAsDate.Month()] = map[int]MoodEntryS{}
					}
					moodEntries.setMood(MoodOptions.DateAsDate.Year(), MoodOptions.DateAsDate.Month(), MoodOptions.DateAsDate.Day(), MoodEntryS{
						Text:  MoodOptions.Text,
						Score: MoodOptions.Score,
					})
					moodEntries.writeMoods()
				}
			} else {
				log.Fatalf("Missing or invalid values for %s", strings.Join(errors, ", "))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(moodCmd)
	moodCmd.Flags().IntVarP(&MoodOptions.Score, "score", "s", -1, "Score (0-10)")
	moodCmd.Flags().StringVarP(&MoodOptions.Text, "text", "t", "", "Description of the mood")
	moodCmd.Flags().StringVarP(&MoodOptions.Date, "date", "d", time.Now().Local().Format("20060102"), "Date (YYYYMMDD) of entry")
	moodCmd.Flags().StringVarP(&MoodOptions.Filename, "filename", "f", "", "File to save into")
	moodCmd.Flags().BoolVarP(&MoodOptions.Read, "read", "r", false, "Read to screen")
	moodCmd.Flags().BoolVarP(&MoodOptions.Write, "update", "u", false, "Update the moodfile")
	moodCmd.Flags().StringVarP(&MoodOptions.Token, "token", "", "", "Token for auth")
}

func (m *MoodEntriesS) readMoodsFile() ([]byte, error) {
	var returning = []byte{}
	if MoodOptions.Filename == "" {
		return returning, fmt.Errorf("empty filename")
	}
	GitPull()
	return os.ReadFile(MoodOptions.Filename)
}

func (m *MoodEntriesS) byteToMoods(bytes []byte) error {
	err := yaml.Unmarshal(bytes, &m)
	return err
}

func (m *MoodEntriesS) readMoods() error {
	content, err := m.readMoodsFile()
	if err == nil {
		m.byteToMoods(content)
	}
	return err
}

func (m *MoodEntriesS) writeMoods() error {
	x, err := yaml.Marshal(m)
	if err == nil {
		err = os.WriteFile(MoodOptions.Filename, x, 0666)
	}
	GitAdd(MoodOptions.Filename)
	GitCommit("Mood update")
	GitPush()
	return err
}

func (m *MoodEntriesS) setMood(year int, month time.Month, day int, moodEntry MoodEntryS) error {
	m.Moods[year][month][day] = moodEntry
	return nil
}
