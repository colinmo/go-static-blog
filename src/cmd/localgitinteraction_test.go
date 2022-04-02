package cmd

import "testing"

func TestGitRunCommand(t *testing.T) {
	version := runGitCommand("git", []string{"version"})
	if version[0:11] != "git version" {
		t.Fatalf(`Curses, it was supposed to be "git version": "%s"\n`, version[0:11])
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
