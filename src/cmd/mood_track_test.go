package cmd

import (
	"os"
	"testing"
)

func TestMoodSetup(t *testing.T) {
	var err error
	MoodOptions.Filename = ""
	ConfigData.Moods.Token = ""
	ConfigData.Moods.Filename = "steve"
	MoodOptions.Token = ""

	err = setupMoods()
	if err.Error() != "invalid token information " {
		t.Errorf(err.Error())
	}
	if MoodOptions.Filename != "steve" {
		t.Errorf("did not default name correctly %s", MoodOptions.Filename)
	}

	MoodOptions.Token = "jeff"
	ConfigData.Moods.Token = "steve"
	err = setupMoods()
	if err.Error() != "invalid token " {
		t.Errorf("did not find bad token correctly %s", err.Error())
	}
}

func TestReadMoods(t *testing.T) {
	MoodOptions.Filename = `C:\Bob`
	err := doReadMoods()
	if err != nil {
		t.Errorf("failed nonexistent file %v\n", err)
	}

	MoodOptions.Filename = `C:\Users\relap\Dropbox\swap\golang\vonblog\features\tests\moods\basemood.yaml`
	err = doReadMoods()
	if err != nil {
		t.Errorf("bad read %v", err)
	}
	if moodEntries.Moods[2023][1][1].Score != 7 {
		t.Errorf("read fail %v", moodEntries.Moods)
	}
}

func TestWriteMoods(t *testing.T) {
	MoodOptions.Filename = `C:\Users\relap\Dropbox\swap\golang\vonblog\features\tests\moods\newmood.yaml`
	MoodOptions.Text = "OK"
	MoodOptions.Score = 3
	MoodOptions.Date = "20230102"

	os.Remove(MoodOptions.Filename)

	err := doWriteMoods()
	if err != nil {
		t.Errorf("bad write %v", err)
	}

	bob, err := os.ReadFile(MoodOptions.Filename)
	if err != nil {
		t.Errorf("couldn't read written file %v", err)
	}
	if string(bob) != "moods:\n  2023:\n    1:\n      2:\n        text: OK\n        score: 3\n" {
		t.Errorf("bad content\n%s\n", bob)
	}

	moodEntries.setMood(2025, 1, 20, MoodEntryS{Text: "Yo", Score: 1})
	err = moodEntries.writeMoods()
	if err != nil {
		t.Errorf("bad write %v", err)
	}

	bob, err = os.ReadFile(MoodOptions.Filename)
	if err != nil {
		t.Errorf("couldn't read written file %v", err)
	}
	if string(bob) != "moods:\n  2023:\n    1:\n      2:\n        text: OK\n        score: 3\n  2025:\n    1:\n      20:\n        text: Yo\n        score: 1\n" {
		t.Errorf("bad content\n%s\n", bob)
	}

}
