package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

var gitEnvSet bool

func setEnvironments() {
	envThing := viper.GetString("gitenv")
	if envThing != "" {
		mep := strings.SplitN(envThing, "=", 2)
		if len(mep) > 1 {
			os.Setenv(mep[0], mep[1])
		}
		gitEnvSet = true
	}
}

var gitCommand = "git"

func GitPull() {
	runGitCommand(gitCommand, []string{"pull"})
}

func GitFetch() {
	runGitCommand(gitCommand, []string{"fetch"})
}

type GitDiffs struct {
	Modified   []string
	CopyEdit   []string
	RenameEdit []string
	Added      []string
	Deleted    []string
	Unmerged   []string
}

func ProcessGitDiffs(resultsStr string) GitDiffs {
	returnDiffs := GitDiffs{
		Modified:   make([]string, 0),
		CopyEdit:   make([]string, 0),
		RenameEdit: make([]string, 0),
		Added:      make([]string, 0),
		Deleted:    make([]string, 0),
		Unmerged:   make([]string, 0),
	}
	for _, line := range strings.Split(resultsStr, "\n") {
		if len(line) == 0 {
			break
		}
		index := string(line[0])
		rest := strings.TrimSpace(string(line[1:]))
		switch index {
		case "M":
			returnDiffs.Modified = append(returnDiffs.Modified, rest)
		case "C":
			returnDiffs.CopyEdit = append(returnDiffs.CopyEdit, rest)
		case "R":
			returnDiffs.RenameEdit = append(returnDiffs.RenameEdit, rest)
		case "A":
			returnDiffs.Added = append(returnDiffs.Added, rest)
		case "D":
			returnDiffs.Deleted = append(returnDiffs.Deleted, rest)
		case "U":
			returnDiffs.Unmerged = append(returnDiffs.Unmerged, rest)
		}
	}
	return returnDiffs
}

func GitRunDiff() GitDiffs {
	return ProcessGitDiffs(runGitCommand(gitCommand, []string{"diff", "master", "origin/master", "--name-status"}))
}

func GitGetFileAges(file string) {
	runGitCommand(gitCommand, []string{"diff", "master", "origin/master", "--name-status"})
}

func runGitCommand(command string, parameters []string) string {
	if !gitEnvSet {
		setEnvironments()
	}
	var out bytes.Buffer
	cmd := exec.Command(command, parameters...)
	cmd.Dir = ConfigData.RepositoryDir
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Printf("Command didn't run %s, %v\n", command, parameters)
		log.Fatal(err)
	}
	return out.String()
}
