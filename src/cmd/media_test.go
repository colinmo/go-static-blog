package cmd

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	testdataloader "github.com/peteole/testdata-loader"
)

func TestDefaults(t *testing.T) {
	ConfigData.Thumbnails.Width = 20
	ConfigData.Thumbnails.Height = 20
	ConfigData.Thumbnails.Extension = ".dude"
	defaultsForMe()
	if ThumbnailOptions.Width != int(ConfigData.Thumbnails.Width) {
		t.Fatalf("Width default failed %d\n", ThumbnailOptions.Width)
	}
	if ThumbnailOptions.Height != int(ConfigData.Thumbnails.Height) {
		t.Fatalf("Height default failed %d\n", ThumbnailOptions.Height)
	}
	if ThumbnailOptions.Extension != ConfigData.Thumbnails.Extension {
		t.Fatalf("Extension default failed %s\n", ThumbnailOptions.Extension)
	}
	ThumbnailOptions.Width = 0
	ThumbnailOptions.Height = 0
	ThumbnailOptions.Extension = ""
	defaultsForMe()
	if ThumbnailOptions.Width != int(ConfigData.Thumbnails.Width) {
		t.Fatalf("Width default failed %d\n", ThumbnailOptions.Width)
	}
	if ThumbnailOptions.Height != int(ConfigData.Thumbnails.Height) {
		t.Fatalf("Height default failed %d\n", ThumbnailOptions.Height)
	}
	if ThumbnailOptions.Extension != ConfigData.Thumbnails.Extension {
		t.Fatalf("Extension default failed %s\n", ThumbnailOptions.Extension)
	}
	ThumbnailOptions.Width = 50
	ThumbnailOptions.Height = 60
	ThumbnailOptions.Extension = ".gif"
	defaultsForMe()
	if ThumbnailOptions.Width != int(50) {
		t.Fatalf("Width default failed %d\n", ThumbnailOptions.Width)
	}
	if ThumbnailOptions.Height != int(60) {
		t.Fatalf("Height default failed %d\n", ThumbnailOptions.Height)
	}
	if ThumbnailOptions.Extension != ".gif" {
		t.Fatalf("Extension default failed %s\n", ThumbnailOptions.Extension)
	}

	ThumbnailOptions.Type = "bob"
	err := defaultsForMe()
	if err == nil {
		t.Fatalf("Didn't flinch on a bob type\n")
	}
}

func TestMakeThumbnailFor(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	// for _, r := range []string{"bob.jpg", "bob.gif", "bob.jpeg", "bob.png"} {
	// 	if !canMakeThumbnailFor(r) {
	// 		t.Fatalf("Should have made a thumbnail for %s", r)
	// 	}
	// }
	for _, r := range []string{"bob-thumb.jpg", "mike-and-friends.txt", "SteampunkStridesProfessor-233x300-thumb-thumb.jpg"} {
		if canMakeThumbnailFor(r) {
			t.Fatalf("Shouldn't make a thumbnail for %s", r)
		}
	}
}

func TestGetThumbnailFilename(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	name := getThumbnailFilename("/usr/local/zend/apache2/mep.jpg")
	expect := "/usr/local/zend/apache2/mep-thumb.jpg"

	if name != expect {
		t.Fatalf("Bad filename result %s|%s\n", name, expect)
	}
	name = getThumbnailFilename("/usr/local/zend/apache2/mep.gif")
	expect = "/usr/local/zend/apache2/mep-thumb.jpg"

	if name != expect {
		t.Fatalf("Bad filename result %s|%s\n", name, expect)
	}
}

