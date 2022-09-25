package cmd

import (
	"os"
	"testing"

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
	testroot := "f:/dropbox/swap/golang/vonblog/features/tests/gits/config/"
	viper.SetConfigFile(testroot + "empty.yaml")
	viper.ReadInConfig()
	setEnvironments()
	if gitEnvSet {
		t.Fatalf(`How did this happen?`)
	}
	x, e := os.LookupEnv("GIT_SSH_COMMAND")
	if e {
		t.Fatalf(`Somehow managed to set GIT_SSH_COMMAND %s`, x)
	}

	viper.SetConfigFile(testroot + "noequals.yaml")
	viper.ReadInConfig()
	setEnvironments()
	if gitEnvSet {
		t.Fatalf(`How did this happen?`)
	}
	x, e = os.LookupEnv("GIT_SSH_COMMAND")
	if e {
		t.Fatalf(`Somehow managed to set GIT_SSH_COMMAND %s`, x)
	}

	viper.SetConfigFile(testroot + "full.yaml")
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

// TestGitRunDiff talks to Git and gets a response type.
//func TestGitRunDiff(t *testing.T) {
//	ConfigData.RepositoryDir = "E:/xampp/vonexplaino-bitbucket-static"
//	mep := GitRunDiff()
//	if len(mep.Added) != 1 {
//		t.Fatalf(`Well I got the wrong count of Addeds of %d`, len(mep.Added))
//	}
//	if len(mep.Modified) != 2 {
//		t.Fatalf(`Well I got the wrong count of Modifieds of %d`, len(mep.Modified))
//	}
//	if len(mep.CopyEdit) != 0 {
//		t.Fatalf(`Well I got the wrong amount of CopyEdits of %d`, len(mep.CopyEdit))
//	}
//}

/*
Updating 9e7b035..7f5dec9
Fast-forward
 .../2021/07/reply-pretty-good-hat-platinum.md      | 10 +++
 posts/resume/2018.md                               |  2 +-
 posts/resume/2021.md                               | 93 +++++++++++-----------
 3 files changed, 58 insertions(+), 47 deletions(-)
 create mode 100644 posts/indieweb/2021/07/reply-pretty-good-hat-platinum.md
*/

func TestProcessGitDiffs(t *testing.T) {
	x := ProcessGitDiffs("M	posts/page/uses.md\nA	posts/page/add.md\nA	media/dude.jpg\nC	post/page/change.md\nR	posts/page/rename.md\nD	posts/page/deleted\nU	posts/page/unmerged.md\n")
	if x.Modified[0] != "posts/page/uses.md" {
		t.Fatalf("Failed to parse git status for Modified [%v]\n", x)
	}
	if x.Added[0] != "posts/page/add.md" || x.Added[1] != "media/dude.jpg" {
		t.Fatalf("Failed to parse git status for Added")
	}
}
