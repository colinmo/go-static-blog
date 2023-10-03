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
var gitCommand = "git"

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

func GitPull() {
	l, e := runGitCommand(gitCommand, []string{"pull"})
	if e != nil {
		log.Fatalf("Failed %s\n%s\n", l, e)
	}
}

func GitFetch() {
	runGitCommand(gitCommand, []string{"fetch"})
}

func GitAdd(filename string) {
	runGitCommand(gitCommand, []string{"add", filename})
}

func GitCommit(message string) {
	runGitCommand(gitCommand, []string{"commit", "--message", fmt.Sprintf(`"%s"`, message)})
}

func GitPush() {
	runGitCommand(gitCommand, []string{"push"})
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
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			break
		}
		index := line[0:1]
		rest := strings.TrimSpace(line[2:])
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
	runGitCommand(gitCommand, []string{"fetch", "origin", "master"})
	result, _ := runGitCommand(gitCommand, []string{"diff", "--name-status", "FETCH_HEAD"})
	return ProcessGitDiffs(result)
}

func runGitCommand(command string, parameters []string) (string, error) {
	if !gitEnvSet {
		setEnvironments()
	}
	var out bytes.Buffer
	cmd := exec.Command(command, parameters...)
	cmd.Dir = ConfigData.RepositoryDir
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command didn't run %s, %v", command, parameters)
	}
	return out.String(), nil
}