func TestMakeThumbnail(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/thumbnail/`)
	ThumbnailOptions.Width = 100
	ThumbnailOptions.Height = 100
	ThumbnailOptions.Type = "jpeg"
	ThumbnailOptions.Regenerate = true
	err := makeThumbnail(testdataloader.GetBasePath() + `/../features/tests/thumbnail/LogVisualiser.png`)
	if err != nil {
		t.Fatalf("Ded %v\n", err)
	}

	err = makeThumbnail(testdataloader.GetBasePath() + `/../features/tests/rss/rss1_out.xml`)
	if err == nil {
		t.Fatalf("Created a thumbnail where I shoudln't")
	}

	ThumbnailOptions.Type = "gif"
	ThumbnailOptions.Extension = "-thumb.gif"
	err = makeThumbnail(filepath.Join(ConfigData.BaseDir, `LogVisualiser.png`))
	if err != nil {
		t.Fatalf("Ded %v\n", err)
	}
	ThumbnailOptions.Type = "png"
	ThumbnailOptions.Extension = "-thumb.png"
	err = makeThumbnail(filepath.Join(ConfigData.BaseDir, `LogVisualiser.png`))
	if err != nil {
		t.Fatalf("Ded %v\n", err)
	}
	ThumbnailOptions.Type = "txt"
	ThumbnailOptions.Extension = "-thumb.txt"
	err = makeThumbnail(filepath.Join(ConfigData.BaseDir, `LogVisualiser.png`))
	if err == nil {
		t.Fatalf("Created a thumbnail where I shoudln't")
	}
}

func TestLetsGoThumbnailSingle(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/thumbnail/`)
	ThumbnailOptions.Width = 100
	ThumbnailOptions.Height = 100
	ThumbnailOptions.Type = "jpeg"
	ThumbnailOptions.Regenerate = false
	os.Remove(testdataloader.GetBasePath() + `/../features/tests/thumbnail/LogVisualiser-thumb.jpg`)
	ThumbnailOptions.Filename = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/thumbnail/LogVisualiser.png`)
	err := letsGoThumbnail()
	if err != nil {
		t.Fatalf("Ded %v\n", err)
	}

}

func TestDetectImage(t *testing.T) {
	contents, err := os.Open(testdataloader.GetBasePath() + `/../features/tests/thumbnail/LogVisualiser.png`)
	if err != nil {
		t.Fatalf("Could not process file %s\n%v\n", testdataloader.GetBasePath()+`/../features/tests/thumbnail/LogVisualiser.png`, err)
	}
	defer contents.Close()
	body := make([]byte, 512)
	_, err = contents.Read(body)
	if err != nil {
		t.Fatalf("Cound not read file %s\n", testdataloader.GetBasePath()+`/../features/tests/thumbnail/LogVisualiser.png`)
	}
	fileType := http.DetectContentType(body)
	if fileType != "image/png" {
		t.Fatalf("Detect failed for %s\n", fileType)
	}
}

func TestWebp2(t *testing.T) {
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/thumbnail/`)
	i, e := readImage(filepath.Join(ConfigData.BaseDir, `code-of-the-coder-cover.webp`))
	if e != nil {
		t.Fatalf("Failed to run %s\n", e)
	}
	if i == nil {
		t.Fatalf("i failed\n")
	}
	writeImage(i, filepath.Join(ConfigData.BaseDir, `code-of-the-coder-cover2.webp`))
}

func TestRecursiveThumbnail(t *testing.T) {
	ConfigData.BaseDir = filepath.Clean(testdataloader.GetBasePath() + `/../features/tests/thumbnail/`)
	ThumbnailOptions.Extension = "-thumb.jpg"
	ThumbnailOptions.Width = 100
	ThumbnailOptions.Height = 100
	ThumbnailOptions.Type = "jpeg"
	ThumbnailOptions.Regenerate = true
	ThumbnailOptions.Filename = ``
	os.Remove(filepath.Join(ConfigData.BaseDir, `media/code-of-the-coder-cover-thumb.jpg`))
	os.Remove(filepath.Join(ConfigData.BaseDir, `media/LogVisualiser-thumb.jpg`))
	os.Remove(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser-thumb.jpg`))
	os.Remove(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser copy-thumb.jpg`))
	os.Remove(filepath.Join(ConfigData.BaseDir, `media/x/y/LogVisualiser-thumb.jpg`))
	letsGoThumbnail()
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`x\LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser copy-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`x\LogVisualiser copy-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/y/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`y\LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/code-of-the-coder-cover-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`code-of-the-coder-cover-thumb-thumb.jpg`)
	}
	letsGoThumbnail()
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`x\LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser copy-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`x\LogVisualiser copy-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/y/LogVisualiser-thumb.jpg`)) != nil {
		t.Fatalf("File %s does not exist\n", ConfigData.BaseDir+`y\LogVisualiser-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/LogVisualiser-thumb-thumb.jpg`)) == nil {
		t.Fatalf("File %s exist\n", ConfigData.BaseDir+`LogVisualiser-thumb-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser-thumb-thumb.jpg`)) == nil {
		t.Fatalf("File %s exist\n", ConfigData.BaseDir+`x\LogVisualiser-thumb-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/LogVisualiser copy-thumb-thumb.jpg`)) == nil {
		t.Fatalf("File %s exist\n", ConfigData.BaseDir+`x\LogVisualiser copy-thumb-thumb.jpg`)
	}
	if fileExists(filepath.Join(ConfigData.BaseDir, `media/x/y/LogVisualiser-thumb-thumb.jpg`)) == nil {
		t.Fatalf("File %s exist\n", ConfigData.BaseDir+`y\LogVisualiser-thumb-thumb.jpg`)
	}
}

func fileExists(path string) error {
	_, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("file not exists error")
	}
	return nil
}
