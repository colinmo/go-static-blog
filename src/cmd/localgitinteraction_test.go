package cmd

import (
	"os"
	"path/filepath"
	"testing"

	testdataloader "github.com/peteole/testdata-loader"
	"github.com/spf13/viper"
)

func TestGitRunCommand(t *testing.T) {
	version, _ := runGitCommand(gitCommand, []string{"version"})
	if version[0:11] != "git version" {
		t.Fatalf(`Curses, it was supposed to be "git version": "%s"\n`, version[0:11])
	}
}

func TestProcessEnvVariables(t *testing.T) {
	var err error
	testroot := filepath.Clean(testdataloader.GetBasePath() + "/../features/tests/gits/config/")
	viper.SetConfigFile(filepath.Join(testroot, "empty.yaml"))
	viper.ReadInConfig()
	setEnvironments()
	if gitEnvSet {
		t.Fatalf(`How did this happen?`)
	}
	x, e := os.LookupEnv("GIT_SSH_COMMAND")
	if e {
		t.Fatalf(`Somehow managed to set GIT_SSH_COMMAND %s`, x)
	}

	viper.SetConfigFile(filepath.Join(testroot, "noequals.yaml"))
	viper.ReadInConfig()
	setEnvironments()
	if gitEnvSet {
		t.Fatalf(`How did this happen?`)
	}
	x, e = os.LookupEnv("GIT_SSH_COMMAND")
	if e {
		t.Fatalf(`Somehow managed to set GIT_SSH_COMMAND %s`, x)
	}

	viper.SetConfigFile(filepath.Join(testroot, "full.yaml"))
	err = viper.ReadInConfig()
	if err != nil {
		t.Fatalf(`Failed to read in %v`, err)
	}
	setEnvironments()
	if !gitEnvSet {
		t.Fatalf(`How did this happen?`)
	}
	x, e = os.LookupEnv("GIT_SSH_COMMAND")
	if !e {
		t.Fatalf(`Somehow managed to not set GIT_SSH_COMMAND %v`, e)
	}
	if x != "ssh -i /home/relapse/blog-bitbucket/bitbucket" {
		t.Fatalf(`GIT_SSH_COMMAND set wrong %s`, x)
	}
}

func TestProcessGitDiffs(t *testing.T) {
	x := ProcessGitDiffs("M	posts/page/uses.md\nA	posts/page/add.md\nA	media/dude.jpg\nC	post/page/change.md\nR	posts/page/rename.md\nD	posts/page/deleted\nU	posts/page/unmerged.md\n")
	if x.Modified[0] != "posts/page/uses.md" {
		t.Fatalf("Failed to parse git status for Modified [%v]\n", x)
	}
	if x.Added[0] != "posts/page/add.md" || x.Added[1] != "media/dude.jpg" {
		t.Fatalf("Failed to parse git status for Added")
	}
}
